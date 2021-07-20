package main

import (
	"flag"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getRegExHandler(matcher string, target string) func(w http.ResponseWriter, r *http.Request) {
	// Compile regexp to match path during start and save in a closure
	re := regexp.MustCompile(matcher)

	return func(w http.ResponseWriter, r *http.Request) {
		location := re.ReplaceAllString(r.URL.Path, target)
		http.Redirect(w, r, location, http.StatusMovedPermanently)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x := mux.CurrentRoute(r)
		log.Printf("[%s] %s%s\n", x.GetName(), r.Host, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func main() {
	withCaddyPtr := flag.Bool("with-caddy", false, "Automatically configure a Caddy server")
	portPtr := flag.String("port", "9090", "Port to listen on")
	bucketNamePtr := flag.String("bucket-name", "", "Name of the S3 bucket holding the configuration")
	bucketRegionPtr := flag.String("bucket-region", "eu-central-1", "Region of the S3 bucket")
	logLevelPtr := flag.String("logging", "WARN", "Define the log level")

	flag.Parse()

	level, err := log.ParseLevel(*logLevelPtr)
	if err != nil {
		log.Error("Defined invalid loglevel. Defaulting to ERROR")
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(level)
	}

	var c *Config
	if *bucketNamePtr != "" {
		c = loadConfigFromS3(*bucketNamePtr, *bucketRegionPtr)
	} else {
		c = loadLocalConfig("config.yml")
	}

	if len(*c) == 0 {
		log.Fatal("Config file doesn't contain any rules")
	}

	if *withCaddyPtr {
		configureCaddy(c)
	}

	r := buildRouter(c)

	log.Infof("Listening on :%s", *portPtr)
	panic(http.ListenAndServe(fmt.Sprintf(":%s", *portPtr), r))
}
