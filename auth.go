package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type authbase struct {
	conf *oauth2.Config
	URL  string
}

func newauthbase() *authbase {
	conf := &oauth2.Config{
		ClientID:     os.Getenv("AUTH_GITHUB_KEY"),
		ClientSecret: os.Getenv("AUTH_GITHUB_SECRET"),
		Scopes:       []string{"read:user"},
		Endpoint:     github.Endpoint,
	}
	return &authbase{
		conf: conf,
		// Redirect user to consent page to ask for permission
		// for the scopes specified above.
		URL: conf.AuthCodeURL("state", oauth2.AccessTypeOnline),
	}
}

func (a *Evaluator) handleAuth(rw http.ResponseWriter, r *http.Request) {
	type Auth struct {
		User    User
		AuthURL string
	}
	if path.Base(r.URL.Path) == "callback" {
		query := r.URL.Query()
		ctx := context.Background()
		code := query.Get("code")
		tok, err := a.auth.conf.Exchange(ctx, code)
		if err != nil {
			httpErr(rw, "while exchanging tokens", err, http.StatusInternalServerError)
			return
		}
		// We fetch user information. See https://docs.github.com/en/rest/guides/basics-of-authentication.
		cl := a.auth.conf.Client(ctx, tok)
		resp, err := cl.Get("https://api.github.com/user")
		if err != nil {
			httpErr(rw, "while accessing oauth2 provider API", err, http.StatusInternalServerError)
			return
		}
		var auth Auth
		err = json.NewDecoder(resp.Body).Decode(&auth.User)
		if err != nil {
			httpErr(rw, "decoding oauth2 provider API response", err, http.StatusInternalServerError)
			return
		}
		setUserSession(rw, auth.User.Email)
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
		return
	}
	u, err := getUserSession(r)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("got user ", u)
	}
	err = a.tmpls.Lookup("auth.tmpl").Execute(rw, Auth{
		AuthURL: a.auth.URL,
		User:    u,
	})
	if err != nil {
		log.Println(err)
	}
}

type User struct {
	Email string
	Login string
}

func setUserSession(rw http.ResponseWriter, email string) {
	if email == "" {
		return // Don't set empty cookie.
	}
	const hour = 60 * 60
	cookie := &http.Cookie{
		Name:    "user_email",
		Path:    "/", // very important, or else path will be autoset to current path.
		Value:   email,
		MaxAge:  3 * hour,
		Expires: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	log.Printf("user %s logging in", email)
	http.SetCookie(rw, cookie)
}

func getUserSession(r *http.Request) (User, error) {
	c, err := r.Cookie("user_email")
	if err != nil {
		return User{}, err
	}
	return User{Email: c.Value}, nil
}
