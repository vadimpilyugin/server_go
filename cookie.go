package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"debug_print_go"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	authPage      = "/__auth__"
	usernameField = "username"
	passwordField = "password"
	cookieName    = "SESSID"
)

const (
	messageCheck       = "check"
	messageGen         = "gen"
	cookieExists       = "true"
	cookieDoesNotExist = "false"
)

var cookieCh chan string = func() chan string {
	var ch chan string = make(chan string)

	go func() {
		rand.Seed(time.Now().Unix())
		var cookies map[string]bool = make(map[string]bool)

		for {
			msg := <-ch
			if msg == messageCheck {
				ch <- fmt.Sprintf("%v", cookies[<-ch])
			} else if msg == messageGen {
				data := []byte(config.Auth.Password + fmt.Sprintf("%d", rand.Uint64()))
				cookie := fmt.Sprintf("%x", sha256.Sum256(data))
				cookies[cookie] = true
				ch <- cookie
			}
		}
	}()

	return ch
}()
var cookieMutex sync.Mutex

func trySetCookie(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != authPage {
		printer.Note("Got request without cookie, redirecting to "+authPage, "Auth")
		http.Redirect(w, r, authPage, http.StatusFound)
	} else if r.Method == "GET" {
		printer.Note("GET request to auth page")
		buf := new(bytes.Buffer)
		generateAuthPage(buf)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeContent(w, r, "", serverStartTime, bytes.NewReader(buf.Bytes()))
	} else if r.Method == "POST" {
		if len(r.Form["username"]) > 0 && r.Form["username"][0] == config.Auth.Username &&
			len(r.Form["password"]) > 0 && r.Form["password"][0] == config.Auth.Password {

			cookieMutex.Lock()
			cookieCh <- messageGen
			http.SetCookie(w, &http.Cookie{Name: cookieName, Value: <-cookieCh})
			cookieMutex.Unlock()
			printer.Note("Got POST request, sent new cookie")
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			printer.Error("Wrong credentials")
			http.Error(w, "Wrong username or password", http.StatusUnauthorized)
		}
	}
}

func isCorrect(cookie string) bool {
	cookieMutex.Lock()
	cookieCh <- messageCheck
	cookieCh <- cookie
	exists := <-cookieCh
	cookieMutex.Unlock()
	return exists == cookieExists
}

var ErrAlreadyHasCookie = errors.New("Client already has cookies!")

func checkCookieInHeaders(r *http.Request) string {
	// cookie, err :=
	for _, cookie := range r.Cookies() {
		if cookie.Name == cookieName && isCorrect(cookie.Value) {
			return cookie.Value
		}
	}
	return ""
}

func checkCookieInUrl(r *http.Request) (string, bool) {
	if len(r.Form[cookieName]) == 0 {
		return "", false
	}
	cookie := r.Form[cookieName][0]
	if isCorrect(cookie) {
		return cookie, true
	}
	return cookie, false
}

func checkCookie(w http.ResponseWriter, r *http.Request) (string, error) {

	// check headers for cookies
	headerCookie := checkCookieInHeaders(r)
	urlCookie, isUrlCookieCorrect := checkCookieInUrl(r)

	if isUrlCookieCorrect && headerCookie == "" {
		// Logging
		printer.Note("Renewing Cookie header", "Cookie-auth")
		//
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: urlCookie})
		return urlCookie, nil
	}
	if headerCookie != "" {
		// Logging
		if isUrlCookieCorrect {
			printer.Note("direct link access, but has valid Cookie header", "Cookie-auth")
		} else if urlCookie == "" {
			printer.Note("normal request", "Cookie-auth")
		} else {
			printer.Note("old direct link, but request is from authenticated Client!", "Cookie-auth")
		}
		//
		return headerCookie, nil
	}

	// redirect to auth page

	trySetCookie(w, r)
	return "", http.ErrNoCookie
}
