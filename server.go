package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
)

var htmlTemplate = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.FullPath}} {{.RepoType}} {{.Repo}}">
<meta http-equiv="refresh" content="0; url={{.Redirect}}">
</head>
<body>
Nothing to see here; <a href="{{.Redirect}}">move along</a>.
</body>
`))

var cache = make(map[string][]byte)

func listenAndServe(c *config) error {
	if c.HTTP.EnableInsecure {
		go startHTTPRedirector(c)
	}

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", c.HTTP.Address, c.HTTP.TLSPort),
		Handler: makeServe(c),
	}
	return s.ListenAndServeTLS(c.HTTP.TLS.Cert, c.HTTP.TLS.Key)
}

func startHTTPRedirector(c *config) {
	s := &http.Server{
		Addr: fmt.Sprintf("%s:%d", c.HTTP.Address, c.HTTP.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c.HTTP.Port == 80 && c.HTTP.TLSPort == 443 {
				http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
				return
			}

			host, _, _ := net.SplitHostPort(r.Host)
			http.Redirect(w, r, fmt.Sprintf("https://%s:%d+%s", host, c.HTTP.TLSPort, r.RequestURI), http.StatusMovedPermanently)
		}),
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func makeServe(c *config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		handleRequest(c, w, r)
	})
}

func logRequest(r *http.Request) {
	host := r.Host
	if r.URL.Port() != "" {
		host, _, _ = net.SplitHostPort(r.Host)
	}
	r.Host = host

	log.Printf(
		"%s %s %s \"%s\"",
		host,
		r.RemoteAddr,
		r.Method,
		r.URL.Path,
	)
}

func handleRequest(c *config, w http.ResponseWriter, r *http.Request) {
	path := r.Host + r.URL.Path

	// Check cache
	if resp, ok := cache[path]; ok {
		w.Write(resp)
		return
	}

	// Cache miss, generate template
	repo, exists := c.paths[path]
	if !exists {
		return
	}

	var buf bytes.Buffer
	if err := htmlTemplate.Execute(&buf, repo); err != nil {
		log.Println(err.Error())
		return
	}

	// Add to cache
	cache[path] = buf.Bytes()
	w.Write(buf.Bytes())
}
