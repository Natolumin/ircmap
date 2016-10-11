package main

import (
	"encoding/json"
)

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Lag    int    `json:"lag"`
}

type Graph struct {
	Servers []Server
	Links   []Link
}

func BuildJson(ircmap []Server) []byte {

	var links []Link
	for _, server := range ircmap {
		if server.ParentName != "" {
			links = append(links, Link{
				Source: server.ParentName,
				Target: server.ServerName,
				Lag:    server.Lag,
			})
		}
	}

	res, _ := json.Marshal(Graph{Servers: ircmap, Links: links})
	return res
}
