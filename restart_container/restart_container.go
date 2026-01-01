package restartcontainer

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

func RestartContainer(containerName string) (string, error) {
	// Crear cliente Docker
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("error while creating Docker client: %v", err)
	}
	defer cli.Close()

	// Crear contexto
	ctx := context.Background()

	// Inspeccionar contenedor
	containerJSON, err := cli.ContainerInspect(ctx, containerName)
	if err != nil {
		return "", fmt.Errorf("error while inspecting container '%s': %v", containerName, err)
	}

	// Construir información del contenedor
	var output string
	output += "=================================================\n"
	output += fmt.Sprintf("Container '%s' information:\n", containerName)
	output += "=================================================\n"
	output += fmt.Sprintf("ID: %s\n", containerJSON.ID[:12])
	output += fmt.Sprintf("Status: %s\n", containerJSON.State.Status)
	output += fmt.Sprintf("Running: %t\n", containerJSON.State.Running)
	output += fmt.Sprintf("Image: %s\n", containerJSON.Config.Image)
	output += "=================================================\n"

	if containerJSON.State.Running {
		output += fmt.Sprintf("Initializing container '%s'\n", containerJSON.State.StartedAt)
		output += fmt.Sprintf("PID: %d\n", containerJSON.State.Pid)
	} else {
		output += fmt.Sprintf("Finishing container '%s'\n", containerJSON.State.FinishedAt)
		output += fmt.Sprintf("Exit code: %d\n", containerJSON.State.ExitCode)
	}

	// Mostrar puertos
	if len(containerJSON.NetworkSettings.Ports) > 0 {
		output += "\nPorts:\n"
		for port, bindings := range containerJSON.NetworkSettings.Ports {
			if len(bindings) > 0 {
				for _, binding := range bindings {
					output += fmt.Sprintf(" %s -> %s:%s\n", port, binding.HostIP, binding.HostPort)
				}
			}
		}
	}

	output += "=================================================\n"
	return output, nil
}
