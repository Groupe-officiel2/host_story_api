package main

type Server struct {
	ID      int    `json:"ID"`
	Name    string `json:"Name"`
	Players int    `json:"Players"`
	Slots   int    `json:"Slots"`
}

type ServerDTO struct {
	ID      int    `json:"ID"`
	Name    string `json:"Name"`
	Players int    `json:"Players"`
	Slots   int    `json:"Slots"`
}

var servers []Server