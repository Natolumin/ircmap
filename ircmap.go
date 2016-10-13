package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Natolumin/ircmap/formatters"
	"github.com/Natolumin/ircmap/irctree"
)

var (
	serverDomain = ".rezosup.org"
	hubPrefix    = "hub."
	leafPrefix   = "irc."
	displayAll   = false
)

func init() {
	configValues := []struct {
		Env string
		Var *string
	}{
		{Env: "IRCMAP_SERVER_DOMAIN", Var: &serverDomain},
		{Env: "IRCMAP_HUB_PREFIX", Var: &hubPrefix},
		{Env: "IRCMAP_LEAF_PREFIX", Var: &leafPrefix},
	}
	for _, val := range configValues {
		if env := os.Getenv(val.Env); env != "" {
			*val.Var = env
		}
	}
}

type Stats struct {
	XMLName    xml.Name         `xml:"inspircdstats"`
	ServerList []irctree.Server `xml:"serverlist>server"`
}

func main() {

	var output = flag.String("o", "raw", "Output format (dot, json, raw)")
	flag.BoolVar(&displayAll, "all", false, "Don't scrub unrecognized nodes")
	flag.Parse()

	dec := xml.NewDecoder(os.Stdin)
	var ircmap Stats
	err := dec.Decode(&ircmap)
	//FIXME: error handling
	if err != nil {
		log.Fatalf("Error building the map: %s", err)
	}
	scrubValues(ircmap.ServerList)
	tree := irctree.BuildTree(ircmap.ServerList)
	switch *output {
	case "json":
		fmt.Print(formatters.BuildJson(tree, displayAll))
	case "dot":
		fmt.Print(formatters.BuildDot(tree.Slice(), displayAll))
	default:
		fmt.Print(tree.String())
	}
}
func scrubValues(ircmap []irctree.Server) {
	for i := range ircmap {
		node := &ircmap[i]
		node.Label = strings.TrimSuffix(node.ServerName, serverDomain)
		if strings.HasPrefix(node.Label, hubPrefix) {
			node.Label = strings.TrimPrefix(node.Label, hubPrefix)
			node.Position = irctree.PositionHub
		} else if strings.HasPrefix(node.Label, leafPrefix) {
			node.Position = irctree.PositionLeaf
			node.Label = strings.TrimPrefix(node.Label, leafPrefix)
		}
	}
}
