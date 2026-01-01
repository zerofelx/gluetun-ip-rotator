# Docker Gluetun Manager

A simple Go application to manage and restart [Gluetun](https://github.com/qdm12/gluetun) VPN containers remotely. This tool is designed to facilitate the update of the Public IP address for external tools that rely on a VPN connection.

## 🚀 Purpose

The main goal is to allow external services or scripts to trigger a restart of a specific Gluetun container to rotate the IP address. It provides a simple HTTP API to restart the container and waits until the new IP is acquired and verified.

## 💡 Context (Why use this?)

This tool is particularly useful when using **ProtonVPN** with the **WireGuard** protocol.

While OpenVPN implementations often allow IP rotation via standard HTTP control signals, rotating IPs with WireGuard on ProtonVPN (and similar providers) is often most reliably achieved by simply restarting the container to force a fresh connection to a new endpoint (assuming your container is configured to select random servers). This tool automates that process safely.

## ⚠️ Security Notice

**This application DOES NOT implement any authentication mechanism.**

It is strongly recommended to use this tool **ONLY** within a secure, private network (e.g., [Tailscale](https://tailscale.com/), WireGuard, or a local private network). Exposing this service directly to the public internet is insecure and not recommended.

Authentication can be added (e.g., Basic Auth or Token-based) if required for your specific use case.

## 🛠 Usage

### Prerequisites
- Docker
- Go (if building from source)
- A running Gluetun container with the control port open (default internal port 8000).

### API Endpoint

**POST** `/restart`

**Body:**
```json
{
  "container_name": "gluetun",
  "gluetun_port": "8000"
}
```

- `container_name`: The name of the Docker container to restart.
- `gluetun_port`: The host port mapped to the Gluetun control port (8000/tcp).

### Example

Restart `gluetun` which has its control port mapped to `8000` on the host:

```bash
curl -X POST http://localhost:6968/restart \
     -H "Content-Type: application/json" \
     -d '{"container_name": "gluetun", "gluetun_port": "8000"}'
```

**Response:**

```json
{
  "message": "Container restarted successfully",
  "old_ip": "203.0.113.45",
  "new_ip": "198.51.100.23"
}
```

## ⚙️ Configuration

The application is configured via environment variables (or a `.env` file):

- `SERVER_IP`: The IP address of the server (required).
- `SERVER_PORT`: The port for this API manager (default: 6968).

## 🐋 Docker

You can run this manager alongside your other containers. Ensure it has access to the Docker socket (`/var/run/docker.sock`) to perform container restarts.
