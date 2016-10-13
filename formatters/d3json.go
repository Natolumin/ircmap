package formatters

import (
	"encoding/json"

	"github.com/Natolumin/ircmap/irctree"
)

type Link struct {
	Source  string `json:"source"`
	Target  string `json:"target"`
	Lag     int    `json:"lag"`
	Transit int    `json:"transit"`
}

type Graph struct {
	Servers []irctree.Server
	Links   []Link
}

func BuildJson(ircmap *irctree.Servers, displayAll bool) string {

	res := Graph{}
	for _, server := range ircmap.Lookup {
		if displayAll || server.Position != irctree.PositionUnknown {
			res.Servers = append(res.Servers, server.Server)
		}
	}
	for _, server := range ircmap.Lookup {
		if server.Parent == nil {
			continue
		}
		if displayAll || (server.Position != irctree.PositionUnknown && server.Parent.Position != irctree.PositionUnknown) {
			res.Links = append(res.Links, Link{
				Source:  server.ParentName,
				Target:  server.ServerName,
				Lag:     server.Parent.Lag,
				Transit: server.Parent.Transit,
			})
		}
	}

	jsonobj, _ := json.Marshal(res)
	return string(jsonobj)
}
