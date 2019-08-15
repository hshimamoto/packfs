// packfs / packweb
// MIT License Copyright(c) 2019 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
package main

import (
    "log"
    "net/http"
    "os"

    "github.com/hshimamoto/packfs/pack"
)

type PackServ struct {
    p *pack.Pack
}

func (ps *PackServ)index() string {
    html := "<html>"
    html += "<head>"
    html += "<title>PackWeb</title>"
    html += "</head>"
    html += "<body>"
    html += "<ul>"
    for _, f := range ps.p.Files {
	html += "<li>"
	html += `<a href="` + f.Name + `">` + f.Name + "</a>"
	html += "</li>"
    }
    html += "</ul>"
    html += "</body>"
    html += "</html>"
    return html
}

func (ps *PackServ)Handler(w http.ResponseWriter, req *http.Request) {
    url := req.URL
    log.Printf("path %s", url.Path)
    if url.Path == "/" {
	w.Write([]byte(ps.index()))
	return
    }
    for _, f := range ps.p.Files {
	name := url.Path[1:] // remove first '/'
	if name == f.Name {
	    w.Write(f.Bytes())
	    f.Clear()
	    return
	}
    }
    w.WriteHeader(http.StatusNotFound)
}

func (ps *PackServ)Serve(addr string) {
    http.ListenAndServe(addr, http.HandlerFunc(ps.Handler))
}

func NewPackServ(path string) (*PackServ, error) {
    p, err := pack.OpenPack(path)
    if err != nil {
	return nil, err
    }
    return &PackServ{
	p: p,
    }, nil
}

func main() {
    if len(os.Args) < 3 {
	log.Fatal("Usage:\n\tpackweb packfile LISTEN")
    }
    ps, err := NewPackServ(os.Args[1])
    if err != nil {
	log.Fatal(err)
    }
    ps.Serve(os.Args[2])
}
