package main

import (
  "net/http"

  "github.com/urfave/negroni"
  "github.com/unrolled/render"
)

func main() {
  rend := render.New(render.Options{IsDevelopment: true})
  mux := http.NewServeMux()

  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    rend.HTML(w, http.StatusOK, "index", "")
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
  http.ListenAndServe(":3000", n)
}
