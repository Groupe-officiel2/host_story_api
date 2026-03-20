// players.go

package main

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func GetPlayersForServer(containerName string) int {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return 0
	}

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return 0
	}

	for _, c := range containers {
		for _, name := range c.Names {

			if strings.Contains(name, containerName) {

				logs, err := cli.ContainerLogs(context.Background(), c.ID, container.LogsOptions{
					ShowStdout: true,
					Tail:       "100",
				})
				if err != nil {
					return 0
				}

				buf := new(strings.Builder)
				_, _ = buf.ReadFrom(logs)

				logContent := buf.String()

				return strings.Count(logContent, "joined")
			}
		}
	}

	return 0
}