package main

import (
	"flag"
	"log"
	"net/http"

	"jacobo.tarrio.org/jtweb/comments"
	"jacobo.tarrio.org/jtweb/comments/service"
	"jacobo.tarrio.org/jtweb/comments/testing"
	"jacobo.tarrio.org/jtweb/comments/web"
	"jacobo.tarrio.org/jtweb/config/fromflags"
	"jacobo.tarrio.org/jtweb/site"
)

var flagServerAddress = flag.String("server_address", "127.0.0.1:8080", "The address where the server will be listening.")

func main() {
	flag.Parse()

	cfg, err := fromflags.GetConfig()
	if err != nil {
		panic(err)
	}

	content, err := site.Read(cfg)
	if err != nil {
		panic(err)
	}

	engine := testing.NewMemoryEngine()
	for _, post := range content.Pages {
		engine.AddPost(comments.PostId(post.Name))
	}

	commentsService := service.NewCommentsService(engine)

	mux := http.NewServeMux()
	mux.Handle("/_/", http.StripPrefix("/_", web.Serve(commentsService)))
	server := &http.Server{Addr: *flagServerAddress, Handler: mux}
	log.Printf("Now serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
