package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	core_auth "sport_app/internal/core/auth"
	core_logger "sport_app/internal/core/logger"
	core_models_simpleconnection "sport_app/internal/core/models/simple_connection"
	simplesql "sport_app/internal/core/models/simple_sql"
	core_http_middleware "sport_app/internal/core/transport/http/middleware"
	core_http_responce "sport_app/internal/core/transport/http/responce"
	"sport_app/internal/features/mlclient"
	"time"

	"go.uber.org/zap"
)

type Result struct {
	plan mlclient.Plan
	err  error
}

func GuestHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	respHandler := core_http_responce.NewHTTPResponce(log, w)

	userID, err := simplesql.InsertRowsUsers(ctx, dbpool.Pool, true, "free")
	if err != nil {
		respHandler.ErrorResponse(err, "Failed to create guest user")
		return
	}

	jwtToken, err := core_auth.CreateToken(userID, true)
	if err != nil {
		respHandler.ErrorResponse(err, "Failed to issue token")
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	respHandler.JSONResponse(map[string]string{"token": jwtToken}, http.StatusOK)
}

func ProfileGetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	respHandler := core_http_responce.NewHTTPResponce(log, w)
	userID, ok := r.Context().Value(core_http_middleware.UserIDKey).(string)
	if !ok {
		err := fmt.Errorf("user id not found in context")
		respHandler.ErrorResponse(err, "Failed to get user id")
		return
	}

	profile, err := simplesql.GetProfile(r.Context(), dbpool.Pool, userID)
	if err != nil {
		respHandler.ErrorResponse(err, "Failed to load profile: "+err.Error())
		return
	}

	respHandler.JSONResponse(profile, http.StatusOK)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	respHandler := core_http_responce.NewHTTPResponce(log, w)
	userID, ok := r.Context().Value(core_http_middleware.UserIDKey).(string)
	if !ok {
		err := fmt.Errorf("user id not found in context")
		respHandler.ErrorResponse(err, "Failed to get user id")
		return
	}

	var requestData map[string]any
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		respHandler.ErrorResponse(err, "Error reading HTTP request body")
		return
	}

	defer r.Body.Close()

	if len(requestData) == 0 {
		respHandler.ErrorResponse(fmt.Errorf("empty profile data"), "Data for profile update not transferred")
		return
	}

	if err := simplesql.UpdateProfile(r.Context(), dbpool.Pool, userID, requestData); err != nil {
		respHandler.ErrorResponse(err, "Error profile update: "+err.Error())
		return
	}

	profile, err := simplesql.GetProfile(r.Context(), dbpool.Pool, userID)
	if err != nil {
		respHandler.ErrorResponse(err, "The profile has been updated, but the current data could not be loaded.")
		return
	}

	respHandler.JSONResponse(profile, http.StatusOK)
}

func ResponceGenerateHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	respHandler := core_http_responce.NewHTTPResponce(log, w)
	userID, ok := r.Context().Value(core_http_middleware.UserIDKey).(string)
	if !ok {
		err := fmt.Errorf("user id not found in context")
		respHandler.ErrorResponse(err, "Failed to get user id")
		return
	}

	profile, err := simplesql.GetProfile(r.Context(), dbpool.Pool, userID)
	if err != nil {
		respHandler.ErrorResponse(err, "The profile has been updated, but the current data could not be loaded.")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*52)
	defer cancel()

	log.Debug("Generating plan from ML", zap.String("user_id", userID))

	ch := make(chan Result, 1)

	start := time.Now()

	go func() {
		plan, err := mlclient.GeneratePlan(ctx, profile)
		if err != nil {
			ch <- Result{err: err}
			return
		}
		ch <- Result{plan: *plan, err: nil}
	}()

	duration := time.Since(start)

	var res Result
	select {
	case <-ctx.Done():
		log.Debug("ML request cancelled",
			zap.Duration("duration", duration),
			zap.Error(ctx.Err()),
		)
		respHandler.ErrorResponse(ctx.Err(), "Request timeout or cancelled")
		return
	case res = <-ch:
		if res.err != nil {
			log.Debug("ML request failed",
				zap.Duration("duration", duration),
				zap.Error(res.err),
			)
			respHandler.ErrorResponse(res.err, "Failed to generate plan")
			return
		}
	}

	exercises_plan_weight, exercises_plan := simplesql.GetExercises(r.Context(), dbpool.Pool, res.plan, userID)
	err = simplesql.InsertRowsPrograms(userID, true, res.plan, exercises_plan, r.Context(), dbpool.Pool)
	if err != nil {
		respHandler.ErrorResponse(err, "Error inserting plan into database")
		return
	}

	existingWeights, err := simplesql.GetUserWorkingWeightsMap(r.Context(), dbpool.Pool, userID)
	if err != nil {
		respHandler.ErrorResponse(err, "Failed to load working weights: "+err.Error())
		return
	}
	simplesql.ApplyExistingWeightsToPlan(&exercises_plan_weight, existingWeights)

	workingWeightsJSON, err := simplesql.MergeWorkingWeightsJSON(existingWeights, exercises_plan_weight)
	if err != nil {
		respHandler.ErrorResponse(err, "Failed to merge working weights")
		return
	}

	err = simplesql.InsertRowsData(r.Context(), dbpool.Pool, userID, workingWeightsJSON)
	if err != nil {
		respHandler.ErrorResponse(err, "Failed to save working weights")
		return
	}

	log.Debug("ML request succeeded", zap.Duration("duration", duration))
	respHandler.JSONResponse(exercises_plan_weight, http.StatusOK)
}

func WorkoutCompleteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	respHandler := core_http_responce.NewHTTPResponce(log, w)
	userID, ok := r.Context().Value(core_http_middleware.UserIDKey).(string)
	if !ok {
		err := fmt.Errorf("user id not found in context")
		respHandler.ErrorResponse(err, "Failed to get user id")
		return
	}

	var req simplesql.WorkoutCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error reading HTTP request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := simplesql.CompleteWorkout(r.Context(), dbpool.Pool, userID, req); err != nil {
		respHandler.ErrorResponse(err, "Error completing workout: "+err.Error())
		return
	}

	respHandler.JSONResponse(map[string]any{
		"ok":       true,
		"saved_id": fmt.Sprintf("real-%d", time.Now().Unix()),
	}, http.StatusOK)
}

var dbpool *core_models_simpleconnection.ConnectionPool

func InitDBPool(p *core_models_simpleconnection.ConnectionPool) {
	dbpool = p
}
