package main

import (
	"net/http"

	"github.com/LoronsoDev/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (apiCfg apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")
	chirpId, err := uuid.Parse(chirpIDStr)

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
	retChirp, err := apiCfg.db.GetChirp(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	if retChirp.UserID == userID {
		err := apiCfg.db.RemoveChirp(r.Context(), chirpId)
		if err != nil {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithJSON(w, http.StatusNoContent, struct{}{})
		return
	}
	respondWithError(w, http.StatusForbidden, "you are not the original poster of this chirp")
	return
}
