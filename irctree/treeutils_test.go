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

func TestNormalize(t *testing.T) {
	tree := BuildTree(exampleServers)

	slice := tree.Slice()
	for i, node := range slice {
		if node != normalizedServers[i] {
			t.Errorf("Server mismatch: expected %s got %s\n", normalizedServers[i], node)
		}
	}
	if len(slice) != len(normalizedServers) {
		t.Errorf("Missing servers: %d out of %d present\n", len(slice), len(normalizedServers))
	}
}
