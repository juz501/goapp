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
  "github.com/juz501/go-logger-middleware"
)

func main() {
  logfilename := os.Getenv("GO_LOGFILE")
  if logfilename == "" {
    logfilename = "logs/goapp.log"
  }
  errorLog, err := os.OpenFile(logfilename, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  if err != nil {
    log.Fatal("error writing to log: " + logfilename)
  }
  defer errorLog.Close()

  rend := render.New(render.Options{IsDevelopment: true})
  mux := http.NewServeMux()


  n := negroni.New()
  r := negroni.NewRecovery()
  l := juz501.NewLoggerWithStream( errorLog )
  r.Logger = l
  handleRender(mux, rend, l)
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
  l.Println("Starting Goapp Service")
  l.Println("----------------------")
  http.ListenAndServe( addr + port, n )
}

func handleRender(mux *http.ServeMux, rend *render.Render, logger juz501.ALogger) {
  mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
    baseURI, templateName, err := getBaseURIAndTemplate(req, logger)
    if err != nil {
      return
    }
    data, err := loadData(templateName + ".json", baseURI, logger)

    if err != nil {
      rend.HTML(w, http.StatusServiceUnavailable, "default/dataUnavailable", "")
    } else if false == Exists(templateName, ".tmpl", "templates") {
      rend.HTML(w, http.StatusServiceUnavailable, "default/templateUnavailable", "")
    } else {
      rend.HTML(w, http.StatusOK, templateName, data)
    }
  })
}

func getBaseURIAndTemplate(req *http.Request, logger juz501.ALogger) (string, string, error) {
  proto, host, _ := getRequestVars(req, logger)
  baseURI := proto + "://" + host + "/" 
  templateName, err := getTemplate(req, baseURI, logger)
  return baseURI, templateName, err
}

func getRequestVars(req *http.Request, logger juz501.ALogger) (string, string, string) {
  proto := req.URL.Scheme
  if proto == "" {
    proto = "http"
  } 
  forwardedProto := req.Header.Get( "X-Forwarded-Proto" )
  if forwardedProto != "" {
    proto = forwardedProto
  }
  
  forwardedHost := req.Header.Get( "X-Forwarded-Host" )
  host := req.Host
  if forwardedHost != "" {
    host = forwardedHost
  }
  forwardedPath := req.Header.Get( "X-Forwarded-Path" )
  path := req.URL.Path
  if forwardedPath != "" {
    path = forwardedPath
  }
	return proto, host, path 
}

func getTemplate(req *http.Request, baseURI string, logger juz501.ALogger) (string, error) {
  var err error
	templateName := strings.TrimSuffix(strings.TrimPrefix(req.RequestURI, "/"), "/")
	if templateName == "" {
		templateName = "index"
	} else if Exists( templateName, "", "public" ) == true {
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


func loadData(filename string, basepath string, logger juz501.ALogger) (interface{}, error) {
  var raw []byte
  raw, err := ioutil.ReadFile("data/" + filename)
  if err != nil {
    return raw, err
  }
  var data map[string]interface{}
  err = json.Unmarshal(raw, &data)
  data["BasePath"] = basepath 

  return data, err
}
