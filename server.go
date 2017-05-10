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
  "github.com/juz501/go-logger-middleware"
)

func main() {
  logfilename := os.Getenv("GO_LOGFILE")
  if logfilename == "" {
    logfilename = "log/goapp.log"
  }
  errorLog, err := os.OpenFile(logfilename, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  if err != nil {
    log.Fatal("error writing to log: " + logfilename)
  }
  defer errorLog.Close()

  rend := render.New(render.Options{IsDevelopment: true})
  mux := http.NewServeMux()
  n := negroni.New()
  l := juz501.NewLoggerWithStream( errorLog )

  r := negroni.NewRecovery()
  r.Logger = l
  r.PrintStack = false
  baseRoute := os.Getenv("GO_BASE_ROUTE")
  if baseRoute == "" {
    baseRoute = "/"
  }
  
  handleRender(mux, rend, l, baseRoute)
  s := negroni.NewStatic(http.Dir("public"))
  s.Prefix = strings.TrimSuffix(baseRoute, "/")

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

func handleRender(mux *http.ServeMux, rend *render.Render, logger juz501.ALogger, baseRoute string) {
  if baseRoute != "/" {
    mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
      
    })
  }
  handleRoute := func(w http.ResponseWriter, req *http.Request) {
    baseURI := getBaseURI(req, baseRoute, logger)

    logger.Println( baseURI )
    templateName, hasTemplate, isPublicFile := getTemplate(req, baseURI, logger)
    if isPublicFile == true {
      logger.Println( "public file" )
      return 
    }
    if false == hasTemplate {
      rend.HTML(w, http.StatusServiceUnavailable, "default/templateUnavailable", "")
      return
    }
    data, err := loadData(templateName + ".json", baseURI, logger)

    if err != nil {
      logger.Println( err )
      rend.HTML(w, http.StatusServiceUnavailable, "default/dataUnavailable", "")
      return
    }
    rend.HTML(w, http.StatusOK, templateName, data)
  }
  mux.HandleFunc(baseRoute, handleRoute)
  if strings.HasSuffix(baseRoute, "/") {
    newBaseRoute := strings.TrimSuffix(baseRoute, "/")
    mux.HandleFunc(newBaseRoute, handleRoute)
  }
}

func getBaseURI(req *http.Request, baseRoute string, logger juz501.ALogger) string {
  _, _, prefix, _ := getRequestVars(req, baseRoute, logger)
  baseURI := prefix
  if false == strings.HasSuffix(baseURI, "/") {
    baseURI = baseURI + "/"
  }
  return baseURI
}

func getRequestVars(req *http.Request, baseRoute string, logger juz501.ALogger) (string, string, string, string) {
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
  forwardedPrefix := req.Header.Get( "X-Forwarded-Prefix" )
  prefix := baseRoute
  if forwardedPrefix != "" {
    prefix = forwardedPrefix
  }
  forwardedPath := req.Header.Get( "X-Forwarded-Path" )
  path := req.URL.Path
  if forwardedPath != "" {
    path = forwardedPath
  }
	return proto, host, prefix, path 
}

func getTemplate(req *http.Request, baseURI string, logger juz501.ALogger) (string, bool, bool) {
	templateName := strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(req.RequestURI, baseURI), "/"), "/")
  isPublicFile := false
  hasTemplate := false

  logger.Println( "RequestURI Template: " + req.RequestURI )
  logger.Println( "BaseURI Template: " + baseURI )
  logger.Println( "Template Name: [" + templateName + "]" )

	if templateName == "" {
		templateName = "index"
    hasTemplate = Exists( templateName, ".tmpl", "templates", logger )
    isPublicFile = false
    return templateName, hasTemplate, isPublicFile
	} 
  isPublicFile = Exists( templateName, "", "public", logger )
  if false == isPublicFile {
    hasTemplate = Exists( templateName, ".tmpl", "templates", logger ) 
  }
  return templateName, hasTemplate, isPublicFile
}

func Exists(name string, extension string, folder string, logger juz501.ALogger) bool {
  filename := folder + "/" + name + extension
  logger.Println( "file exists?: " + filename )
  _, err := os.Stat( filename )
  if os.IsNotExist(err) {
    logger.Println( "No")
    return false 
  }
  logger.Println( "Yes")
  return true
}


func loadData(filename string, baseURI string, logger juz501.ALogger) (interface{}, error) {
  var raw []byte
  logger.Println( "datafile: " + filename )
  raw, err := ioutil.ReadFile("data/" + filename)
  if err != nil {
    return raw, err
  }
  var data map[string]interface{}
  err = json.Unmarshal(raw, &data)
  data["BaseURI"] = baseURI

  return data, err
}
