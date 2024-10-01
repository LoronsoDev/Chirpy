package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/LoronsoDev/chirpy/internal/auth"
	"github.com/LoronsoDev/chirpy/internal/database"
	"github.com/google/uuid"
)

func (apiCfg apiConfig) handlerUpdateCredentials(w http.ResponseWriter, r *http.Request) {
	type incomingParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	incParams := incomingParams{}
	err := decoder.Decode(&incParams)

	token, err := auth.GetBearerToken(r.Header)

	defer r.Body.Close()

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Provided JWT token can't be retrieved correctly")
		return
	}
	userID, err := auth.ValidateJWT(token, apiCfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	hashedPassword, err := auth.HashPassword(incParams.Password)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	credParams := database.ChangeUserCredentialsParams{
		ID:             userID,
		Email:          incParams.Email,
		HashedPassword: hashedPassword,
	}

	userResource, err := apiCfg.db.ChangeUserCredentials(r.Context(), credParams)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	retUserNoPassw := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		Email     string    `json:"email"`
		ChirpyRed bool      `json:"is_chirpy_red"`
	}{
		ID:        userResource.ID,
		CreatedAt: userResource.CreatedAt,
		UpdatedAt: userResource.UpdatedAt,
		Email:     userResource.Email,
		ChirpyRed: userResource.ChirpyRed,
	}
	respondWithJSON(w, http.StatusOK, retUserNoPassw)
}
