package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

    "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func main() {

	http.HandleFunc("/servers", func(w http.ResponseWriter, r *http.Request) {

		cli, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

        containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
        if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var servers []map[string]interface{}

		for _, c := range containers {

       		if c.Labels["app"] != "vintagestory" {
     			continue
        	}

        	name := c.Labels["name"]

        	slots := 0
        	if val, ok := c.Labels["slots"]; ok {
        		p, _ := strconv.Atoi(val)
        		players = p
        	}

        	servers = append(servers, map[string]interface{}{
        		"ID":      c.ID[:12],
        		"Name":    name,
        		"Players": 0,
        		"Slots":   players,
        	})
        }

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(servers)
	})

    http.HandleFunc("/template", func(w http.ResponseWriter, r *http.Request) {

    		name := r.URL.Query().Get("name")
    		playersStr := r.URL.Query().Get("players")
    		image := r.URL.Query().Get("image")

    		if name == "" || playersStr == "" || image == "" {
    			http.Error(w, "Missing parameters", 400)
    			return
    		}

    		players, _ := strconv.Atoi(playersStr)

    		cli, err := client.NewClientWithOpts(client.FromEnv)
    		if err != nil {
    			http.Error(w, err.Error(), 500)
    			return
    		}

    		resp, err := cli.ContainerCreate(
    			context.Background(),
    			&container.Config{
    				Image: image,
    				Labels: map[string]string{
    					"app":     "vintagestory",
    					"name":    name,
    					"players": strconv.Itoa(players),
    				},
    			},
    			nil, nil, nil, "",
    		)

    		if err != nil {
    			http.Error(w, err.Error(), 500)
    			return
    		}

    		err = cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{})
    		if err != nil {
    			http.Error(w, err.Error(), 500)
    			return
    		}

    		json.NewEncoder(w).Encode(map[string]interface{}{
    			"id":      resp.ID[:12],
    			"name":    name,
    			"players": players,
    		})
    	})

	fmt.Println("Server running on :8082")

	http.ListenAndServe("0.0.0.0:8082", nil)
}