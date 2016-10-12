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
		if displayAll || server.Node.Position != irctree.PositionUnknown {
			res.Servers = append(res.Servers, server.Node)
		}
	}
	for _, server := range ircmap.Lookup {
		if server.Parent == nil {
			continue
		}
		if displayAll || (server.Node.Position != irctree.PositionUnknown && server.Parent.End.Node.Position != irctree.PositionUnknown) {
			res.Links = append(res.Links, Link{
				Source:  server.Node.ParentName,
				Target:  server.Node.ServerName,
				Lag:     server.Parent.Lag,
				Transit: server.Parent.Transit,
			})
		}
	}

	jsonobj, _ := json.Marshal(res)
	return string(jsonobj)
}
