package main

import (
  "encoding/json"
  "errors"
  "io/ioutil"
  "log"
  "net/http"
  "net/http/httputil"
  "os"
  "strings"

  "github.com/urfave/negroni"
  "github.com/unrolled/render"
  "github.com/juz501/go-logger-middleware"
)

func main() {
  logfilename := os.Getenv("GO_LOGFILE")
  if logfilename == "" {
    logfilename = "/var/log/gologs/goapp.log"
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
    requestDump, err := httputil.DumpRequest(req, true)
    if err != nil {
      logger.Println(err)
    } else {
      logger.Println( string( requestDump ) ) 
    }
    path, templateName, err := getPathAndTemplate(req, logger)
    if err != nil {
      return
    }
    data, err := loadData(templateName + ".json", path, logger)

    if err != nil {
      rend.HTML(w, http.StatusServiceUnavailable, "default/dataUnavailable", "")
    } else if false == Exists(templateName, ".tmpl", "templates") {
      rend.HTML(w, http.StatusServiceUnavailable, "default/templateUnavailable", "")
    } else {
      rend.HTML(w, http.StatusOK, templateName, data)
    }
  })
}

func getPathAndTemplate(req *http.Request, logger juz501.ALogger) (string, string, error) {
  path := getPath(req, logger)
  templateName, err := getTemplate(req)
  return path, templateName, err
}

func getPath(req *http.Request, logger juz501.ALogger) string {
  proto := "http://"
	sslProxyHeader := req.Header.Get( "X-Forwarded-Proto" )
	if sslProxyHeader == "https" {
		proto = "https://"
	}

	host := req.Header.Get( "Host" )
	if host == "" {
    logger.Println( "test" )
    logger.Println( req.Host )
    logger.Println( host )
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


func loadData(filename string, path string, logger juz501.ALogger) (interface{}, error) {
  var raw []byte
  raw, err := ioutil.ReadFile("data/" + filename)
  if err != nil {
    logger.Println( err )
    return raw, err
  }
  var data map[string]interface{}
  err = json.Unmarshal(raw, &data)
  data["Path"] = path

  return data, err
}
