package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/upframe/api"
)

type response struct {
	ID      string `json:"ID,omitempty"`
	Code    int
	Content interface{}
	Error   error `json:"-"`
}

type handler func(w http.ResponseWriter, r *http.Request, c *api.Config) (int, interface{}, error)

func i(h handler, c *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			code    int
			err     error
			content interface{}
		)

		defer func() {
			if code == 0 && err == nil {
				return
			}

			msg := &response{Code: code}

			if err != nil {
				msg.Content = err.Error()
			} else if content != nil {
				msg.Content = content
			} else {
				msg.Content = http.StatusText(code)
			}

			if code >= 400 {
				t := time.Now()
				msg.ID = t.Format("20060102150405")
			}

			if code >= 400 && err != nil {
				c.Logger.Error(err)
			}

			if code != 0 {
				w.WriteHeader(code)
			}

			data, e := json.MarshalIndent(msg, "", "\t")
			if e != nil {
				c.Logger.Error(e)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
			return
		}()

		code, content, err = h(w, r, c)
	}
}

func s(h handler, c *api.Config) http.HandlerFunc {
	return i(func(w http.ResponseWriter, r *http.Request, c *api.Config) (int, interface{}, error) {
		auth := strings.Split(r.Header.Get("Authorization"), " ")
		if len(auth) != 2 {
			return http.StatusUnauthorized, nil, nil
		}

		fn := func(token *jwt.Token) (interface{}, error) {
			// Make sure token's signature wasn't changed
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected siging method")
			}
			return api.JWTSecret, nil
		}

		// Return a Token using the cookie
		token, err := jwt.ParseWithClaims(auth[1], &jwt.StandardClaims{}, fn)
		if err != nil {
			return http.StatusInternalServerError, nil, err
		}

		if !token.Valid {
			return http.StatusUnauthorized, nil, nil
		}

		return h(w, r, c)
	}, c)
}
