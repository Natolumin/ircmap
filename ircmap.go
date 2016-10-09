package main

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/awalterschulze/gographviz"
)

const (
	graphName  = "IRC Map"
	undirected = false
)

type Stats struct {
	XMLName    xml.Name `xml:"inspircdstats"`
	ServerList []Server `xml:"serverlist>server"`
}

type Server struct {
	XMLName     xml.Name `xml:"server"`
	ServerName  string   `xml:"servername"`
	ParentName  string   `xml:"parentname"`
	Lag         int      `xml:"lagmillisecs"`
	Users       int      `xml:"usercount"`
	Description string   `xml:"gecos"`
}

func main() {

	//FIXME: pluggable file source
	statfile, err := os.Open("stats.xml")
	//FIXME: error handling
	if err != nil {
		panic(err)
	}
	dec := xml.NewDecoder(statfile)
	var ircmap Stats
	err = dec.Decode(&ircmap)
	//FIXME: error handling
	if err != nil {
		panic(err)
	}
	graph := BuildDot(ircmap.ServerList)
	fmt.Print(graph)
}

// The Escape API for gographviz doesn't escape names containing a dot. dot doesn't accept those
func esc(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

func usersToWeight(users int) string {
	uf := (math.Log10((float64)(users + 1)))
	return strconv.FormatFloat(uf, 'f', -1, 64)
}
func BuildDot(ircmap []Server) *gographviz.Graph {

	graph := gographviz.NewGraph()
	for _, node := range ircmap {
		attrs := make(gographviz.Attrs)
		attrs["fixedsize"] = "shape"
		attrs["width"] = usersToWeight(node.Users)
		attrs["height"] = usersToWeight(node.Users)
		attrs["tooltip"] = esc(node.Description)
		graph.AddNode("", esc(node.ServerName), attrs)
	}
	for _, node := range ircmap {
		if node.ParentName != "" {
			attrs := make(gographviz.Attrs)
			//attrs["weight"] = strconv.FormatFloat(1/math.Log10((float64)(node.Lag+1)), 'f', -1, 64)
			attrs["len"] = strconv.FormatFloat(math.Log10((float64)(node.Lag+1)), 'f', -1, 64)
			attrs["tooltip"] = strconv.Itoa(node.Lag)
			graph.AddEdge(esc(node.ServerName), esc(node.ParentName), undirected, attrs)
		}
	}
	// The weight of a hub is added to the weight of its children
	return graph
}
