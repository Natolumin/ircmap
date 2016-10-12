package main

import (
	"encoding/json"
)

type Link struct {
	Source  string `json:"source"`
	Target  string `json:"target"`
	Lag     int    `json:"lag"`
	Transit int    `json:"transit"`
}

type Graph struct {
	Servers []Server
	Links   []Link
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func BuildJson(ircmap *Servers) []byte {

	res := Graph{}
	for _, server := range ircmap.Lookup {
		if displayAll || server.Node.Position != PositionUnknown {
			res.Servers = append(res.Servers, server.Node)
		}
	}
	for _, server := range ircmap.Lookup {
		if server.Parent != nil && (displayAll || (server.Node.Position != PositionUnknown && server.Parent.Node.Position != PositionUnknown)) {
			currentLink := Link{
				Source: server.Node.ParentName,
				Target: server.Node.ServerName,
				Lag:    server.Node.Lag,
			}
			switch {
			case server.Node.Position == PositionLeaf:
				currentLink.Transit = server.Node.Transit
			case server.Parent.Node.Position == PositionLeaf:
				currentLink.Transit = server.Parent.Node.Transit
			case server.Node.Position == PositionHub && server.Parent.Node.Position == PositionHub:
				currentLink.Transit = max(server.Node.Transit, server.Parent.Node.Transit)
			}
			res.Links = append(res.Links, currentLink)
		}
	}

	jsonobj, _ := json.Marshal(res)
	return jsonobj
}
