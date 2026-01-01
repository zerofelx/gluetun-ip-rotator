package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"docker_manager/check_container"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	ip := os.Getenv("SERVER_IP")
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "6968"
	}

	r := mux.NewRouter()
	r.HandleFunc("/health", HandleHealth).Methods("GET")
	r.HandleFunc("/restart", HandleRestart).Methods("POST")

	addr := fmt.Sprintf("%s:%s", ip, port)
	log.Printf("Server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	// We can hardcode "gluetun" for now as per previous logic, or make it dynamic later
	info, err := check_container.GetContainerInfo("gluetun")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(info))
}

func HandleRestart(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Restart endpoint placeholder"))
}
