package api

import (
	"io/ioutil"
	"log"

	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func setup() *router {
	SetAppConfig("../testdata/app-config.json")

	b, err := ioutil.ReadFile("../testdata/secrets.json")
	if err != nil {
		log.Fatal(err)
	}
	SetAppSecrets(b)

	return &router{}
}

func TestConfigure(t *testing.T) {
	router := mux.NewRouter()
	p := newProducer()

	configureRouter(router, p)

	count := 0
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		_, err := route.GetPathTemplate()
		if err != nil {
			assert.Fail(t, err.Error())
		}
		count++
		return nil
	})

	assert.Equal(t, 4, count)
}

func TestHealthSuccess(t *testing.T) {
	h := setup()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	resp := w.Result()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestCheckStatusPingSuccess(t *testing.T) {
	r := setup()
	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	r.Ping(w, req)

	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
