package irctree

import (
	"fmt"
)

func (s *Servers) Add(node *Server) error {
	if parent := s.Lookup[node.ParentName]; parent != nil {
		stree := ServerTree{
			Parent: &Link{
				End: parent,
				Lag: (node.Lag - parent.Lag),
			},
			Server:   *node,
			Children: []Link{},
		}
		s.Lookup[node.ServerName] = &stree
		parent.Children = append(parent.Children, Link{
			Lag: abs(node.Lag - parent.Lag),
			End: &stree,
		})
		return nil
	}
	return fmt.Errorf("Parent not found: %s", node.ParentName)
}

func (s *ServerTree) GetList() []Server {
	ret := []Server{s.Server}
	for _, child := range s.Children {
		ret = append(ret, child.End.GetList()...)
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
		node.End.string(acc, depth+1, i == len(s.Children)-1)
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
	// Since the tree changes we can assume maxD has a parent
	servers := Servers{
		Root: &ServerTree{
			Parent: nil,
			Server: maxD.Server,
			Children: append(maxD.Children, Link{
				End: rerootTree(maxD.Parent.End, maxD),
				Lag: abs(maxD.Parent.End.Lag - maxD.Lag),
			}),
		}}
	servers.Root.ParentName = ""
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
	for _, child := range t.Children {
		child.End.dfmap(fn)
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
	rerootedTree := ServerTree{
		Parent: &Link{
			End: newParent,
			Lag: abs(newParent.Lag - node.Lag),
		},
		Server: node.Server,
	}
	rerootedTree.ParentName = newParent.ServerName
	copy(rerootedTree.Children, node.Children)
	for i, _ := range node.Children {
		if node.Children[i].End == newParent {
			node.Children[i] = node.Children[len(node.Children)-1]
			node.Children = node.Children[:len(node.Children)-1]
		}
	}
	if node.Parent != nil {
		rerootedTree.Children = append(node.Children, Link{
			End: rerootTree(node.Parent.End, node),
			Lag: abs(node.Parent.End.Lag - node.Lag),
		})
	}
	return &rerootedTree
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
		child.End.buildTransit()
	}
	acc := t.Users
	if t.Position == PositionHub {
		if t.Parent != nil && t.Parent.End.Position == PositionLeaf {
			acc = t.Parent.End.Users
		}
		for _, child := range t.Children {
			if child.End.Position == PositionLeaf {
				acc += child.End.Users
			}
		}
	}
	t.Transit = acc
	classifyLink := func(origin *Server, currentLink *Link) {
		switch {
		case origin.Position == PositionLeaf || origin.Position == PositionUnknown:
			currentLink.Transit = origin.Transit
		case origin.Position == PositionHub && currentLink.End.Position == PositionHub:
			currentLink.Transit = max(origin.Transit, currentLink.End.Transit)
		case currentLink.End.Position == PositionLeaf || currentLink.End.Position == PositionUnknown:
			currentLink.Transit = currentLink.End.Transit
		}
	}
	for i, _ := range t.Children {
		classifyLink(&t.Server, &t.Children[i])
		t.Children[i].End.Parent = &Link{
			End:     t,
			Lag:     t.Children[i].Lag,
			Transit: t.Children[i].Transit,
		}
	}
}
