package irctree

import (
	"testing"
)

// Our example tree is :
//      A
//    4/ \8
//    B   C__
//      2/ \1\1
//      D   E F

var exampleServers = []Server{
	{ServerName: "A", ParentName: "", Lag: 0, Users: 10, Position: PositionHub},
	{ServerName: "B", ParentName: "A", Lag: 4, Users: 5, Position: PositionLeaf},
	{ServerName: "C", ParentName: "A", Lag: 8, Users: 1, Position: PositionHub},
	{ServerName: "D", ParentName: "C", Lag: 10, Users: 20, Position: PositionLeaf},
	{ServerName: "E", ParentName: "C", Lag: 9, Users: 5, Position: PositionLeaf},
	{ServerName: "F", ParentName: "C", Lag: 9, Users: 14, Position: PositionLeaf},
}

// The normalized version would be:
//      C____
//    8/ \2\1\1
//    A   D E F
//  4/
//  B

var normalizedServers = []Server{
	{ServerName: "C", ParentName: "", Lag: 0, Users: 1, Transit: 40, Position: PositionHub},
	{ServerName: "D", ParentName: "C", Lag: 2, Users: 20, Transit: 20, Position: PositionLeaf},
	{ServerName: "E", ParentName: "C", Lag: 1, Users: 5, Transit: 5, Position: PositionLeaf},
	{ServerName: "F", ParentName: "C", Lag: 1, Users: 14, Transit: 14, Position: PositionLeaf},
	{ServerName: "A", ParentName: "C", Lag: 8, Users: 10, Transit: 15, Position: PositionHub},
	{ServerName: "B", ParentName: "A", Lag: 12, Users: 5, Transit: 5, Position: PositionLeaf},
}

var stringServers = `C
├──D
├──E
├──F
└──A
   └──B
`

func TestNormalize(t *testing.T) {
	tree := BuildTree(exampleServers)

	slice := tree.Slice()
	compareTrees(slice, normalizedServers, t)
}

func testNormalizeIdempotent(t *testing.T) {
	tree := BuildTree(exampleServers)
	normtree := tree.Normalize()
	compareTrees(tree.Slice(), normtree.Slice(), t)
}

func compareTrees(a, b []Server, t *testing.T) {
	for i, node := range a {
		if node != b[i] {
			t.Errorf("Server mismatch: expected %v got %v\n", b[i], node)
		}
	}
	if len(a) != len(b) {
		t.Errorf("Missing servers: %d out of %d present\n", len(a), len(b))
	}
}

func TestPrint(t *testing.T) {
	tree := BuildTree(exampleServers)

	if tree.String() != stringServers {
		t.Errorf(`String repr mismatch: expected :
%s
, got
%s`, stringServers, tree.String())
	}
}
