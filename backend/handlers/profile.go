package handlers

import (
	"encoding/json"
	"fmt"
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
