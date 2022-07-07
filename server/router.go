package main

import (
	"net/http"
)

func RouterAddRoutes(mux *http.ServeMux) {
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	RouterAddUserRoutes(mux)
	RouterAddBlogRoutes(mux)
}
