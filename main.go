package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getRegExHandler(data map[string]string) func(w http.ResponseWriter, r *http.Request) {
	// Compile regexp to match path during start and save in a closure
	re := regexp.MustCompile(data["match"])

	return func(w http.ResponseWriter, r *http.Request) {
		var destination = re.ReplaceAllString(r.URL.Path, data["target"])
		http.Redirect(w, r, destination, 301)
	}
}

func getWebhookHandler(data map[string]string) func(w http.ResponseWriter, r *http.Request) {

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(data["endpoint"])
		request, err := client.Get(data["endpoint"])
		check(err)
		destination, _ := request.Location()
		http.Redirect(w, r, destination.String(), 301)
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
	bucketNamePtr := flag.String("bucket-name", "", "Name of the S3 bucket holding the configuration")
	bucketRegionPtr := flag.String("bucket-region", "eu-central-1", "Region of the S3 bucket")
	logLevelPtr := flag.String("logging", "ERROR", "Define the log level")
	flag.Parse()

	if *bucketNamePtr == "" {
		log.Println("-bucket-name is mising")
		flag.PrintDefaults()
		os.Exit(1)
	}

	level, err := log.ParseLevel(*logLevelPtr)
	if err != nil {
		log.Error("Defined invalid loglevel. Defaulting to ERROR")
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(level)
	}

	c := loadConfig(*bucketNamePtr, *bucketRegionPtr)

	if len(c.Entrys) == 0 {
		log.Fatal("Config file doesn't contain any rules")
	}

	r := buildRouter(c)

	err = updateTraefikConfig(c)
	if err != nil {
		log.Fatal("could not write config to consul: ", err)
	}

	http.Handle("/", r)
	panic(http.ListenAndServe(":9090", nil))
}
