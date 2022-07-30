package main

import (
	"log"
	"net/http"
	"strings"
)

type baseMiddleware struct {
	onRequest func(rw http.ResponseWriter, r *http.Request)
	handler   http.Handler
	onExit    func(rw http.ResponseWriter, r *http.Request)
}

func (rh *baseMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if rh.onRequest != nil {
		rh.onRequest(rw, r)
	}
	rh.handler.ServeHTTP(rw, r)
	if rh.onExit != nil {
		rh.onExit(rw, r)
	}
}

func userMiddleware(h http.Handler) *baseMiddleware {
	return &baseMiddleware{
		handler: h,
		onRequest: func(rw http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/assets/css/") {
				rw.Header().Add("Content-Type", "text/css")
			} else if strings.HasPrefix(r.URL.Path, "/assets/js/") {
				rw.Header().Add("Content-Type", "application/js")
			}
		},
	}
}

func httpErr(rw http.ResponseWriter, msg string, err error, status int) {
	if err != nil {
		msg += ": " + err.Error()
	}
	log.Println("error in request: ", msg)
	http.Error(rw, msg, status)
}
