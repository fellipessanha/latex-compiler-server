// @title          LaTeX Compiler API
// @version        1.0
// @description    Compiles LaTeX sources received from a git repository or an uploaded archive.
// @host           localhost:8080
// @schemes        http
// @BasePath       /
//
// @securityDefinitions.apikey BearerAuth
// @in             header
// @name           Authorization
package main

import (
	"log"
	"net/http"
	"os"

	_ "latex-compiler-api/docs"
	"latex-compiler-api/internal/auth"
	"latex-compiler-api/internal/handlers"
	"latex-compiler-api/internal/logging"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	authenticator, err := auth.New()
	if err != nil {
		log.Fatalf("auth setup: %v", err)
	}

	mux := http.NewServeMux()

	// Health check — no auth required.
	mux.HandleFunc("GET /health", healthHandler)

	// Swagger UI — no auth required.
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	// Authenticated routes. Both bare and /{output} variants are registered
	// because Go 1.22's mux does not match /git against /git/{output}.
	mux.Handle("POST /git", logging.WithLog(auth.WithAuth(authenticator, http.HandlerFunc(handlers.Git))))
	mux.Handle("POST /git/{output}", logging.WithLog(auth.WithAuth(authenticator, http.HandlerFunc(handlers.Git))))
	mux.Handle("POST /blob", logging.WithLog(auth.WithAuth(authenticator, http.HandlerFunc(handlers.Blob))))
	mux.Handle("POST /blob/{output}", logging.WithLog(auth.WithAuth(authenticator, http.HandlerFunc(handlers.Blob))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("latex-compiler-api listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// healthHandler returns 200 OK to indicate the service is alive.
//
// @Summary  Health check
// @Tags     system
// @Success  200
// @Router   /health [get]
func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
