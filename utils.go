package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func connectToDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4", os.Getenv("DATABASE_USER"), os.Getenv("DATABASE_PASSWORD"), os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT"), os.Getenv("DATABASE_NAME")))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func returnError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": true}`, code, message)
}

func returnErrorJson(w http.ResponseWriter, code int, message, key string, json []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": true, "%s": %s}`, code, message, key, json)
}

func returnSuccess(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": false}`, code, message)
}

func returnSuccessJson(w http.ResponseWriter, code int, message, key string, json []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"code": %d, "msg":"%s", "error": false, "%s": %s}`, code, message, key, json)
}

func checkJWT(w http.ResponseWriter, r *http.Request) (CustomClaims, error) {
	jwt, err := r.Cookie("JWT")
	if err != nil {
		returnError(w, http.StatusUnauthorized, "No JWT cookie found")
		return CustomClaims{}, err
	}

	jwtContent, err := ParseToken(jwt.Value)
	if err != nil {
		returnError(w, http.StatusUnauthorized, "Invalid JWT")
		return CustomClaims{}, err
	}
	return jwtContent, nil
}
