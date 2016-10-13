package irctree

import (
	"fmt"
)

func (s *Servers) Add(node *Server) error {
	if parent := s.Lookup[node.ParentName]; parent != nil {
		stree := ServerTree{
			Parent: &Link{
				ServerTree: parent,
				Lag:        (node.Lag - parent.Lag),
			},
			Server:   *node,
			Children: []Link{},
		}
		s.Lookup[node.ServerName] = &stree
		parent.Children = append(parent.Children, Link{
			Lag:        abs(node.Lag - parent.Lag),
			ServerTree: &stree,
		})
		return nil
	}
	return fmt.Errorf("Parent not found: %s", node.ParentName)
}

func (s *ServerTree) GetList() []Server {
	ret := []Server{s.Server}
	for _, child := range s.Children {
		ret = append(ret, child.GetList()...)
	}
	return ret
}

func (s Servers) Slice() []Server {
	return s.Root.GetList()
}

func (s *Servers) String() string {
	ret := ""
	return s.Root.string(&ret, 0, false)
}

func (s *ServerTree) string(acc *string, depth int, last bool) string {
	padding := ""
	for i := 0; i < depth-1; i++ {
		padding += "│  "
	}
	if depth > 0 {
		if last && len(s.Children) == 0 {
			padding += "└──"
		} else {
			padding += "├──"
		}
	}
	*acc += padding + fmt.Sprint(s.ServerName) + "\n"
	for i, node := range s.Children {
		node.string(acc, depth+1, i == len(s.Children)-1)
	}
	return *acc
}

func buildTree(ircmap []Server) *Servers {
	rootIndex := findRoot(ircmap)
	root := ServerTree{
		Server:   ircmap[rootIndex],
		Children: []Link{},
	}
	s := Servers{
		Root: &root,
		Lookup: map[string]*ServerTree{
			root.ServerName: &root,
		},
	}

	for changed := true; changed; {
		changed = false
		for _, node := range ircmap {
			if s.Lookup[node.ServerName] == nil {
				err := s.Add(&node)
				if err == nil {
					changed = true
				}
			}
		}
	}

	return &s
}

func BuildTree(ircmap []Server) (s *Servers) {
	s = buildTree(ircmap).Normalize()
	s.buildTransit()
	return s
}

func findRoot(ircmap []Server) int {
	for i, node := range ircmap {
		if node.ParentName == "" {
			return i
		}
	}
	return 0
}

// Normalize finds a better root than the current one and shifts to there, if it exists
//! Normalize destroys transit information
func (t *Servers) Normalize() *Servers {
	// Find the node with the highest degree
	maxD := t.Root
	for _, node := range t.Lookup {
		if deg(node) > deg(maxD) {
			maxD = node
		}
	}
	if maxD == t.Root {
		return t
	}
	// Pivot the tree to have that node be the root
	servers := Servers{Root: rerootTree(maxD, nil)}
	servers.rebuildLookup()
	return &servers
}

func (s *Servers) rebuildLookup() {
	s.Lookup = make(map[string]*ServerTree)
	s.Root.dfmap(func(t *ServerTree) {
		s.Lookup[t.ServerName] = t
	})
}

func (t *ServerTree) dfmap(fn func(*ServerTree)) {
	fn(t)
	for i, _ := range t.Children {
		t.Children[i].ServerTree.dfmap(fn)
	}
}

func abs(a int) int {
	if a > 0 {
		return a
	} else {
		return -a
	}
}
func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func rerootTree(node, newParent *ServerTree) *ServerTree {

	// build this item, without children
	var rerootedTree ServerTree
	if newParent != nil {
		rerootedTree = ServerTree{
			Parent: &Link{
				ServerTree: newParent,
			},
			Server: node.Server,
		}
		rerootedTree.ParentName = newParent.ServerName
	} else {
		rerootedTree = ServerTree{
			Parent: nil,
			Server: node.Server,
		}
		rerootedTree.ParentName = ""
		rerootedTree.Lag = 0
	}

	//Correct Lag with information from the ex-child-new-parent
	for _, child := range node.Children {
		if newParent != nil && child.ServerName == newParent.ServerName {
			// Correct lag with information from the new-parent-ex-child though
			rerootedTree.Parent.Lag = child.Lag
			rerootedTree.Lag = newParent.Lag + child.Lag
		}
	}

	// deep copy children with references to the new node
	for _, child := range node.Children {
		// Don't copy the new parent from the children table
		if newParent == nil || child.ServerName != newParent.ServerName {
			rerootedTree.Children = append(rerootedTree.Children, copychild(&child, &rerootedTree))
		}
	}

	// Move the ex-parent to the child list, rerooting it at the same time (providing we are rerooting not merely
	// copying)
	if node.Parent != nil && (newParent == nil || node.Parent.ServerName != newParent.ServerName) {
		rerootedTree.Children = append(rerootedTree.Children, copychild(node.Parent, &rerootedTree))
	}
	return &rerootedTree
}

func copychild(link *Link, parent *ServerTree) Link {
	copiedLink := Link{
		ServerTree: rerootTree(link.ServerTree, parent),
		Lag:        link.Lag,
	}
	copiedLink.ServerTree.Lag = parent.Lag + link.Lag
	return copiedLink
}

func deg(t *ServerTree) (deg int) {
	if t.Parent != nil {
		deg = 1 + len(t.Children)
	} else {
		deg = len(t.Children)
	}
	return deg
}

func (t *Servers) buildTransit() {
	t.Root.buildTransit()
}

func (t *ServerTree) buildTransit() {
	for _, child := range t.Children {
		child.buildTransit()
	}
	acc := t.Users
	if t.Position == PositionHub {
		if t.Parent != nil && t.Parent.Position == PositionLeaf {
			acc = t.Parent.Users
		}
		for _, child := range t.Children {
			if child.Position == PositionLeaf {
				acc += child.Users
			}
		}
	}
	t.Transit = acc
	classifyLink := func(origin *Server, currentLink *Link) {
		switch {
		case origin.Position == PositionLeaf || origin.Position == PositionUnknown:
			currentLink.Transit = origin.Transit
		case origin.Position == PositionHub && currentLink.Position == PositionHub:
			currentLink.Transit = max(origin.Transit, currentLink.ServerTree.Transit)
		case currentLink.Position == PositionLeaf || currentLink.Position == PositionUnknown:
			currentLink.Transit = currentLink.ServerTree.Transit
		}
	}
	for i, _ := range t.Children {
		classifyLink(&t.Server, &t.Children[i])
		t.Children[i].Parent = &Link{
			ServerTree: t,
			Lag:        t.Children[i].Lag,
			Transit:    t.Children[i].Transit,
		}
	}
}
