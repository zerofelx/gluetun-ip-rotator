package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
)

type IPResponse struct {
	PublicIP string `json:"public_ip"`
}

// Hace una petición GET a la API del container de Gluetun para conocer la IP Pública
func getGluetunIP(apiURL string) (string, error) {
	resp, err := http.Get(apiURL + "/v1/publicip/ip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result IPResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.PublicIP, nil
}

func waitForConnection(apiURL string, maxWait time.Duration) (string, error) {
	fmt.Println("Waiting for connection to Gluetun...")

	timeout := time.After(maxWait)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timeout waiting for connection after %v", maxWait)
		case <-ticker.C:
			ip, err := getGluetunIP(apiURL)
			if err == nil && ip != "" {
				return ip, nil
			}
			fmt.Print("Waiting for connection to Gluetun...")
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	ip := os.Getenv("SERVER_IP")
	containerName := "gluetun-1"
	gluetunAPIURL := "http://" + ip + ":6969"

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	defer cli.Close()

	ctx := context.Background()

	fmt.Printf("Container: %s\n", containerName)

	oldIP, err := getGluetunIP(gluetunAPIURL)
	if err != nil {
		fmt.Printf("Error getting Gluetun IP: %v\n", err)
		oldIP = "unknown"
	} else {
		fmt.Printf("Gluetun IP: %s\n", oldIP)
	}

	fmt.Println("Restarting Gluetun container...")
	timeout := 30 // Seconds
	err = cli.ContainerRestart(ctx, containerName, container.StopOptions{Timeout: &timeout})
	if err != nil {
		log.Fatalf("Error restarting Gluetun container: %v", err)
	}

	fmt.Println("Container restarted.")
	fmt.Println("Waiting for Gluetun to start (10 seconds)...")
	time.Sleep(10 * time.Second)

	newIP, err := waitForConnection(gluetunAPIURL, 60*time.Second)
	if err != nil {
		log.Fatalf("\nError waiting for Gluetun to start: %v", err)
	}

	fmt.Println("Reconnected to Gluetun.")
	fmt.Printf("Old IP: %s\nNew IP: %s\n", oldIP, newIP)

	if oldIP != "unknown" && oldIP == newIP {
		fmt.Println("IP has not changed.")
	} else if oldIP != "unknown" {
		fmt.Println("IP has changed.")
	} else {
		fmt.Println("IP is unknown.")
	}
}
