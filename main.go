package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type RegexData struct {
	Match  string `json:"match"`
	Target string `json:"target"`
}
type WebhookData struct {
	Endpoint string `json:"endpoint"`
}

type Entry struct {
	Name   string            `json:"name"`
	Domain string            `json:"domain"`
	Type   string            `json:"type"`
	Path   string            `json:"path"`
	Data   map[string]string `json:"data"`
}

type Config struct {
	Entrys []Entry `json:"entrys"`
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
	flag.Parse()

	if *bucketNamePtr == "" {
		log.Println("-bucket-name is mising")
		flag.PrintDefaults()
		os.Exit(1)
	}

	tmpFile, err := ioutil.TempFile("", "config")
	if err != nil {
		log.Fatal("Couldn't create temp file: ", err)
	}

	defer os.Remove(tmpFile.Name())

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(*bucketRegionPtr)},
	)

	downloader := s3manager.NewDownloader(sess)

	_, err = downloader.Download(tmpFile,
		&s3.GetObjectInput{
			Bucket: aws.String(*bucketNamePtr),
			Key:    aws.String("config.yml"),
		})

	if err != nil {
		log.Fatalf("Unable to download configuration from S3: %s", err)
	}

	var configData []byte
	configData, err = ioutil.ReadAll(tmpFile)
	if err != nil {
		log.Fatalf("Couldn't read downloaded config file: %s", err)
	}
	c := &Config{}
	err = yaml.Unmarshal(configData, &c)
	check(err)

	fmt.Println(tmpFile.Name())

	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	if len(c.Entrys) == 0 {
		log.Fatal("Config file doesn't contain any rules")
	}

	for _, e := range c.Entrys {
		fmt.Printf("> Loading rule: %q\n", e.Name)
		s := r.Host(e.Domain).Subrouter()
		switch e.Type {
		case "regex":
			s.PathPrefix(e.Path).HandlerFunc(getRegExHandler(e.Data)).Name(e.Name)
		case "webhook":
			s.PathPrefix(e.Path).HandlerFunc(getWebhookHandler(e.Data)).Name(e.Name)
		}
	}

	http.Handle("/", r)
	panic(http.ListenAndServe(":9090", nil))
}
