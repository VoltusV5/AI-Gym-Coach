package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sport_app/auth"
	simpleconnection "sport_app/models/simple_connection"
	simplesql "sport_app/models/simple_sql"
)

type Plan struct {
	Plan_template map[string]Days   `json:"plan_template"`
	Weights       map[string]string `json:"weights"`
}

type Days struct {
	Groupe    []string `json:"groupe"`
	Exercises []string `json:"exercises"`
	Sets      []int    `json:"sets"`
	Reps      string   `json:"reps"`
	Rest      string   `json:"rest"`
}

func JsonInStructHandler(w http.ResponseWriter, r *http.Request) {
	var plan Plan
	if err := json.NewDecoder(r.Body).Decode(&plan); err != nil {
		fmt.Println("Не удалось прочитать тело запроса:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	for k, v := range plan.Plan_template {
		fmt.Println("День", k, ":")
		fmt.Println("----------------")
		v.PrintStruct()
		fmt.Println()
	}
}

func (d Days) PrintStruct() {
	fmt.Println("Группа мышц:", d.Groupe)
	fmt.Println("Упражнения:", d.Exercises)
	fmt.Println("Кол-во подходов:", d.Sets)
	fmt.Println("Кол-во повторений", d.Reps)
	fmt.Println("Отдых:", d.Rest)
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
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jwt_token))
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	log.Printf("userID from context: %q, ok: %v", userID, ok)
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

	delete(requestData, "token")

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
