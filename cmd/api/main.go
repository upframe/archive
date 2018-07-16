package main

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"runtime"
	"strconv"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/logpacker/PayPal-Go-SDK"
	"github.com/upframe/api"
	h "github.com/upframe/api/http"
	"github.com/upframe/api/smtp"
)

var (
	configPath  string
	debug       bool
	development bool
)

type config struct {
	Errors           string
	Port             int
	Templates        string
	ApplyFile        string
	NewsletterFile   string
	ParticipantsFile string
	EmailsFile       string
	SMTP             struct {
		User     string
		Password string
		Host     string
		Port     string
	}
	PayPal struct {
		Client string
		Secret string
	}
}

func init() {
	flag.StringVar(&configPath, "config", "config.json", "Path to the configuration file")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.BoolVar(&development, "development", false, "Development")
}

func main() {
	flag.Parse()

	// Execute with all of the CPUs available
	runtime.GOMAXPROCS(runtime.NumCPU())

	api.Password = os.Getenv("API_SECRET")
	if api.Password == "" {
		panic(errors.New("API_SECRET must not be empty"))
	}

	api.JWTSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(api.JWTSecret) == 0 {
		panic(errors.New("JWT_SECRET must not be empty"))
	}

	f := &config{}

	configFile, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&f)
	if err != nil {
		panic(err)
	}

	email := smtp.InitSMTP(f.SMTP.User, f.SMTP.Password, f.SMTP.Host, f.SMTP.Port)
	email.TemplatesPath = f.Templates
	email.FromDefaultEmail = "noreply@upframe.co"

	c := &api.Config{
		Port:      strconv.Itoa(f.Port),
		Templates: f.Templates,
		Services: &api.Services{
			Email: email,
		},
	}

	if f.Errors == "" {
		f.Errors = "stdout"
	}

	c.Logger = log.New()

	if debug {
		c.Logger.Level = log.DebugLevel
	}

	if f.Errors == "stdout" {
		c.Logger.Out = os.Stdout
	} else {
		var file *os.File
		file, err = os.OpenFile(f.Errors, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			panic(err)
		}

		defer file.Close()
		c.Logger.Out = file
	}

	c.Apply = &api.File{
		Location: f.ApplyFile,
		Mu:       &sync.Mutex{},
	}

	c.Newsletter = &api.File{
		Location: f.NewsletterFile,
		Mu:       &sync.Mutex{},
	}

	c.Participants = &api.File{
		Location: f.ParticipantsFile,
		Mu:       &sync.Mutex{},
	}

	c.Emails = &api.File{
		Location: f.EmailsFile,
		Mu:       &sync.Mutex{},
	}

	link := paypalsdk.APIBaseLive
	if development {
		link = paypalsdk.APIBaseSandBox
	}

	paypal, err := paypalsdk.NewClient(f.PayPal.Client, f.PayPal.Secret, link)
	if err != nil {
		panic(err)
	}

	c.PayPal = paypal
	h.Serve(c)
}
