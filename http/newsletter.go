package http

import (
	"encoding/csv"
	"net/http"
	"os"
	"time"

	"github.com/upframe/api"
)

func newsletter(w http.ResponseWriter, r *http.Request, c *api.Config) (int, interface{}, error) {
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, nil, nil
	}

	email := r.FormValue("email")
	if email == "" {
		return http.StatusBadRequest, nil, nil
	}

	go saveNewsletterEmail(c, email, false)
	return http.StatusOK, nil, nil
}

func saveNewsletterEmail(c *api.Config, email string, retry bool) {
	if retry {
		time.Sleep(time.Minute * 5)
	}

	c.Newsletter.Mu.Lock()
	defer c.Newsletter.Mu.Unlock()

	f, err := os.OpenFile(c.Newsletter.Location, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer f.Close()

	if err != nil {
		c.Logger.Warnf("Error writing '%v' on %v: %v. Retrying in 5 minutes.", email, c.Newsletter.Location, err)
		go saveNewsletterEmail(c, email, true)
		return
	}

	w := csv.NewWriter(f)
	err = w.Write([]string{email})

	if err != nil {
		c.Logger.Warnf("Error writing '%v' on %v: %v. Retrying in 5 minutes.", email, c.Newsletter.Location, err)
		go saveNewsletterEmail(c, email, true)
		return
	}

	if retry {
		c.Logger.Infof("'%v' written to %v.", email, c.Newsletter.Location)
	}

	w.Flush()
}
