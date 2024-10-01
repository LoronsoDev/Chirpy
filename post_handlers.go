package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/LoronsoDev/chirpy/internal/auth"
	"github.com/LoronsoDev/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

// func handlerHealth(res http.ResponseWriter, req *http.Request) {
// 	res.Header().Add("Content-Type", "text/plain; charset=utf-8")
// 	res.WriteHeader(http.StatusOK)
// 	res.Write([]byte(http.StatusText(http.StatusOK)))
// }

// func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
// 	html := fmt.Sprintf(`
// 	<html>

// 	<body>
// 		<h1>Welcome, Chirpy Admin</h1>
// 		<p>Chirpy has been visited %d times!</p>
// 	</body>

// 	</html>`,
// 		cfg.fileserverHits)

// 	w.Header().Add("Content-Type", "text/html; charset=utf-8")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(fmt.Sprintf(html)))
// }

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	currentPlatform := os.Getenv("PLATFORM")
	if currentPlatform != "dev" {
		respondWithError(w, http.StatusForbidden, "This is a very dangerous operation, can only be accessed in a local dev environment")
		return
	}
	cfg.db.RemoveAllUsers(r.Context())
	respondWithJSON(w, http.StatusOK, struct {
		Message string `json:"msg"`
	}{
		Message: "All users where removed",
	})
}

func (apiCfg apiConfig) handlerNewChirp(w http.ResponseWriter, r *http.Request) {
	const maxChars = 140
	type incomingParams struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(r.Body)
	incParams := incomingParams{}
	err := decoder.Decode(&incParams)

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Provided JWT token can't be retrieved correctly")
		return
	}

	userID, err := auth.ValidateJWT(token, apiCfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	defer r.Body.Close()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	if len(incParams.Body) > maxChars {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	addChirpParams := database.AddChirpParams{}
	addChirpParams.Body = getCleanBody(incParams.Body)
	addChirpParams.UserID = userID

	newChirp, err := apiCfg.db.AddChirp(r.Context(), addChirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp, err: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body:      newChirp.Body,
		UserID:    newChirp.UserID,
	})
}

func (apiCfg *apiConfig) handlerNewUser(w http.ResponseWriter, r *http.Request) {
	type incomingParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	incParams := incomingParams{}
	err := decoder.Decode(&incParams)

	defer r.Body.Close()

	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	hashed_password, err := auth.HashPassword(incParams.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	createUserParams := database.CreateUserParams{
		Email:          incParams.Email,
		HashedPassword: hashed_password,
	}

	newUser, err := apiCfg.db.CreateUser(r.Context(), createUserParams)

	if err != nil {
		respondWithError(w, 400, err.Error())
	}

	respondWithJSON(w, http.StatusCreated, struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		ChirpyRed bool      `json:"is_chirpy_red"`
	}{
		ID:        newUser.ID,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
		Email:     newUser.Email,
		ChirpyRed: newUser.ChirpyRed,
	})
}

func (apiCfg apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tokenData, err := apiCfg.db.GetTokenInfo(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if tokenData.RevokedAt.Valid || tokenData.ExpiresAt.Time.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Token is revoked or has expired")
		return
	}
	token, _ := auth.MakeRefreshToken()
	rtParams := database.AddRefreshTokenParams{Token: token, UserID: tokenData.UserID}
	apiCfg.db.AddRefreshToken(r.Context(), rtParams)

	accessToken, err := auth.MakeJWT(tokenData.UserID, apiCfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	respondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{Token: accessToken})
}

func (apiCfg apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	err = apiCfg.db.RevokeToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	respondWithJSON(w, http.StatusNoContent, struct{}{})
}

func (apiCfg apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type incomingParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	incParams := incomingParams{}
	err := decoder.Decode(&incParams)

	defer r.Body.Close()

	userStoredData, err := apiCfg.db.GetUserByEmail(r.Context(), incParams.Email)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = auth.CheckPasswordHash(incParams.Password, userStoredData.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	//At this point, user is the same and password has been guessed...

	token, err := auth.MakeJWT(userStoredData.ID, apiCfg.jwtSecret, time.Duration(time.Hour))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}

	rtParams := database.AddRefreshTokenParams{
		Token:  refreshToken,
		UserID: userStoredData.ID,
	}
	apiCfg.db.AddRefreshToken(r.Context(), rtParams)

	respondWithJSON(w, http.StatusOK, struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		ChirpyRed    bool      `json:"is_chirpy_red"`
	}{
		ID:           userStoredData.ID,
		CreatedAt:    userStoredData.CreatedAt,
		UpdatedAt:    userStoredData.UpdatedAt,
		Email:        userStoredData.Email,
		Token:        token,
		RefreshToken: refreshToken,
		ChirpyRed:    userStoredData.ChirpyRed,
	})
}

func (apiCfg apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type incomingParams struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(r.Body)
	incParams := incomingParams{}
	err := decoder.Decode(&incParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	apiKey, err := auth.GetAPIKey(r.Header)
	validKey := apiCfg.polkaKey == apiKey
	if err != nil || !validKey {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}
	defer r.Body.Close()

	if incParams.Event == "user.upgraded" {
		err = apiCfg.db.UpgradeUser(r.Context(), incParams.Data.UserID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithJSON(w, http.StatusNoContent, struct{}{})
	}
	respondWithError(w, http.StatusNoContent, "")
}
