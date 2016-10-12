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
