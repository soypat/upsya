package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

type authbase struct{}

func (sv *Server) handleAuth(rw http.ResponseWriter, r *http.Request) {
	var u User
	query := r.URL.Query()
	if query.Has("legajo") {
		uid, err := strconv.ParseUint(query.Get("legajo"), 10, 64)
		if err != nil || uid == 0 {
			sv.httpErr(rw, "legajo invalido", err, http.StatusBadRequest)
			return
		}
		u.ID = uid
		sv.auth.setUserSession(rw, u)
	} else {
		gotu, err := sv.auth.getUserSession(r)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("got user ", gotu)
			u = gotu
		}
	}
	err := sv.tmpls.Lookup("auth.tmpl").Execute(rw, struct{ User User }{
		User: u,
	})
	if err != nil {
		log.Println(err)
	}
}

type User struct {
	ID uint64
}

func (a *authbase) setUserSession(rw http.ResponseWriter, u User) {
	if u.ID == 0 {
		return // Don't set zero value.
	}
	const hour = 60 * 60
	cookie := &http.Cookie{
		Name:    "user_id",
		Path:    "/", // very important, or else path will be autoset to current path.
		Value:   strconv.FormatUint(u.ID, 10),
		MaxAge:  3 * hour,
		Expires: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	log.Printf("user %v logging in", u)
	http.SetCookie(rw, cookie)
}

func (a *authbase) getUserSession(r *http.Request) (User, error) {
	c, err := r.Cookie("user_id")
	if err != nil {
		return User{}, err
	}
	v, err := strconv.ParseUint(c.Value, 10, 64)
	if err == nil && v == 0 {
		return User{}, errors.New("user ID must be non-zero")
	}
	return User{ID: v}, err
}
