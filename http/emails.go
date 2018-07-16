package http

import (
	"bytes"
	"io"
	"net/http"
	"net/mail"
	"strings"

	"github.com/tealeg/xlsx"
	"github.com/upframe/api"
)

const mainTemplate = `<!DOCTYPE html >
<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
    <title>{{ .Subject }}</title>
    <link href="https://fonts.googleapis.com/css?family=Open+Sans:400,600" rel="stylesheet">
</head>
<body style="background-color: #fff;margin: 0;padding: 0;">
  <div style="margin: 2em auto; width: 95%; max-width: 35em; font-family: 'Open Sans',helvetica neue,helvetica,arial,sans-serif; font-size: 17px;     color: #686868; line-height: 1.4;">
    <p style="text-align:center;"><img width="90px" src="https://next.upframe.co/img/email/mark.png"></p>
		#CONTENT#
  </div>
</body>
</html>
`

// TODO: a lot messy... fix that
func emails(w http.ResponseWriter, r *http.Request, c *api.Config) (int, interface{}, error) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	to, _, err := r.FormFile("to")
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	defer to.Close()

	body := r.FormValue("body")
	subject := r.FormValue("subject")
	fromName := r.FormValue("from_name")
	fromEmail := r.FormValue("from_email")
	if body == "" || subject == "" || fromName == "" || fromEmail == "" {
		return http.StatusBadRequest, nil, nil
	}

	template := strings.Replace(mainTemplate, "#CONTENT#", body, 1)

	buf := &bytes.Buffer{}
	io.Copy(buf, to)

	excel, err := xlsx.OpenBinary(buf.Bytes())
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	status, log := http.StatusOK, ""

	sheet := excel.Sheets[0]
	for i := range sheet.Rows {
		name := sheet.Rows[i].Cells[0].Value
		email := sheet.Rows[i].Cells[1].Value

		if name == "" || email == "" {
			continue
		}

		e := &api.Email{
			From: &mail.Address{
				Name:    fromName,
				Address: fromEmail,
			},
			To: &mail.Address{
				Name:    name,
				Address: email,
			},
			Subject: subject,
		}

		log += "Sending to " + email

		err := c.Services.Email.UseTemplate(e,
			map[string]string{
				"Name":    name,
				"Subject": subject,
			}, template)

		if err != nil {
			status = 500
			log += " ...Error: " + err.Error() + "\n"
			break
		}

		err = c.Services.Email.Send(e)
		if err != nil {
			status = 500
			log += " ...Error: " + err.Error() + "\n"
			break
		}

		log += " ...Sent\n"
	}

	w.WriteHeader(status)
	w.Write([]byte(log))
	return 0, nil, nil
}
