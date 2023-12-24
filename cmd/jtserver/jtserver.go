package main

import (
	"flag"
	"log"
	"net/http"

	"jacobo.tarrio.org/jtweb/comments/web"
	"jacobo.tarrio.org/jtweb/config/fromflags"
)

var flagServerAddress = flag.String("server_address", "127.0.0.1:8080", "The address where the server will be listening.")

func main() {
	flag.Parse()

	cfg, err := fromflags.GetConfig()
	if err != nil {
		panic(err)
	}

	if cfg.Comments() == nil {
		panic("Comments not configured")
	}

	mux := http.NewServeMux()
	mux.Handle("/_/", http.StripPrefix("/_", web.Serve(cfg.Comments().Service())))
	server := &http.Server{Addr: *flagServerAddress, Handler: mux}
	log.Printf("Now serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
