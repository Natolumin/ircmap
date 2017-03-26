package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/Natolumin/ircmap/formatters"
	"github.com/Natolumin/ircmap/irctree"
)

type protoHandler struct {
	UpstreamURL string
	Timeout     time.Duration
}

func (p protoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(404)
		return
	}
	if r.URL.Path == "/" {
		w.Header().Add("Location", "ircmap.html")
		w.WriteHeader(301)
		return
	}

	var (
		ircmap *irctree.Servers
		err    error
	)
	if ircmap, err = GetMap(p.UpstreamURL, p.Timeout); err != nil {
		log.Printf("Could not get map from %s, got error %s\n", p.UpstreamURL, err)
		w.WriteHeader(500)
		return
	}
	switch r.URL.Path {
	case "/map.json":
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, formatters.BuildJson(ircmap, displayAll))
	case "/map.dot":
		w.Header().Set("Content-Type", "text/vnd.graphviz")
		io.WriteString(w, formatters.BuildDot(ircmap.Slice(), displayAll).String())
	case "/map.png":
		if !doPng {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		cmd := exec.Command("dot", dotOptions...)
		out, err := formatters.BuildPNG(formatters.BuildDot(ircmap.Slice(), displayAll), cmd)
		if err != nil {
			w.WriteHeader(500)
		} else {
			w.Write(out)
		}
	case "/map.txt":
		io.WriteString(w, ircmap.String())
	default:
		w.WriteHeader(404)
	}
}

func GetMap(url string, timeout time.Duration) (*irctree.Servers, error) {

	client := http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return errors.New("Unexpected Redirect")
		},
		Jar:     nil,
		Timeout: timeout,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	servs, err := decodeMap(resp.Body)
	resp.Body.Close()
	return servs, err
}
