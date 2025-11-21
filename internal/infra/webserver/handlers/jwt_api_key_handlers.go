package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/HalexV/pos-go-expert-desafio-rate-limiter/internal/usecase"
	"github.com/go-chi/jwtauth"
)

type Error struct {
	Message string `json:"message"`
}

type JWTAPIKeyHandler struct{}

func NewJWTAPIKeyHandler() *JWTAPIKeyHandler {
	return &JWTAPIKeyHandler{}
}

func (h *JWTAPIKeyHandler) CreateJWTAPIKey(w http.ResponseWriter, r *http.Request) {
	var payload usecase.CreateJWTAPIKeyInputDTO
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	createJWTAPIKey := usecase.NewCreateJWTAPIKeyUseCase()

	apiTokenConfig, err := createJWTAPIKey.Execute(payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err := Error{Message: err.Error()}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	jwt := r.Context().Value("jwt").(*jwtauth.JWTAuth)
	jwtExpiresIn := r.Context().Value("jwtExpiresIn").(int)

	_, tokenString, _ := jwt.Encode(map[string]interface{}{
		"sub":            apiTokenConfig.ID.String(),
		"exp":            time.Now().Add(time.Duration(jwtExpiresIn) * time.Second).Unix(),
		"maxReqsBySec":   apiTokenConfig.MaxReqsBySec,
		"blockTimeBySec": apiTokenConfig.BlockTimeBySec,
	})

	accessToken :=
		struct {
			AccessToken string `json:"access_token"`
		}{
			AccessToken: tokenString,
		}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accessToken)
}
