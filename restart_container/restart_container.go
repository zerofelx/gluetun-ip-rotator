package restart_container

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type IPResponse struct {
	PublicIP string `json:"public_ip"`
}

// Make a GET request to the Gluetun container API to get the Public IP
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

// RestartContainer restarts the specified container and waits for a new IP.
// Returns newIP, oldIP, error.
func RestartContainer(containerName string, ip string, port string) (string, string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", "", fmt.Errorf("error creating Docker client: %v", err)
	}
	defer cli.Close()

	ctx := context.Background()

	gluetunAPIURL := "http://" + ip + ":" + port

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
		return "", oldIP, fmt.Errorf("error restarting Gluetun container: %v", err)
	}

	fmt.Println("Container restarted.")
	fmt.Println("Waiting for Gluetun to start (10 seconds)...")
	time.Sleep(10 * time.Second)

	newIP, err := waitForConnection(gluetunAPIURL, 60*time.Second)
	if err != nil {
		return "", oldIP, fmt.Errorf("error waiting for Gluetun to start: %v", err)
	}

	fmt.Println("Reconnected to Gluetun.")
	fmt.Printf("Old IP: %s\nNew IP: %s\n", oldIP, newIP)

	return newIP, oldIP, nil
}
