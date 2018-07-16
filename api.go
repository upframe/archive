package api

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/logpacker/PayPal-Go-SDK"
)

var (
	Password  string
	JWTSecret []byte
)

// Config ...
type Config struct {
	Port         string
	Scheme       string
	Templates    string
	Apply        *File
	Newsletter   *File
	Participants *File
	Emails       *File
	Logger       *logrus.Logger
	PayPal       *paypalsdk.Client
	Services     *Services
}

// File ...
type File struct {
	Location string
	Mu       *sync.Mutex
}

// Services ...
type Services struct {
	Email EmailService
}
