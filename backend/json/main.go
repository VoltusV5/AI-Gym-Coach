package main

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func main() {
	http.HandleFunc("/json", JsonInStructHandler)
	if err := http.ListenAndServe(":9091", nil); err != nil {
		fmt.Println("Ошибка при работе сервера:", err)
	}
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
