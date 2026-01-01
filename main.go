package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"docker_manager/check_container"
	"docker_manager/restart_container"
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

type RestartRequest struct {
	ContainerName string `json:"container_name"`
	GluetunPort   string `json:"gluetun_port"`
}

type RestartResponse struct {
	Message string `json:"message"`
	OldIP   string `json:"old_ip"`
	NewIP   string `json:"new_ip"`
}

func HandleRestart(w http.ResponseWriter, r *http.Request) {
	var req RestartRequest
	// Try to decode, if fails we just assume default container
	json.NewDecoder(r.Body).Decode(&req)

	containerName := req.ContainerName
	if containerName == "" {
		http.Error(w, "container_name is required", http.StatusBadRequest)
		return
	}

	gluetunPort := req.GluetunPort
	if gluetunPort == "" {
		http.Error(w, "gluetun_port is required", http.StatusBadRequest)
		return
	}

	// We need the server IP for the gluetun API check
	serverIP := os.Getenv("SERVER_IP")
	if serverIP == "" {
		http.Error(w, "SERVER_IP environment variable is not set", http.StatusInternalServerError)
		return
	}

	newIP, oldIP, err := restart_container.RestartContainer(containerName, serverIP, gluetunPort)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := RestartResponse{
		Message: "Container restarted successfully",
		OldIP:   oldIP,
		NewIP:   newIP,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
