package main

import (
  "encoding/json"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "strings"

  "github.com/urfave/negroni"
  "github.com/unrolled/render"
)

func main() {
  rend := render.New(render.Options{IsDevelopment: true})
  mux := http.NewServeMux()

 
  handleRender(mux, rend)

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
  http.ListenAndServe( addr + port, n )
}

func handleRender(mux *http.ServeMux, rend *render.Render) {
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    templateName := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")
    path := "/goapp/"
    if templateName == "" {
      templateName = "index"
    }
    data, err := loadData(templateName + ".json", path)
    if err != nil {
      rend.HTML(w, http.StatusServiceUnavailable, "default/dataUnavailable", "")
    } else if false == Exists(templateName) {
      rend.HTML(w, http.StatusServiceUnavailable, "default/templateUnavailable", "")
    } else { 
      rend.HTML(w, http.StatusOK, templateName, data)
    }
  })
}

func Exists(name string) bool {
  filename := "templates/" + name + ".tmpl"
  _, err := os.Stat( filename )
  if os.IsNotExist(err) {
    return false
  }
  return true
}


func loadData(filename string, path string) (interface{}, error) {
  var raw []byte
  raw, err := ioutil.ReadFile("data/" + filename)
  if err != nil {
    log.Println( err )
    return raw, err
  }
  var data map[string]interface{}
  err = json.Unmarshal(raw, &data)
  data["Path"] = path

  return data, err
}
