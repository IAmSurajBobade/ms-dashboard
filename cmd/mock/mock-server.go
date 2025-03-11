package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	services := strings.Split(os.Getenv("SERVICE_NAMES"), ",")
	versions := strings.Split(os.Getenv("VERSIONS"), ",")
	// envs := strings.Split(os.Getenv("ENVIRONMENTS"), ",")

	// router := http.NewServeMux()

	for _, version := range versions {
		for _, service := range services {
			// for _, env := range envs {
			http.HandleFunc("/base/"+service+version+"/health", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"status": "up",
				})
			})
			// }
		}
	}

	port := os.Getenv("REST_PORT")
	if port == "" {
		port = "9001"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
