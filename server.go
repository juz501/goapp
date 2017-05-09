package main

import (
  "net/http"
  "os"

  "github.com/urfave/negroni"
  "github.com/unrolled/render"
)

func main() {
  rend := render.New(render.Options{IsDevelopment: true})
  mux := http.NewServeMux()

  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    rend.HTML(w, http.StatusOK, "index", "David")
  })

  n := negroni.New()
  r := negroni.NewRecovery()
  l := negroni.NewLogger()
  r.PrintStack = false
  n.Use(r)
  n.Use(l)
  s := negroni.NewStatic(http.Dir("public"))
  n.Use(s)
  n.UseHandler(mux)
  port := ":" + os.Getenv("PORT")
  if port == ":" {
    port = ":80"
  }
  addr := os.Getenv("SERVER_ADDR")
  http.ListenAndServe( addr + port, n)
}
