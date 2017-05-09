package main

import (
  "encoding/json"
  "errors"
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
  n.UseHandler(mux)
  s := negroni.NewStatic(http.Dir("public"))
  n.Use(s)
  port := ":" + os.Getenv("PORT")
  if port == ":" {
    port = ":80"
  }
  addr := os.Getenv("SERVER_ADDR")
  http.ListenAndServe( addr + port, n )
}

func handleRender(mux *http.ServeMux, rend *render.Render) {
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    path, templateName, err := getPathAndTemplate(req)
    if err != nil {
      return
    }
    data, err := loadData(templateName + ".json", path)

    if err != nil {
      rend.HTML(w, http.StatusServiceUnavailable, "default/dataUnavailable", "")
    } else if false == Exists(templateName, ".tmpl", "templates") {
      rend.HTML(w, http.StatusServiceUnavailable, "default/templateUnavailable", "")
    } else {
      rend.HTML(w, http.StatusOK, templateName, data)
    }
  })
}

func getPathAndTemplate(req *http.Request) (string, string, error) {
  path := getPath(req)
  templateName, err := getTemplate(req)
  return path, templateName, err
}

func getPath(req *http.Request) string {
  proto := "http://"
	sslProxyHeader := req.Header.Get( "X-Forwarded-Proto" )
	if sslProxyHeader == "https" {
		proto = "https://"
	}

	host := req.Header.Get( "Host" )
	if host == "" {
		host = req.Host
	}
	return proto + host + "/"
}

func getTemplate(req *http.Request) (string, error) {
  var err error
	templateName := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")
	if templateName == "" {
		templateName = "index"
	} else if Exists( req.URL.Path, "", "public" ) == true {
    err = errors.New("Not a template")
  }
  return templateName, err
}

func Exists(name string, extension string, folder string) bool {
  filename := folder + "/" + name + extension
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
