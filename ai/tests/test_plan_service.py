"""Contract, constraint and real-model tests; all profiles and catalog IDs are synthetic."""

import copy
import os
import unittest
from unittest.mock import patch

from fastapi import FastAPI
from fastapi.testclient import TestClient
from pydantic import ValidationError

from plan_service import GeneratePlanRequest, GeneratePlanResponse, create_router, validate_plan


TOKEN = "synthetic-integration-token"
HEADERS = {"Authorization": f"Bearer {TOKEN}"}
GROUPS = ["Ноги", "Грудь", "Спина", "Плечи", "Руки", "Пресс"]
TEMPLATE = {"тип_сплита": "Фулбади", "еженедельный_план": [
    {"упражнения": [{"группа": group} for group in GROUPS]} for _ in range(3)
]}


def payload():
    return {
        "request_id": "req-test", "feature_schema_version": "profile-v2",
        "profile": {"age": 26, "height_cm": 175, "weight_kg": 75.5, "level": "beginner",
                    "goal": "muscle_gain", "training_days": [1, 3, 5], "equipment_ids": ["dumbbell"],
                    "excluded_exercise_ids": [], "session_duration_min": 60,
                    "gender": "м", "activity_level": "Средняя активность", "injuries_notes": False},
        "history": {"completed_28d": 0, "adherence_28d": None, "exercise_metrics": []},
        "exercises": [{"exercise_id": 1001 + i, "muscle_group": group,
                       "muscle_subgroup": None, "equipment_ids": ["dumbbell"]}
                      for i, group in enumerate(GROUPS)],
    }


def test_app(predict=lambda _: ("Фулбади", []), programs=None):
    app = FastAPI()
    app.include_router(create_router(predict, [TEMPLATE] if programs is None else programs))
    return TestClient(app)


