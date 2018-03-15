package main

import (
  "net/http"
  "printer"
  "math/rand"
  "time"
  "crypto/sha256"
  "sync"
  "fmt"
  "bytes"
  "errors"
)

const (
  authPage = "/__auth__"
  usernameField = "username"
  passwordField = "password"
  cookieName = "SESSID"
)

const (
  messageCheck = "check"
  messageGen = "gen"
  cookieExists = "true"
  cookieDoesNotExist = "false"
)

var cookieCh chan string = func () (chan string) {
  var ch chan string = make(chan string)

  go func() {
    rand.Seed(time.Now().Unix())
    var cookies map[string]bool = make(map[string]bool)

    for {
      msg := <-ch
      if msg == messageCheck {
        ch <- fmt.Sprintf("%v",cookies[<-ch])
      } else if msg == messageGen {
        data := []byte(config.Auth.Password+fmt.Sprintf("%d",rand.Uint64()))
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
    http.Redirect(w,r,authPage,http.StatusFound)
  } else if r.Method == "GET" {
    printer.Note("GET request to auth page")
    buf := new(bytes.Buffer)
    generateAuthPage(buf)
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    http.ServeContent(w, r, "", serverStartTime, bytes.NewReader(buf.Bytes()))
  } else if r.Method == "POST" {
    if  len(r.Form["username"]) > 0 && r.Form["username"][0] == config.Auth.Username && 
        len(r.Form["password"]) > 0 && r.Form["password"][0] == config.Auth.Password {

      cookieMutex.Lock()
      cookieCh <- messageGen
      http.SetCookie(w,&http.Cookie{Name:cookieName, Value:<-cookieCh})
      cookieMutex.Unlock()
      printer.Note("Got POST request, sent new cookie")
      http.Redirect(w,r,"/",http.StatusFound)
    } else {
      printer.Error("Wrong credentials")
      http.Error(w, "Wrong username or password", http.StatusUnauthorized)
    }
  }
}

var ErrAlreadyHasCookie = errors.New("Client already has cookies!")

func checkCookie(w http.ResponseWriter, r *http.Request) error {
  if cookie, err := r.Cookie(cookieName); err != nil {
    // redirect to auth
    trySetCookie(w,r)
    return http.ErrNoCookie
  } else {
    // check correctness
    cookieMutex.Lock()
    cookieCh <- messageCheck
    cookieCh <- cookie.Value
    exists := <-cookieCh
    cookieMutex.Unlock()

    if exists == "false" {
      // redirect to auth
      trySetCookie(w,r)
      return http.ErrNoCookie
    }
    if r.RequestURI == authPage{
      printer.Note("Already has valid cookies, but sent another request!")
      http.Redirect(w,r,"/",http.StatusFound)
      return ErrAlreadyHasCookie
    }
    return nil
  }
}