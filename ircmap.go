package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Natolumin/ircmap/formatters"
	"github.com/Natolumin/ircmap/irctree"
)

var (
	serverDomain = ".rezosup.org"
	hubPrefix    = "hub."
	leafPrefix   = "irc."
)

type Stats struct {
	XMLName    xml.Name         `xml:"inspircdstats"`
	ServerList []irctree.Server `xml:"serverlist>server"`
}

var displayAll bool

func main() {

	var json = flag.Bool("json", false, "Output JSON")
	var dot = flag.Bool("dot", false, "Output a dot file")
	flag.BoolVar(&displayAll, "all", false, "Don't scrub unrecognized nodes")
	flag.StringVar(&serverDomain, "domain", ".rezosup.org", "Domain suffix to remove from server names")
	flag.StringVar(&hubPrefix, "hubprefix", "hub.", "Hostname prefix to identify hubs")
	flag.StringVar(&leafPrefix, "leafprefix", "irc.", "Hostname prefix to identify leaves")
	flag.Parse()

	dec := xml.NewDecoder(os.Stdin)
	var ircmap Stats
	err := dec.Decode(&ircmap)
	//FIXME: error handling
	if err != nil {
		panic(err)
	}
	scrubValues(ircmap.ServerList)
	tree := irctree.BuildTree(ircmap.ServerList)
	switch {
	case *json:
		fmt.Print(formatters.BuildJson(tree, displayAll))
	case *dot:
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
