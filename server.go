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
  "github.com/juz501/go_logger_middleware"
  "github.com/juz501/go_static_middleware"
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

  rend := render.New(render.Options{IsDevelopment: true, Directory: "theme/templates" })
  mux := http.NewServeMux()
  n := negroni.New()
  l := go_logger_middleware.NewLoggerWithStream( errorLog )
  r := negroni.NewRecovery()
  r.Logger = l
  r.PrintStack = false
  baseRoute := os.Getenv("GOBASEROUTE")

  handleRender(mux, rend, l, baseRoute)
  s := go_static_middleware.NewStatic(http.Dir("public"))
  if baseRoute != "" {
    s.Prefix = "/" + baseRoute
  }

  n.Use(r)
  n.Use(l)
  n.UseHandler(mux)

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

func handleRender(mux *http.ServeMux, rend *render.Render, logger go_logger_middleware.ALogger, base string) {
  mux.HandleFunc( "/", func(w http.ResponseWriter, req *http.Request) {
    logger.Println( "start" )
    baseURI, prefix := getBaseURI(req, base, logger)

    templateName, hasTemplate, isPublicFile := getTemplate(req, prefix, logger)
    if isPublicFile == true {
      logger.Println( "public file" )
      return
    }
    if false == hasTemplate {
      rend.HTML(w, http.StatusServiceUnavailable, "default/templateUnavailable", "")
      return
    }
    data := loadData(req, templateName + ".json", baseURI, prefix, logger)

    rend.HTML(w, http.StatusOK, templateName, data)
  })
}

func getBaseURI(req *http.Request, baseRoute string, logger go_logger_middleware.ALogger) (string, string) {
  scheme, host, prefix, _ := getRequestVars(req, baseRoute, logger)
  baseURI := scheme + "://" + host
  if prefix != "" {
    baseURI = baseURI + "/" + prefix
  }
  return baseURI, prefix
}

func getRequestVars(req *http.Request, baseRoute string, logger go_logger_middleware.ALogger) (string, string, string, string) {
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

func getTemplate(req *http.Request, baseURI string, logger go_logger_middleware.ALogger) (string, bool, bool) {
  logger.Println( "RequestURI Template: " + req.RequestURI )
  logger.Println( "BaseURI Template: " + baseURI )
	templateName := strings.TrimSuffix(strings.TrimPrefix(req.RequestURI, "/" + baseURI), "/")
  isPublicFile := false
  hasTemplate := false

  if strings.HasPrefix(templateName, "/") {
    templateName = strings.TrimPrefix( templateName, "/" )
  }

  logger.Println( "Template Name: [" + templateName + "]" )
	if templateName == "" {
		templateName = "index"
    hasTemplate = Exists( templateName, ".tmpl", "theme/templates", logger )
    isPublicFile = false
    return templateName, hasTemplate, isPublicFile
	}

  isPublicFile = Exists( templateName, "", "public", logger )
  if false == isPublicFile {
    hasTemplate = Exists( templateName, ".tmpl", "theme/templates", logger )
  }
  return templateName, hasTemplate, isPublicFile
}

func Exists(name string, extension string, folder string, logger go_logger_middleware.ALogger) bool {
  filename := folder + "/" + name + extension
  logger.Println( "file exists?: " + filename )
  _, err := os.Stat( filename )
  if os.IsNotExist(err) {
    logger.Println( "No" )
    return false
  }
  logger.Println( "Yes" )
  return true
}


func loadData(req *http.Request, filename string, baseURI string, prefix string, logger go_logger_middleware.ALogger) interface{} {
  var raw []byte
  var data map[string]interface{}
  logger.Println( "datafile: " + filename )
  raw, err := ioutil.ReadFile("theme/data/" + filename)
  if err != nil {
    data = make( map[string]interface{} )
  } else {
    err = json.Unmarshal(raw, &data)
    if err != nil {
      logger.Println( "Unmarshalling JSON failed" )
    }
  }
  data["BaseURI"] = baseURI + "/"
  basePath := ""
  if prefix != "" {
    basePath = "/" + prefix
  }
  path := strings.TrimPrefix(req.RequestURI, basePath)
  data["CanonicalURI"] = strings.TrimSuffix(baseURI + path, "/")
  return data
}
