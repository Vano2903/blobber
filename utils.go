package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func connectToDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:33061)/blobber?parseTime=true&charset=utf8") //
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

func checkJWT(w http.ResponseWriter, r *http.Request) CustomClaims {
	jwt, err := r.Cookie("JWT")
	if err != nil {
		returnError(w, http.StatusUnauthorized, "No JWT cookie found")
		return CustomClaims{}
	}

	jwtContent, err := ParseToken(jwt.Value)
	if err != nil {
		returnError(w, http.StatusUnauthorized, "Invalid JWT")
		return CustomClaims{}
	}
	return jwtContent
}
