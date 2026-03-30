package main

import (
	"bytes"
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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

				reader, err := cli.ContainerLogs(context.Background(), c.ID, container.LogsOptions{
					ShowStdout: true,
					ShowStderr: true,
					Tail:       "1000",
				})
				if err != nil {
					return 0
				}

				var stdout, stderr bytes.Buffer

				// 🔥 TRÈS IMPORTANT
				_, err = stdcopy.StdCopy(&stdout, &stderr, reader)
				if err != nil {
					return 0
				}

				logContent := stdout.String()

				joins := strings.Count(logContent, " joins.")
				leaves := strings.Count(logContent, " disconnected")

				players := joins - leaves

				if players < 0 {
					return 0
				}

				return players
			}
		}
	}

	return 0
}