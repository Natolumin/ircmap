package irctree

import (
	"encoding/xml"
)

const (
	PositionUnknown = iota
	PositionHub
	PositionLeaf
)

// Server represents the information that can be obtained about a server
// xml values map to those in the output of inspircd's m_httpd_stats
// json values map to those used in the JS vizualisation
type Server struct {
	XMLName     xml.Name `xml:"server" json:"-"`
	ServerName  string   `xml:"servername" json:"id"`
	ParentName  string   `xml:"parentname" json:"-"`
	Label       string   `xml:"-" json:"label"`
	Lag         int      `xml:"lagmillisecs" json:"lagmillisecs"`
	Users       int      `xml:"usercount" json:"usercount"`
	Transit     int      `xml:"-" json:"-"`
	Description string   `xml:"gecos" json:"desc"`
	Descb64     string   `xml:",chardata" json:"-"`
	Position    int      `xml:"-" json:"group"`
}

// ServerTree is a node from a "doubly-linked tree" (a tree where each node has both references to its parent and its
// children
type ServerTree struct {
	Parent *Link
	Server
	Children []Link
}

type Link struct {
	Lag     int
	Transit int
	*ServerTree
}

// Servers is a helper structure to quickly iterate (in no particular order) on the contents of a ServerTree
type Servers struct {
	Root   *ServerTree
	Lookup map[string]*ServerTree
}
