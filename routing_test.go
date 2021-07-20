package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

type RedirectCheck struct {
	Host       string
	Path       string
	Target     string
	StatusCode int
}

func executeRequest(router *mux.Router, t *testing.T, host, path, target string, status int) {
	req, _ := http.NewRequest("GET", path, nil)
	req.Host = host

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != status {
		t.Fatalf("Did not receive expected HTTP status code, wanted %d but got %d", status, rr.Code)
	}

	location, err := rr.Result().Location()

	if status == 301 {
		if err != nil {
			t.Fatal("Response did not contain a 'Location' header")
		}

		if location.String() != target {
			t.Errorf("Received wrong target location, wanted %q but got %q", target, location.String())
		}
	}
}

func TestRouting(t *testing.T) {
	c := loadLocalConfig("config_testing.yml")
	r := buildRouter(c)

	checks := []RedirectCheck{
		{
			Host:       "a.example.com",
			Path:       "/hello",
			StatusCode: 301,
			Target:     "https://www.example.com",
		},
		{
			Host:       "b.example.com",
			Path:       "/hello",
			StatusCode: 301,
			Target:     "https://blog.example.com/hello",
		},
		{
			Host:       "b.example.com",
			Path:       "/images/header.png",
			StatusCode: 301,
			Target:     "https://blog.example.com/assets/header.png",
		},
		{
			Host:       "c.example.com",
			Path:       "/hello",
			StatusCode: 404,
		},
	}

	for _, check := range checks {
		executeRequest(r, t, check.Host, check.Path, check.Target, check.StatusCode)
	}

}
