package main

import (
	"fmt"
)

type ServerTree struct {
	Parent   *ServerTree
	Node     Server
	Children []*ServerTree
}

type Servers struct {
	Root   *ServerTree
	Lookup map[string]*ServerTree
}

func (s *Servers) Add(node *Server) error {
	if parent := s.Lookup[node.ParentName]; parent != nil {
		stree := ServerTree{
			Parent:   parent,
			Node:     *node,
			Children: []*ServerTree{},
		}
		s.Lookup[node.ServerName] = &stree
		parent.Children = append(parent.Children, &stree)
		return nil
	}
	return fmt.Errorf("Parent not found: %s", node.ParentName)
}

func (s *ServerTree) GetList() []Server {
	ret := []Server{s.Node}
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
	*acc += padding + s.Node.ServerName + "\n"
	for i, node := range s.Children {
		node.string(acc, depth+1, i == len(s.Children)-1)
	}
	return *acc
}

func buildTree(ircmap []Server) *Servers {
	rootIndex := findRoot(ircmap)
	root := ServerTree{
		Node:     ircmap[rootIndex],
		Children: []*ServerTree{},
	}
	s := Servers{
		Root: &root,
		Lookup: map[string]*ServerTree{
			root.Node.ServerName: &root,
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

func findRoot(ircmap []Server) int {
	for i, node := range ircmap {
		if node.ParentName == "" {
			return i
		}
	}
	return 0
}

func (t *Servers) FlattenLag() {
	t.Root.flattenLag(0)
}

func (t *ServerTree) flattenLag(accLag int) {
	t.Node.Lag -= accLag
	for _, node := range t.Children {
		node.flattenLag(t.Node.Lag)
	}
}

// Normalize finds a better root than the current one and shifts to there, if it exists
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
		Lookup: t.Lookup,
		Root: &ServerTree{
			Parent:   nil,
			Node:     maxD.Node,
			Children: append(maxD.Children, rerootTree(maxD.Parent, maxD)),
		}}
	servers.Root.Node.ParentName = ""

	return &servers
}

func rerootTree(node, newParent *ServerTree) *ServerTree {
	rerootedTree := ServerTree{
		Parent: newParent,
		Node:   node.Node,
	}
	rerootedTree.Node.ParentName = newParent.Node.ServerName
	copy(rerootedTree.Children, node.Children)
	for i, _ := range node.Children {
		if node.Children[i] == newParent {
			node.Children[i] = node.Children[len(node.Children)-1]
			node.Children = node.Children[:len(node.Children)-1]
		}
	}
	if node.Parent != nil {
		rerootedTree.Children = append(node.Children, rerootTree(node.Parent, node))
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

func (t *Servers) BuildTransit() {
	t.Root.buildTransit()
}

func (t *ServerTree) buildTransit() {
	for _, child := range t.Children {
		child.buildTransit()
	}
	acc := t.Node.Users
	if t.Node.Position == PositionHub {
		if t.Parent != nil && t.Parent.Node.Position == PositionLeaf {
			acc = t.Parent.Node.Users
		}
		for _, child := range t.Children {
			if child.Node.Position == PositionLeaf {
				acc += child.Node.Users
			}
		}
	}
	t.Node.Transit = acc
}
