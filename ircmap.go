package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Natolumin/ircmap/formatters"
	"github.com/Natolumin/ircmap/irctree"
)

var (
	serverDomain  = ".rezosup.org"
	hubPrefix     = "hub."
	leafPrefix    = "irc."
	statsURL      = "http://localhost:8000/stats"
	listenAddress string
	clientTimeout = time.Second
	displayAll    = false
	tlsCert       string
	tlsKey        string
)

func init() {

	var stringTimeout string

	configValues := []struct {
		Env string
		Var *string
	}{
		{Env: "IRCMAP_SERVER_DOMAIN", Var: &serverDomain},
		{Env: "IRCMAP_HUB_PREFIX", Var: &hubPrefix},
		{Env: "IRCMAP_LEAF_PREFIX", Var: &leafPrefix},
		{Env: "IRCMAP_STATS_URL", Var: &statsURL},
		{Env: "IRCMAP_LISTEN_ADDRESS", Var: &listenAddress},
		{Env: "IRCMAP_CLIENT_TIMEOUT", Var: &stringTimeout},
		{Env: "IRCMAP_TLS_CERT", Var: &tlsCert},
		{Env: "IRCMAP_TLS_KEY", Var: &tlsKey},
	}
	for _, val := range configValues {
		if env := os.Getenv(val.Env); env != "" {
			*val.Var = env
		}
	}

	if stringTimeout != "" {
		var err error
		clientTimeout, err = time.ParseDuration(stringTimeout)
		if err != nil {
			clientTimeout = time.Second
		}
	}

}

type Stats struct {
	XMLName    xml.Name         `xml:"inspircdstats"`
	ServerList []irctree.Server `xml:"serverlist>server"`
}

func main() {

	var output = flag.String("o", "raw", "Output format (dot, json, raw)")
	var stdin = flag.Bool("stdin", false, "Use stdin instead of network for map source")
	flag.BoolVar(&displayAll, "all", false, "Don't scrub unrecognized nodes")
	flag.StringVar(&listenAddress, "listen", "", "Address to listen on")
	flag.StringVar(&statsURL, "url", statsURL, "Location of the inspircd stats page")
	var serve = flag.Bool("serve", true, "Run as server")
	flag.Parse()

	if *stdin {
		*serve = false
	}
	if *serve {
		err := doServe()
		log.Fatalf("Server encountered error: %s", err)
	}

	var (
		tree *irctree.Servers
		err  error
	)
	if *stdin {
		tree, err = decodeMap(os.Stdin)
	} else {
		tree, err = GetMap(statsURL, clientTimeout)
	}
	if err != nil {
		log.Fatalf("Error building the map: %s", err)
	}
	switch *output {
	case "json":
		fmt.Print(formatters.BuildJson(tree, displayAll))
	case "dot":
		fmt.Print(formatters.BuildDot(tree.Slice(), displayAll))
	default:
		fmt.Print(tree.String())
	}
}

func doServe() error {
	http.Handle("/", protoHandler{
		UpstreamURL: "http://localhost:8000/stats",
		Timeout:     clientTimeout,
	})
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	http.Handle("/ircmap.html", http.FileServer(http.Dir("static/")))

	var err error
	if tlsCert != "" && tlsKey != "" {
		err = http.ListenAndServeTLS(listenAddress, tlsCert, tlsKey, nil)
	} else {
		err = http.ListenAndServe(listenAddress, nil)
	}
	return err
}

func decodeMap(read io.Reader) (*irctree.Servers, error) {

	dec := xml.NewDecoder(read)
	var ircmap Stats
	err := dec.Decode(&ircmap)
	if err != nil {
		return nil, err
	}
	scrubValues(ircmap.ServerList)
	tree := irctree.BuildTree(ircmap.ServerList)
	return tree, nil
}

func scrubValues(ircmap []irctree.Server) {
	for i := range ircmap {
		node := &ircmap[i]
		node.Label = strings.TrimSuffix(node.ServerName, serverDomain)
		if strings.HasPrefix(node.Label, hubPrefix) {
			node.Label = strings.TrimPrefix(node.Label, hubPrefix)
			node.Position = irctree.PositionHub
		} else if strings.HasPrefix(node.Label, leafPrefix) {
			node.Position = irctree.PositionLeaf
			node.Label = strings.TrimPrefix(node.Label, leafPrefix)
		}
	}
}
