package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func ImageProxy(w http.ResponseWriter, r *http.Request) {
	query_url := r.URL.Query().Get("url")
	url, err := url.ParseRequestURI(query_url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := http.Get(url.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()
	content_type := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(content_type, "image") {
		http.Error(w, fmt.Sprintf("Invalid content-type: %s", content_type), http.StatusBadRequest)
		return
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, header := range []string{"Content-Type", "Cache-Control"} {
		w.Header().Add(header, resp.Header.Get(header))
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(b)
}
