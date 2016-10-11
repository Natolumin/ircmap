package main

import (
	"fmt"
)

type ServerTree struct {
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
