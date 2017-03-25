package formatters

import (
	"testing"

	"github.com/Natolumin/ircmap/irctree"
)

// Our example tree is :
//      A
//    4/ \8
//    B   C__
//      2/ \1\1
//      D   E F

var exampleServers = []irctree.Server{
	{ServerName: "A", ParentName: "", Lag: 0, Users: 10, Position: irctree.PositionHub},
	{ServerName: "B", ParentName: "A", Lag: 4, Users: 5, Position: irctree.PositionLeaf},
	{ServerName: "C", ParentName: "A", Lag: 8, Users: 1, Position: irctree.PositionHub},
	{ServerName: "D", ParentName: "C", Lag: 10, Users: 20, Position: irctree.PositionLeaf},
	{ServerName: "E", ParentName: "C", Lag: 9, Users: 5, Position: irctree.PositionLeaf},
	{ServerName: "F", ParentName: "C", Lag: 9, Users: 14, Position: irctree.PositionLeaf},
}

func TestRenderDot(t *testing.T) {
	if ret := BuildDot(exampleServers, true).String(); ret == "" {
		t.Error("BuildDot returned an empty response")
	}
}

func TestRenderJson(t *testing.T) {
	tree := irctree.BuildTree(exampleServers)
	if ret := BuildJson(tree, true); ret == "" {
		t.Error("BuildJson returned an empty response")
	}
}
