package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/awalterschulze/gographviz"
)

var (
	serverDomain = ".rezosup.org"
	hubPrefix    = "hub."
	leafPrefix   = "irc."
)

const (
	PositionUnknown = iota
	PositionHub
	PositionLeaf
)

type Stats struct {
	XMLName    xml.Name `xml:"inspircdstats"`
	ServerList []Server `xml:"serverlist>server"`
}

type Server struct {
	XMLName     xml.Name `xml:"server" json:"-"`
	ServerName  string   `xml:"servername" json:"id"`
	ParentName  string   `xml:"parentname" json:"-"`
	Label       string   `xml:"-" json:"label"`
	Lag         int      `xml:"lagmillisecs" json:"lagmillisecs"`
	Users       int      `xml:"usercount" json:"usercount"`
	Transit     int      `xml:"-" json:"-"`
	Description string   `xml:"gecos" json:"desc"`
	Position    int      `xml:"-" json:"group"`
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
	tree := buildTree(ircmap.ServerList)
	tree.FlattenLag()
	tree.BuildTransit()
	switch {
	case *json:
		fmt.Print(string(BuildJson(tree)))
	case *dot:
		fmt.Print(BuildDot(tree.Slice()))
	default:
		fmt.Print(tree.String())
	}
}

// The Escape API for gographviz doesn't escape names containing a dot. dot doesn't accept those
func esc(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

func usersToWeight(users int) string {
	uf := (math.Sqrt((float64)(users + 1))) / 10
	return strconv.FormatFloat(uf, 'f', -1, 64)
}

func scrubValues(ircmap []Server) {
	for i := range ircmap {
		node := &ircmap[i]
		node.Label = strings.TrimSuffix(node.ServerName, serverDomain)
		if strings.HasPrefix(node.Label, hubPrefix) {
			node.Label = strings.TrimPrefix(node.Label, hubPrefix)
			node.Position = PositionHub
		} else if strings.HasPrefix(node.Label, leafPrefix) {
			node.Position = PositionLeaf
			node.Label = strings.TrimPrefix(node.Label, leafPrefix)
		}
	}
}

func BuildDot(ircmap []Server) *gographviz.Graph {

	graph := gographviz.NewGraph()
	for _, node := range ircmap {
		if node.Position == PositionUnknown {
			continue
		}
		attrs := make(gographviz.Attrs)
		attrs["fixedsize"] = "shape"
		attrs["width"] = usersToWeight(node.Users)
		attrs["height"] = usersToWeight(node.Users)
		attrs["tooltip"] = esc(node.Description)
		attrs["label"] = esc(node.Label)
		if node.Position == PositionHub {
			attrs["shape"] = "diamond"
			attrs["width"] = "1"
			attrs["height"] = "1"
		}
		graph.AddNode("", esc(node.ServerName), attrs)
	}
	for _, node := range ircmap {
		if node.ParentName != "" && node.Position != PositionUnknown {
			attrs := make(gographviz.Attrs)
			attrs["len"] = strconv.FormatFloat(math.Log10((float64)(node.Lag+1)), 'f', -1, 64)
			attrs["tooltip"] = strconv.Itoa(node.Lag)
			graph.AddEdge(esc(node.ServerName), esc(node.ParentName), false, attrs)
		}
	}
	// The weight of a hub is added to the weight of its children
	return graph
}
