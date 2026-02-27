package main

import (
	"fmt"
	"sport_app/auth"
	"sport_app/handlers"
	simpleconnection "sport_app/models/simple_connection"
	simplesql "sport_app/models/simple_sql"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	simpleconnection.CheckConnection()
	defer simpleconnection.Close()

	if err := simplesql.CreateTableUsers(simpleconnection.Ctx, simpleconnection.Conn); err != nil {
		panic(err)
	}

	if err := simplesql.CreateTableProfile(simpleconnection.Ctx, simpleconnection.Conn); err != nil {
		panic(err)
	}
	fmt.Println("Успешно!")
	jwt_token, err := auth.CreateToken("13412341321", false)
	if err != nil {
		panic(err)
	}
	token, _, err := new(jwt.Parser).ParseUnverified(jwt_token, jwt.MapClaims{})
	if err != nil {
		fmt.Println("Ошибка парсинга:", err)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		fmt.Println("Поля токена:")
		for key, val := range claims {
			fmt.Printf("%s: %v\n", key, val)
		}
	}
	err = handlers.StartHTTPServer()
	if err != nil {
		panic(err)
	}
}
