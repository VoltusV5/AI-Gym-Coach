package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sport_app/auth"
	"sport_app/mlclient"
	simpleconnection "sport_app/models/simple_connection"
	simplesql "sport_app/models/simple_sql"
	"time"
)

type Result struct {
	plan mlclient.Plan
	err  error
}

func GuestHandler(w http.ResponseWriter, r *http.Request) {
	user_id, err := simplesql.InsertRowsUsers(
		simpleconnection.Ctx,
		simpleconnection.Conn,
		true,
		"free",
	)

	if err != nil {
		panic(err)
	}

	jwt_token, err := auth.CreateToken(user_id, true)
	fmt.Println(jwt_token)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", jwt_token))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged is successfully"))
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
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

	if err := simplesql.UpdateProfile(r.Context(), simpleconnection.Conn, userID, requestData); err != nil {
		http.Error(w, "Error profile update: "+err.Error(), http.StatusInternalServerError)
		return
	}

	profile, err := simplesql.GetProfile(r.Context(), simpleconnection.Conn, userID)
	if err != nil {
		http.Error(w, "The profile has been updated, but the current data could not be loaded.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

func ResponceGenerateHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "Failed to get user id", http.StatusInternalServerError)
		return
	}

	profile, err := simplesql.GetProfile(r.Context(), simpleconnection.Conn, userID)
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
		ch <- Result{*plan, err}
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

	exercises_plan_weight, exercises_plan := simplesql.GetExercises(r.Context(), simpleconnection.Conn, res.plan, userID)
	err2 := simplesql.InsertRowsPrograms(userID, true, res.plan, exercises_plan, r.Context(), simpleconnection.Conn)
	if err2 != nil {
		http.Error(w, "Error inserting plan into database", http.StatusInternalServerError)
		return
	}

	workingWeightsMap := make(map[int]*int)
	for _, day := range exercises_plan_weight.Plan {
		for _, exerciseList := range day.Exercises {
			for _, ex := range exerciseList {
				workingWeightsMap[ex.ID] = ex.Weight
			}
		}
	}

	workingWeightsJSON, err3 := json.Marshal(workingWeightsMap)
	if err3 != nil {
		http.Error(w, "Failed to marshal working weights", http.StatusInternalServerError)
		return
	}

	err4 := simplesql.InsertRowsData(r.Context(), simpleconnection.Conn, userID, workingWeightsJSON)
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
