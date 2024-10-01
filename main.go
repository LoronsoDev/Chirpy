package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/LoronsoDev/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits int
	db             *database.Queries
	jwtSecret      string
	polkaKey       string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	// by default, godotenv will look for a file named .env in the current directory
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	dbURL := os.Getenv("DB_URL")

	serveMux := http.NewServeMux()
	// handler := http.FileServer(http.Dir(filepathRoot))

	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	if *dbg {
		os.Remove("database.json")
	}

	db, _ := sql.Open("postgres", dbURL)

	dbQueries := database.New(db)

	cfg := apiConfig{
		db:             dbQueries,
		jwtSecret:      jwtSecret,
		fileserverHits: 0,
		polkaKey:       os.Getenv("POLKA_KEY"),
	}

	// serveMux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(handler)))
	serveMux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	serveMux.HandleFunc("POST /api/users", cfg.handlerNewUser)
	serveMux.HandleFunc("PUT /api/users", cfg.handlerUpdateCredentials)

	serveMux.HandleFunc("POST /api/login", cfg.handlerLogin)

	serveMux.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	serveMux.HandleFunc("POST /api/revoke", cfg.handlerRevoke)

	serveMux.HandleFunc("GET /api/chirps", cfg.handlerGetAllChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetSpecificChirp)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handlerDeleteChirp)
	serveMux.HandleFunc("POST /api/chirps", cfg.handlerNewChirp)

	// Webhooks...
	serveMux.HandleFunc("POST /api/polka/webhooks", cfg.handlerPolkaWebhook)

	// serveMux.HandleFunc("GET /api/healthz", handlerHealth)
	// serveMux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
