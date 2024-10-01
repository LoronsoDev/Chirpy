package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func getCleanBody(body string) (cleanedBody string) {
	//Separate it into words
	words := strings.Split(body, " ")
	badWords := [3]string{"kerfuffle", "sharbert", "fornax"}

	for i, word := range words {
		word = strings.ToLower(word)
		for _, badWord := range badWords {
			if word == badWord {
				words[i] = "****"
			}
		}
	}

	return strings.Join(words, " ")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorParam struct {
		Error string `json:"error"`
	}

	errParam := errorParam{
		Error: msg,
	}
	res, _ := json.Marshal(errParam)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(res)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	res, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(res)
}
