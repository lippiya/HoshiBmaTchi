package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	router := mux.NewRouter()

	swaggerDir := "./swagger"
	if _, err := os.Stat(swaggerDir); os.IsNotExist(err) {
		swaggerDir = "."
	}

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	))

	router.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, swaggerDir+"/swagger.json")
	}).Methods("GET")

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")


	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
