package formatters

import (
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"

	"github.com/awalterschulze/gographviz"

	"github.com/Natolumin/ircmap/irctree"
)

// The Escape API for gographviz doesn't escape names containing a dot. dot doesn't accept those
func esc(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

func usersToWeight(users int) string {
	uf := (math.Sqrt((float64)(users + 1))) / 10
	return strconv.FormatFloat(uf, 'f', -1, 64)
}

func BuildDot(ircmap []irctree.Server, displayAll bool) *gographviz.Graph {
	graph := gographviz.NewGraph()
	for _, node := range ircmap {
		if !displayAll && node.Position == irctree.PositionUnknown {
			continue
		}
		attrs := make(map[string]string)
		attrs["fixedsize"] = "shape"
		attrs["width"] = usersToWeight(node.Users)
		attrs["height"] = usersToWeight(node.Users)
		attrs["tooltip"] = esc(node.Description)
		attrs["label"] = esc(node.Label)
		if node.Position == irctree.PositionHub {
			attrs["shape"] = "diamond"
			attrs["width"] = "1"
			attrs["height"] = "1"
		}
		graph.AddNode("", esc(node.ServerName), attrs)
	}
	for _, node := range ircmap {
		if node.ParentName != "" && (displayAll || node.Position != irctree.PositionUnknown) {
			attrs := make(map[string]string)
			attrs["len"] = strconv.FormatFloat(math.Log10((float64)(node.Lag+1)), 'f', -1, 64)
			attrs["tooltip"] = strconv.Itoa(node.Lag)
			graph.AddEdge(esc(node.ServerName), esc(node.ParentName), false, attrs)
		}
	}
	// The weight of a hub is added to the weight of its children
	return graph
}

//BuildPNG returns a PNG from a gographviz graph
func BuildPNG(g *gographviz.Graph, dot *exec.Cmd) ([]byte, error) {
	stdin, err := dot.StdinPipe()
	if err != nil {
		return nil, err
	}
	io.WriteString(stdin, g.String())
	stdin.Close()
	return dot.Output()
}
