package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/jwtauth"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/evgfitil/gophermart.git/internal/models"
)

const (
	tokenExpireDuration = time.Hour * 3
)

var tokenAuth *jwtauth.JWTAuth

type UserStorage interface {
	CreateUser(ctx context.Context, username string, passwordHash string) error
	GetUserByUsername(ctx context.Context, username string) (string, error)
	GetUserID(ctx context.Context, username string) (int, error)
	IsUserUnique(ctx context.Context, username string) (bool, error)
}

func init() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "jwtDefaultSecret"
	}

	tokenAuth = jwtauth.New("HS256", []byte(jwtSecret), nil)
}

func generateToken(username string) (string, error) {
	expirationTime := time.Now().Add(tokenExpireDuration)
	_, tokenString, err := tokenAuth.Encode(jwt.MapClaims{
		"user_id": username,
		"exp":     expirationTime.Unix(),
	})

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func HandleUserLogin(us UserStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		var user models.User
		if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
			http.Error(res, "invalid request body", http.StatusBadRequest)
			return
		}

		if user.Password == "" {
			http.Error(res, "password is required", http.StatusBadRequest)
			return
		}

		storedUserPassword, err := us.GetUserByUsername(requestContext, user.Username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(res, "user not found", http.StatusUnauthorized)
			} else {
				http.Error(res, "database error", http.StatusInternalServerError)
			}
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

func HandleUserRegistration(us UserStorage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		var user models.User
		if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
			http.Error(res, "invalid request body", http.StatusBadRequest)
			return
		}
		if user.Password == "" {
			http.Error(res, "password is required", http.StatusBadRequest)
			return
		}

		isUnique, err := us.IsUserUnique(requestContext, user.Username)
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

		err = us.CreateUser(requestContext, user.Username, string(hashedPassword))
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
