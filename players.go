// players.go

package main

import "math/rand"

func GetPlayersForServer(name string) int {
    return rand.Intn(20)
}