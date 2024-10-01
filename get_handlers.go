package main

import (
	"net/http"

	"github.com/LoronsoDev/chirpy/internal/database"
	"github.com/google/uuid"
)

func (apiConf *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	sortDir := r.URL.Query().Get("sort")

	var dbChirps []database.Chirp
	var err error

	if authorID != "" {
		uniqueId, _ := uuid.Parse(authorID)
		dbChirps, err = apiConf.db.GetChirpsFromUser(r.Context(), uniqueId)
	} else {
		if sortDir == "desc" {
			dbChirps, err = apiConf.db.GetAllChirpsDescOrder(r.Context())
		} else {
			dbChirps, err = apiConf.db.GetAllChirpsAscOrder(r.Context())
		}
	}
	if err != nil {
		respondWithError(w, http.StatusFailedDependency, err.Error())
	}
	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (apiConf *apiConfig) handlerGetSpecificChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")
	id, err := uuid.Parse(chirpIDStr)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	dbChirp, err := apiConf.db.GetChirp(r.Context(), id)

	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	chirp := Chirp{}
	chirp.ID = dbChirp.ID
	chirp.CreatedAt = dbChirp.CreatedAt
	chirp.UpdatedAt = dbChirp.UpdatedAt
	chirp.Body = dbChirp.Body
	chirp.UserID = dbChirp.UserID

	respondWithJSON(w, http.StatusOK, chirp)
}