class PlanContractTests(unittest.TestCase):
    def setUp(self):
        self.environment = patch.dict(os.environ, {"ML_SERVICE_TOKEN": TOKEN})
        self.environment.start()
        self.addCleanup(self.environment.stop)
        self.client = test_app()

    def post(self, data=None):
        return self.client.post("/v2/plans/generate", json=payload() if data is None else data, headers=HEADERS)

    def test_model_plan_preserves_ids_days_and_metadata(self):
        response = self.post()
        self.assertEqual(response.status_code, 200, response.text)
        result = GeneratePlanResponse.model_validate(response.json())
        self.assertEqual(result.request_id, "req-test")
        self.assertEqual(result.feature_schema_version, "profile-v2")
        self.assertEqual(result.model_version, "catboost-v0-planner-v2")
        self.assertFalse(result.fallback)
        self.assertEqual([day.weekday for day in result.days], [1, 3, 5])
        self.assertIn("INITIAL_WEIGHT_REQUIRES_CONFIRMATION", result.warnings)
        self.assertTrue(all(ex.weight_kg is None for day in result.days for ex in day.exercises))
        validate_plan(GeneratePlanRequest.model_validate(payload()), result)

    def test_private_authentication_and_missing_configuration(self):
        for headers in ({}, {"Authorization": "Bearer wrong"}, {"Authorization": "Basic abc"}):
            response = self.client.post("/v2/plans/generate", json=payload(), headers=headers)
            self.assertEqual(response.status_code, 401)
        with patch.dict(os.environ, {"ML_SERVICE_TOKEN": ""}):
            self.assertEqual(self.post().status_code, 503)

    def test_invalid_inputs_are_rejected(self):
        changes = [
            ("age", True), ("age", "26"), ("age", 0), ("weight_kg", "75"),
            ("training_days", []), ("training_days", [0, 8]), ("training_days", [1, 1]),
            ("equipment_ids", ["dumbbell", "dumbbell"]), ("excluded_exercise_ids", [-1]),
            ("level", "expert"), ("goal", "unknown"), ("session_duration_min", 0),
        ]
        for key, value in changes:
            with self.subTest(key=key, value=value):
                data = payload()
                data["profile"][key] = value
                self.assertEqual(self.post(data).status_code, 422)
        for key, value in [("feature_schema_version", "v999"), ("exercises", []), ("unexpected", "value")]:
            data = payload()
            data[key] = value
            self.assertEqual(self.post(data).status_code, 422)
        data = payload()
        data["exercises"].append(data["exercises"][0])
        self.assertEqual(self.post(data).status_code, 422)
        data = payload()
        data["history"]["adherence_28d"] = 1.1
        self.assertEqual(self.post(data).status_code, 422)

    def test_non_finite_numbers_are_rejected(self):
        for value in (float("nan"), float("inf"), -float("inf")):
            data = payload()
            data["profile"]["weight_kg"] = value
            with self.assertRaises(ValidationError):
                GeneratePlanRequest.model_validate(data)

    def test_explicit_fallback_without_legacy_features_or_compatible_template(self):
        data = payload()
        for key in ("gender", "activity_level", "injuries_notes"):
            data["profile"][key] = None
        result = self.post(data).json()
        self.assertTrue(result["fallback"])
        self.assertEqual(result["model_version"], "rules-v1")
        self.assertIn("BASELINE_FEATURES_UNAVAILABLE", result["warnings"])
        client = test_app(programs=[])
        response = client.post("/v2/plans/generate", json=payload(), headers=HEADERS)
        self.assertTrue(response.json()["fallback"])

    def test_inference_failure_returns_valid_labelled_fallback(self):
        def fail(_):
            raise RuntimeError("internal model path must remain private")
        response = test_app(predict=fail).post("/v2/plans/generate", json=payload(), headers=HEADERS)
        self.assertEqual(response.status_code, 200)
        self.assertIn("BASELINE_INFERENCE_UNAVAILABLE", response.json()["warnings"])
        self.assertNotIn("internal model path", response.text)

    def test_equipment_exclusions_and_impossible_duration(self):
        for change in ("equipment", "excluded", "duration"):
            with self.subTest(change=change):
                data = payload()
                if change == "equipment":
                    data["profile"]["equipment_ids"] = []
                elif change == "excluded":
                    data["profile"]["excluded_exercise_ids"] = [ex["exercise_id"] for ex in data["exercises"]]
                else:
                    data["profile"]["session_duration_min"] = 5
                response = self.post(data)
                self.assertEqual(response.status_code, 422)
                self.assertEqual(response.json()["detail"]["code"], "NO_FEASIBLE_PLAN")

    def test_alternative_exercise_uses_current_catalog_id_and_all_equipment(self):
        data = payload()
        blocked = data["exercises"][0]
        alternative = copy.deepcopy(blocked)
        alternative["exercise_id"] = 9090
        data["exercises"].append(alternative)
        data["profile"]["excluded_exercise_ids"] = [blocked["exercise_id"]]
        result = self.post(data).json()
        for day in result["days"]:
            ids = [ex["exercise_id"] for ex in day["exercises"]]
            self.assertIn(9090, ids)
            self.assertNotIn(blocked["exercise_id"], ids)
        alternative["equipment_ids"].append("bench")
        self.assertEqual(self.post(data).status_code, 422)

    def test_frequency_level_goal_matrix(self):
        for days in range(1, 8):
            for level in ("beginner", "intermediate", "advanced"):
                for goal in ("muscle_gain", "weight_loss", "Набрать мышцы"):
                    with self.subTest(days=days, level=level, goal=goal):
                        data = payload()
                        data["profile"].update(training_days=list(range(1, days + 1)), level=level, goal=goal)
                        response = self.post(data)
                        self.assertEqual(response.status_code, 200, response.text)
                        validate_plan(GeneratePlanRequest.model_validate(data),
                                      GeneratePlanResponse.model_validate(response.json()))

    def test_short_session_keeps_core_groups_and_trims_accessories(self):
        for programs in ([TEMPLATE], []):
            data = payload()
            data["profile"]["session_duration_min"] = 30
            response = test_app(programs=programs).post("/v2/plans/generate", json=data, headers=HEADERS)
            self.assertEqual(response.status_code, 200, response.text)
            result = GeneratePlanResponse.model_validate(response.json())
            validate_plan(GeneratePlanRequest.model_validate(data), result)
            for day in result.days:
                ids = {ex.exercise_id for ex in day.exercises}
                self.assertTrue({1001, 1002, 1003} <= ids)
                self.assertLess(len(day.exercises), len(GROUPS))

    def test_independent_output_validator_rejects_corrupted_plan(self):
        valid = self.post().json()
        mutations = [
            lambda r: r.update(request_id="other"),
            lambda r: r["days"][0].update(weekday=7),
            lambda r: r["days"][0]["exercises"][0].update(exercise_id=46),
            lambda r: r["days"][0]["exercises"][0].update(position=9),
            lambda r: r["days"][0]["exercises"][0].update(sets=100),
            lambda r: r["days"][0]["exercises"][0].update(reps_min=99),
        ]
        for mutate in mutations:
            result = copy.deepcopy(valid)
            mutate(result)
            with self.assertRaises(ValueError):
                validate_plan(GeneratePlanRequest.model_validate(payload()), GeneratePlanResponse.model_validate(result))


class ExistingModelTests(unittest.TestCase):
    def test_real_model_v1_and_v2(self):
        # Import the production entry point: this loads the unchanged dataset and CatBoost weights.
        from main import app
        with patch.dict(os.environ, {"ML_SERVICE_TOKEN": TOKEN}), TestClient(app) as client:
            self.assertEqual(client.get("/health").json(), {"ok": True})
            legacy = {"age": 26, "height_cm": 175, "weight_kg": 75, "gender": "м",
                      "activity_level": "Средняя активность", "injuries_notes": False,
                      "goal": "Набрать мышцы", "fitness_level": "новичок", "training_days_map": ["Пн", "Ср", "Пт"]}
            response = client.post("/plan/user", json=legacy)
            self.assertEqual(response.status_code, 200, response.text)
            self.assertEqual(len(response.json()["еженедельный_план"]), 3)
            response = client.post("/v2/plans/generate", json=payload(), headers=HEADERS)
            self.assertEqual(response.status_code, 200, response.text)
            result = GeneratePlanResponse.model_validate(response.json())
            self.assertFalse(result.fallback)
            self.assertEqual(result.model_version, "catboost-v0-planner-v2")
            validate_plan(GeneratePlanRequest.model_validate(payload()), result)


if __name__ == "__main__":
    unittest.main()
