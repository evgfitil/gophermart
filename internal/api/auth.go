package api

import (
	"context"
	"encoding/json"
	"github.com/evgfitil/gophermart.git/internal/database"
	"github.com/evgfitil/gophermart.git/internal/logger"
	"github.com/evgfitil/gophermart.git/internal/models"
	"github.com/go-chi/jwtauth"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"time"
)

const (
	tokenExpireDuration = time.Hour * 3
)

var tokenAuth *jwtauth.JWTAuth

func init() {
	tokenAuth = jwtauth.New("HS256", []byte(os.Getenv("JWT_SECRET")), nil)
}

func generateToken(username string) (string, error) {
	expirationTime := time.Now().Add(tokenExpireDuration)
	_, tokenString, err := tokenAuth.Encode(jwt.MapClaims{
		"user_id": username,
		"exp":     expirationTime.Unix(),
	})

	if err != nil {
		logger.Sugar.Infof("Something went wrong here: %v", err.Error())
		return "", err
	}

	return tokenString, nil
}

func AuthHandler(db database.DBStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		var user models.User
		if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
			http.Error(res, "invalid request body", http.StatusBadRequest)
			return
		}

		storedUserPassword, err := db.GetUserByUsername(requestContext, user.Username)
		if err != nil {
			http.Error(res, "user not found", http.StatusUnauthorized)
			return
		}

		if err = bcrypt.CompareHashAndPassword([]byte(storedUserPassword), []byte(user.Password)); err != nil {
			http.Error(res, "wrong username or password", http.StatusUnauthorized)
			return
		}

		tokenString, err := generateToken(user.Username)
		if err != nil {
			http.Error(res, "failed to generate auth token", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Authorization", "Bearer "+tokenString)
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("logged in successfully"))
	}
}

func RegisterHandler(db database.DBStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		var user models.User
		if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
			http.Error(res, "invalid request body", http.StatusBadRequest)
			return
		}

		isUnique, err := db.IsUserUnique(requestContext, user.Username)
		if err != nil {
			http.Error(res, "database error", http.StatusInternalServerError)
			return
		}
		if !isUnique {
			http.Error(res, "user already exists", http.StatusConflict)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(res, "error hashing password", http.StatusInternalServerError)
			return
		}

		err = db.CreateUser(requestContext, user.Username, string(hashedPassword))
		if err != nil {
			http.Error(res, "error creating user", http.StatusInternalServerError)
			return
		}

		tokenString, err := generateToken(user.Username)
		if err != nil {
			http.Error(res, "failed to generate auth token", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Authorization", "Bearer "+tokenString)
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("User registered successfully"))
	}
}
