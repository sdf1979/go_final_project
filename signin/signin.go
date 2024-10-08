package signin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtKey = []byte("super_secret_key")

type Credentials struct {
	Password string `json:"password"`
}

type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.StandardClaims
}

func SignInHandler(writer http.ResponseWriter, request *http.Request) {
	var creds Credentials
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil {
		responseWithJson(writer, http.StatusBadRequest, formatErrorForFrontend(err.Error()))
		return
	}

	TODO_PASSWORD := os.Getenv("TODO_PASSWORD")
	if TODO_PASSWORD == "" || creds.Password != TODO_PASSWORD {
		responseWithJson(writer, http.StatusUnauthorized, formatErrorForFrontend("Неверный пароль"))
		return
	}

	expirationTime := time.Now().Add(8 * time.Hour)
	claims := &Claims{
		PasswordHash: TODO_PASSWORD,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	fmt.Println("token:", tokenString)
	if err != nil {
		responseWithJson(writer, http.StatusInternalServerError, formatErrorForFrontend("Не удалось создать токен"))
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(map[string]string{
		"token": tokenString,
	})
}

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// смотрим наличие пароля
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			cookie, err := request.Cookie("token")
			if err != nil {
				http.Error(writer, "Authentification required", http.StatusUnauthorized)
				return
			}
			tokenStr := cookie.Value

			claims := &jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil // jwtKey должен быть определен
			})

			if err != nil || !token.Valid {
				http.Error(writer, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next(writer, request)
	})
}

func formatErrorForFrontend(errorStr string) map[string]interface{} {
	return map[string]interface{}{
		"error": errorStr,
	}
}

func responseWithJson(writer http.ResponseWriter, httpCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	writer.WriteHeader(httpCode)
	writer.Write(response)
}
