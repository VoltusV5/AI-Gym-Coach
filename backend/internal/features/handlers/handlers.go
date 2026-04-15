package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	simplesql "sport_app/internal/core/models/simple_sql"
	middleware "sport_app/internal/core/transport/http/middleware"
	"sport_app/internal/features/mlclient"
	"time"

	auth "sport_app/internal/core/auth"
)

type Result struct {
	plan mlclient.Plan
	err  error
}

func GuestHandler(w http.ResponseWriter, r *http.Request) {
	user_id, err := simplesql.InsertRowsUsers(
		r.Context(),
		dbpool.Pool,
		true,
		"free",
	)

	if err != nil {
		log.Printf("Guest InsertRowsUsers: %v", err)
		http.Error(w, "Failed to create guest user", http.StatusInternalServerError)
		return
	}

	jwt_token, err := auth.CreateToken(user_id, true)
	if err != nil {
		log.Printf("Guest CreateToken: %v", err)
		http.Error(w, "Failed to issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", jwt_token))
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"token": jwt_token})
}

func ProfileGetHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Failed to get user id", http.StatusInternalServerError)
		return
	}

	profile, err := simplesql.GetProfile(r.Context(), dbpool.Pool, userID)
	if err != nil {
		http.Error(w, "Failed to load profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(profile)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Failed to get user id", http.StatusInternalServerError)
		return
	}

	var requestData map[string]any
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Error reading HTTP request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	if len(requestData) == 0 {
		http.Error(w, "Data for profile update not transferred", http.StatusBadRequest)
		return
	}

	if err := simplesql.UpdateProfile(r.Context(), dbpool.Pool, userID, requestData); err != nil {
		http.Error(w, "Error profile update: "+err.Error(), http.StatusInternalServerError)
		return
	}

	profile, err := simplesql.GetProfile(r.Context(), dbpool.Pool, userID)
	if err != nil {
		http.Error(w, "The profile has been updated, but the current data could not be loaded.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

func ResponceGenerateHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Failed to get user id", http.StatusInternalServerError)
		return
	}

	profile, err := simplesql.GetProfile(r.Context(), dbpool.Pool, userID)
	if err != nil {
		http.Error(w, "The profile has been updated, but the current data could not be loaded.", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*52)
	defer cancel()

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
		http.Error(w, "Request timeout or cancelled", http.StatusGatewayTimeout)
		log.Printf("ML request failed after %v: %v", duration, ctx.Err())
		return
	case res = <-ch:
		if res.err != nil {
			http.Error(w, "Failed to generate plan: "+res.err.Error(), http.StatusInternalServerError)
			log.Printf("ML request failed after %v: %v", duration, res.err)
			return
		}
	}

	exercises_plan_weight, exercises_plan := simplesql.GetExercises(r.Context(), dbpool.Pool, res.plan, userID)
	err2 := simplesql.InsertRowsPrograms(userID, true, res.plan, exercises_plan, r.Context(), dbpool.Pool)
	if err2 != nil {
		http.Error(w, "Error inserting plan into database", http.StatusInternalServerError)
		return
	}

	existingWeights, errW := simplesql.GetUserWorkingWeightsMap(r.Context(), dbpool.Pool, userID)
	if errW != nil {
		http.Error(w, "Failed to load working weights: "+errW.Error(), http.StatusInternalServerError)
		return
	}
	simplesql.ApplyExistingWeightsToPlan(&exercises_plan_weight, existingWeights)

	workingWeightsJSON, err3 := simplesql.MergeWorkingWeightsJSON(existingWeights, exercises_plan_weight)
	if err3 != nil {
		http.Error(w, "Failed to merge working weights", http.StatusInternalServerError)
		return
	}

	err4 := simplesql.InsertRowsData(r.Context(), dbpool.Pool, userID, workingWeightsJSON)
	if err4 != nil {
		http.Error(w, "Failed to save working weights", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(exercises_plan_weight); err != nil {
		log.Printf("Error writing response: %v: %v", duration, err)
	}

	log.Printf("ML request succeeded after %v", duration)
}

func WorkoutCompleteHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Failed to get user id", http.StatusInternalServerError)
		return
	}

	var req simplesql.WorkoutCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error reading HTTP request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := simplesql.CompleteWorkout(r.Context(), dbpool.Pool, userID, req); err != nil {
		log.Printf("WorkoutCompleteHandler error: %v", err)
		http.Error(w, "Error completing workout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"saved_id": fmt.Sprintf("real-%d", time.Now().Unix()),
	})
}
