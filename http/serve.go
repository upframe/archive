package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/upframe/api"
)

// Serve ...
func Serve(c *api.Config) {
	r := mux.NewRouter()

	r.NotFoundHandler = &notFoundHandler{Config: c}

	r.HandleFunc("/tokens/validate", s(placebo, c)).Methods("POST")
	r.HandleFunc("/tokens/get", i(tokensGet, c)).Methods("POST")
	r.HandleFunc("/newsletter", i(newsletter, c)).Methods("POST")
	r.HandleFunc("/emails", s(emails, c)).Methods("POST")

	c.Logger.Infof("Listening on port %s.", c.Port)

	if err := http.ListenAndServe(":"+c.Port, r); err != nil {
		c.Logger.Fatal(err)
	}
}

type notFoundHandler struct {
	Config *api.Config
}

func (h *notFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	i(func(w http.ResponseWriter, r *http.Request, c *api.Config) (int, interface{}, error) {
		return http.StatusNotFound, nil, nil
	}, h.Config)(w, r)
}

func placebo(w http.ResponseWriter, r *http.Request, c *api.Config) (int, interface{}, error) {
	return http.StatusOK, nil, nil
}
