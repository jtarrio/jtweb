package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"

	"jacobo.tarrio.org/jtweb/comments/web"
	"jacobo.tarrio.org/jtweb/config"
	"jacobo.tarrio.org/jtweb/config/fromflags"
	"jacobo.tarrio.org/jtweb/languages"
	"jacobo.tarrio.org/jtweb/webcontent"

	"github.com/rs/cors"
)

var flagServerAddress = flag.String("server_address", "127.0.0.1:8080", "The address where the server will be listening.")

func getOrigins(cfg config.Config) ([]string, error) {
	origins := map[string]bool{}
	for _, lang := range languages.AllLanguages() {
		uri, err := url.Parse(cfg.Site(lang).WebRoot())
		if err != nil {
			return nil, err
		}
		origin := url.URL{
			Scheme: uri.Scheme,
			Host:   uri.Host,
		}
		origins[origin.String()] = true
	}
	keys := make([]string, 0, len(origins))
	for k := range origins {
		keys = append(keys, k)
	}
	return keys, nil
}

func main() {
	flag.Parse()

	cfg, err := fromflags.GetConfig()
	if err != nil {
		panic(err)
	}

	if !cfg.Comments().Present() {
		panic("Comments not configured")
	}

	origins, err := getOrigins(cfg)
	if err != nil {
		panic(err)
	}
	corsChecker := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost},
		AllowCredentials: false,
	})

	adminChecker := web.NewAdminChecker(cfg.Comments().AdminPassword())

	mux := http.NewServeMux()
	mux.Handle("/_/", http.StripPrefix("/_", web.Serve(cfg.Comments().Service(), adminChecker)))
	mux.Handle("/comments.js", webcontent.ServeCommentsJs())
	mux.Handle("/admin.html", adminChecker.RequiringAdmin(webcontent.ServeAdminHtml()))
	mux.Handle("/admin.js", adminChecker.RequiringAdmin(webcontent.ServeAdminJs()))
	server := &http.Server{Addr: *flagServerAddress, Handler: corsChecker.Handler(mux)}
	log.Printf("Now serving on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
