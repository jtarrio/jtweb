package main

import (
	"flag"
	"log"
	"net/http"

	"jacobo.tarrio.org/jtweb/comments/web"
	"jacobo.tarrio.org/jtweb/config/fromflags"
	"jacobo.tarrio.org/jtweb/webcontent"
)

var flagServerAddress = flag.String("server_address", "127.0.0.1:8080", "The address where the server will be listening.")
var flagContentRoot = flag.String("content_root", "./", "The location where the static content resides.")

func main() {
	flag.Parse()

	cfg, err := fromflags.GetConfig()
	if err != nil {
		panic(err)
	}

	if !cfg.Comments().Present() {
		panic("Comments not configured")
	}

	mux := http.NewServeMux()
	mux.Handle("/_/", http.StripPrefix("/_", web.Serve(cfg.Comments().Service())))
	mux.Handle("/comments.js", webcontent.ServeCommentsJs())
	mux.Handle("/", http.FileServer(http.Dir(*flagContentRoot)))
	server := &http.Server{Addr: *flagServerAddress, Handler: mux}
	log.Printf("Now serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
