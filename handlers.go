// handlers.go

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
)

// CreateTemplateContainer handles the creation of a new container from a template
func CreateTemplateContainer(w http.ResponseWriter, r *http.Request) {
	serverMutex.Lock()
	defer serverMutex.Unlock()

	ownerID := UserIDFromContext(r.Context())
	if ownerID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get container name from query parameter
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Container name is required", http.StatusBadRequest)
		return
	}

	availablePort, err := findAvailablePort()
	if err != nil {
		http.Error(w, fmt.Sprintf("no available ports: %v", err), http.StatusInternalServerError)
		return
	}
	hostPort := strconv.Itoa(availablePort)

	image := r.URL.Query().Get("image")
	if image == "" {
		image = "server-vintagestory:latest"
	}

	// Get player slots from query parameter
	playerSlotsParam := r.URL.Query().Get("players")
	playerSlots := 2
	if playerSlotsParam != "" {
		if parsedSlots, err := strconv.Atoi(playerSlotsParam); err == nil && parsedSlots > 0 {
			playerSlots = parsedSlots
		}
	}

	const (
		defaultMemoryLimit = int64(1395864371) // 1.30 GB in bytes (base 1024)
		memoryPerPlayer    = int64(314572800)  // 300 MB in bytes (base 1024)
	)

	// Calculate memory limit based on player slots
	totalMemory := defaultMemoryLimit + (int64(playerSlots-1) * memoryPerPlayer)

	containerID, err := CreateContainer(r.Context(), image, name, hostPort, totalMemory, ownerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = StartContainer(r.Context(), containerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to start container: %v", err), http.StatusInternalServerError)
		return
	}

    go func() {
        jsonData := fmt.Sprintf(`{
            "id": "%s",
            "name": "%s",
            "slots": %d
        }`, containerID, name, playerSlots)

        req, err := http.NewRequest("POST", "http://host.docker.internal:8000/api/servers", strings.NewReader(jsonData))
        if err != nil {
            fmt.Println("Laravel request error:", err)
            return
        }

        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-KEY", "SECRET123")

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            fmt.Println("Laravel API error:", err)
            return
        }
        defer resp.Body.Close()

        fmt.Println("Server saved in Laravel:", resp.Status)
    }()

	SVRPort, err := strconv.Atoi(hostPort)
	if err == nil {
		AddSRVRecord(SVRPort, name)
	}

	fmt.Fprintf(w, "Container launched: %s with name %s on host port %s and %d player slots\n", containerID, name, hostPort, playerSlots)
}

func ToggleHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")

	if name == "" {
		http.Error(w, "server name is required", http.StatusBadRequest)
		return
	}

	actorRole := UserRoleFromContext(r.Context())
	if !strings.EqualFold(actorRole, "admin") {
		http.Error(w, "forbidden: admin role required", http.StatusForbidden)
		return
	}

	result, err := ToggleContainer(context.Background(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Container %s %s\n", name, result)
}

func GetPlayers(w http.ResponseWriter, r *http.Request) {

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "server name required", http.StatusBadRequest)
		return
	}

	players := GetPlayersForServer(name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{
		"players": players,
	})
}

func GetServers(w http.ResponseWriter, r *http.Request) {
    cli, err := getDockerClient()
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    var liveServers []map[string]interface{}

    for _, c := range containers {
        if c.Labels["app"] != "vintagestory" {
            continue
        }

        name := c.Labels["name"]
        
        slots := 1
        if val, ok := c.Labels["slots"]; ok {
            if p, err := strconv.Atoi(val); err == nil {
                slots = p
            }
        } else if val, ok := c.Labels["players"]; ok { // Fallback if old code created them
            if p, err := strconv.Atoi(val); err == nil {
                slots = p
            }
        }

        liveServers = append(liveServers, map[string]interface{}{
            "ID":      c.ID[:12],
            "Name":    name,
            "Players": GetPlayersForServer(name),
            "Slots":   slots,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(liveServers)
    fmt.Println("Servers count:", len(liveServers))
}