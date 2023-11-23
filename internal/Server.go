package internal

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/gabriel-vasile/mimetype"
	"github.com/hyperjumptech/jiffy"
	"github.com/newm4n/dokku-home/configuration"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

var (
	//go:embed static/*
	staticFiles embed.FS
)

func configureLogging() {
	lLevel := configuration.Get("server.log.level")
	fmt.Println("Setting log level to ", lLevel)
	switch strings.ToUpper(lLevel) {
	default:
		fmt.Println("Unknown level [", lLevel, "]. Log level set to ERROR")
		log.SetLevel(log.ErrorLevel)
	case "TRACE":
		log.SetLevel(log.TraceLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	}
}

// Start this server
func Start() {
	configureLogging()
	log.Infof("Starting Server")
	startTime := time.Now()

	var wait time.Duration

	graceShut, err := jiffy.DurationOf(configuration.Get("server.timeout.graceshut"))
	if err != nil {
		panic(err)
	}
	WriteTimeout, err := jiffy.DurationOf(configuration.Get("server.timeout.write"))
	if err != nil {
		panic(err)
	}
	ReadTimeout, err := jiffy.DurationOf(configuration.Get("server.timeout.read"))
	if err != nil {
		panic(err)
	}
	IdleTimeout, err := jiffy.DurationOf(configuration.Get("server.timeout.idle"))
	if err != nil {
		panic(err)
	}

	wait = graceShut

	address := fmt.Sprintf("%s:%s", configuration.Get("server.host"), configuration.Get("server.port"))
	log.Info("Server binding to ", address)

	srv := &http.Server{
		Addr: address,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: WriteTimeout,
		ReadTimeout:  ReadTimeout,
		IdleTimeout:  IdleTimeout,
		// Handler:      Router, // Pass our instance of gorilla/mux in.
		Handler: &StaticProcessor{},
	}
	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	dur := time.Now().Sub(startTime)
	durDesc := jiffy.DescribeDuration(dur, jiffy.NewWant())
	log.Infof("Shutting down. This Hansip been protecting the world for %s", durDesc)
	os.Exit(0)
}

func StaticPath(response http.ResponseWriter, request *http.Request) {

}

type StaticProcessor struct {
}

func WriteResponse(response http.ResponseWriter, status int, headers map[string][]string, body []byte) {
	if headers != nil {
		for hkey, hvarr := range headers {
			for _, hval := range hvarr {
				response.Header().Add(hkey, hval)
			}
		}
	}
	response.WriteHeader(status)
	if body != nil {
		response.Write(body)
	}
}

func (proc *StaticProcessor) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if len(request.URL.Path) < 4 || request.URL == nil {
		log.Warnf("Path redirect")
		WriteResponse(response, http.StatusMovedPermanently, map[string][]string{"Location": {"/index.html"}}, nil)
		return
	} else if request.URL.Path[:4] == "/api" {
		endPoints := configuration.GetEndPoint()
		for _, ep := range endPoints {
			//log.Warnf("ep.URLPathPrefix = %s & request.URL.Path = %s", ep.URLPathPrefix, request.URL.Path)
			targetURL := GetActualPath(request.URL.String(), ep.PathPrefix, ep.URLHost, ep.URLPathPrefix)
			if len(targetURL) > 0 {
				//actualPath := targetURL[len(ep.URLHost):]
				var nRequest *http.Request
				var err error
				if request.Body == nil {
					nRequest, err = http.NewRequest(request.Method, targetURL, nil)
					if err != nil {
						WriteResponse(response, http.StatusMovedPermanently,
							map[string][]string{"Content-Type": {"text/plain"}}, []byte(fmt.Sprintf("got error \"%s\"", err.Error())))
						return
					}
				} else {
					requestBody, err := io.ReadAll(request.Body)
					if err != nil {
						WriteResponse(response, http.StatusInternalServerError,
							map[string][]string{"Content-Type": {"text/plain"}}, []byte(fmt.Sprintf("got error \"%s\"", err.Error())))
						return
					}
					nRequest, err = http.NewRequest(request.Method, targetURL, bytes.NewReader(requestBody))
					if err != nil {
						WriteResponse(response, http.StatusInternalServerError,
							map[string][]string{"Content-Type": {"text/plain"}}, []byte(fmt.Sprintf("got error \"%s\"", err.Error())))
						return
					}
				}

				for hk, hv := range request.Header {
					for _, he := range hv {
						nRequest.Header.Add(hk, he)
					}
				}

				client := http.DefaultClient
				client.Timeout = 2 * time.Minute
				resp, err := client.Do(nRequest)

				if err == nil {
					for rhk, khva := range resp.Header {
						for _, ha := range khva {
							response.Header().Add(rhk, ha)
						}
					}
					response.WriteHeader(resp.StatusCode)
					if resp.Body != nil {
						respBody, err := io.ReadAll(resp.Body)
						if err == nil {
							response.Write(respBody)
						} else {
							log.Errorf("can not write gateway response body. got %s", err.Error())
						}
					}
					return
				}
				if errors.Is(err, http.ErrHandlerTimeout) {
					log.Errorf("gateway error. got %s", err.Error())
					WriteResponse(response, http.StatusGatewayTimeout,
						map[string][]string{"Content-Type": {"text/plain"}}, []byte(fmt.Sprintf("gateway timeout error \"%s\"", err.Error())))
					return
				}
				log.Errorf("gateway error. got %s", err.Error())
				WriteResponse(response, http.StatusBadGateway,
					map[string][]string{"Content-Type": {"text/plain"}}, []byte(fmt.Sprintf("bad gateway error \"%s\"", err.Error())))
				return
			}
			log.Errorf("not found, is shorter %s than %s", ep.PathPrefix, request.URL.Path)
			WriteResponse(response, http.StatusNotFound,
				map[string][]string{"Content-Type": {"text/plain"}}, []byte(fmt.Sprintf("notfound error \"%s\"", request.URL.Path)))
			return
		}
		log.Errorf("not found, %s vs empty range.", request.URL.Path)
		WriteResponse(response, http.StatusNotFound,
			map[string][]string{"Content-Type": {"text/plain"}}, []byte(fmt.Sprintf("notfound error \"%s\"", request.URL.Path)))
		return
	} else {
		log.Warnf("Path %s", request.URL.Path)
		nPath := "static" + request.URL.Path
		if request.Method != http.MethodGet {
			response.WriteHeader(http.StatusMethodNotAllowed)
			log.Errorf("%s for %s : %d. only accept GET", request.Method, nPath, http.StatusMethodNotAllowed)
			return
		}
		fil, err := staticFiles.Open(nPath)
		if err != nil {
			WriteResponse(response, http.StatusNotFound,
				map[string][]string{"Content-Type": {"text/html"}}, []byte(fmt.Sprintf("<html><head><title>Not Found</title></head><body>"+
					"%s</body></html>", nPath)))
			log.Errorf("%s %s: %d", request.Method, nPath, http.StatusNotFound)
			return
		}
		finf, err := fil.Stat()
		if err != nil {
			WriteResponse(response, http.StatusNotFound,
				map[string][]string{"Content-Type": {"text/html"}}, []byte(fmt.Sprintf("<html><head><title>Not Found</title></head><body>"+
					"%s</body></html>", nPath)))
			log.Errorf("%s %s: %d", request.Method, nPath, http.StatusNotFound)
			return
		}
		if finf.IsDir() {
			WriteResponse(response, http.StatusNotFound,
				map[string][]string{"Content-Type": {"text/html"}}, []byte(fmt.Sprintf("<html><head><title>Not Found</title></head><body>"+
					"can not open dir : %s</body></html>", nPath)))
			log.Errorf("%s %s is an unhandled dir: %d", request.Method, nPath, http.StatusNotFound)
			return
		}
		content, err := io.ReadAll(fil)
		if err != nil {
			WriteResponse(response, http.StatusInternalServerError,
				map[string][]string{"Content-Type": {"text/html"}}, []byte(fmt.Sprintf("<html><head><title>Internal Server Error</title></head><body>"+
					" %s</body></html>", finf.Name())))
			log.Errorf("%s %s: %d", request.Method, nPath, http.StatusInternalServerError)
			return
		}
		mtype := mimetype.Detect(content)
		log.Debugf("%s %s: %d %d bytes", request.Method, nPath, http.StatusOK, len(content))
		WriteResponse(response, http.StatusOK,
			map[string][]string{"Content-Type": {mtype.String()}}, content)
	}
}

func GetActualPath(Actual, PathPrefix, URLHost, URLPathPrefix string) string {
	if Actual == PathPrefix {
		return fmt.Sprintf("%s%s", URLHost, URLPathPrefix)
	}
	if len(Actual) > len(PathPrefix) {
		extension := Actual[len(PathPrefix):]
		if URLPathPrefix == "/" {
			return fmt.Sprintf("%s%s", URLHost, extension)
		}
		return fmt.Sprintf("%s%s%s", URLHost, URLPathPrefix, extension)
	}
	return ""
}
