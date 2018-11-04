package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Route struct {
	Path   string `json:"path"`
	Match  string `json:"match"`
	Target string `json:"target"`
}

type Entry struct {
	Name   string  `json:"name"`
	Domain string  `json:"domain"`
	Routes []Route `json:"routes"`
}

type Config struct {
	Entrys []Entry `json:"entrys"`
}

func getRedirectFunc(route Route) func(w http.ResponseWriter, r *http.Request) {
	// Compile regexp to match path during start and save in a closure
	re := regexp.MustCompile(route.Match)

	return func(w http.ResponseWriter, r *http.Request) {
		var destination = re.ReplaceAllString(r.URL.Path, route.Target)
		http.Redirect(w, r, destination, 301)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s%s\n", r.Host, r.RequestURI)
		x := mux.CurrentRoute(r)
		log.Println(x.GetName())
		next.ServeHTTP(w, r)
	})
}

func main() {
	dat, err := ioutil.ReadFile("config.yml")
	check(err)
	c := &Config{}
	err = yaml.Unmarshal(dat, &c)
	check(err)

	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	for _, e := range c.Entrys {
		fmt.Printf("> Loading rule: %q\n", e.Name)
		s := r.Host(e.Domain).Subrouter()
		for _, r := range e.Routes {
			s.PathPrefix(r.Path).HandlerFunc(getRedirectFunc(r)).Name(e.Name)
		}
	}

	http.Handle("/", r)

	panic(http.ListenAndServe(":9090", nil))
}
