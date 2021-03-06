package util

import (
	"errors"
	"os"
	"strconv"
)

type Application struct {
	Endpoint struct {
		Username string
		Password string
		URL      string
	}
	SMB struct {
		Servername string
		Sharename  string
		Username   string
		Password   string
		Domain     string
	}
	Interval struct {
		Seconds int
	}
}

var (
	err error        = nil
	App *Application = &Application{}
)

// LoadEnv loads OS environment variables
func LoadEnv() error {

	if App.Endpoint.Username = os.Getenv("ENDPOINT_USERNAME"); App.Endpoint.Username == "" {
		err = errors.New("ENDPOINT_USERNAME is not set")
		return err
	}

	if App.Endpoint.Password = os.Getenv("ENDPOINT_PASSWORD"); App.Endpoint.Password == "" {
		err = errors.New("ENDPOINT_PASSWORD is not set")
		return err
	}

	if App.Endpoint.URL = os.Getenv("ENDPOINT_URL"); App.Endpoint.URL == "" {
		err = errors.New("ENDPOINT_URL is not set")
		return err
	}

	if App.SMB.Servername = os.Getenv("SMB_SERVERNAME"); App.SMB.Servername == "" {
		err = errors.New("SMB_SERVERNAME is not set")
		return err
	}

	if App.SMB.Sharename = os.Getenv("SMB_SHARENAME"); App.SMB.Sharename == "" {
		err = errors.New("SMB_SHARENAME is not set")
		return err
	}

	if App.SMB.Username = os.Getenv("SMB_USERNAME"); App.SMB.Username == "" {
		err = errors.New("SMB_USERNAME is not set")
		return err
	}

	if App.SMB.Password = os.Getenv("SMB_PASSWORD"); App.SMB.Password == "" {
		err = errors.New("SMB_PASSWORD is not set")
		return err
	}

	if App.SMB.Domain = os.Getenv("SMB_DOMAIN"); App.SMB.Domain == "" {
		err = errors.New("SMB_DOMAIN is not set")
		return err
	}

	// get string to int for Interval
	if App.Interval.Seconds, err = strconv.Atoi(os.Getenv("INTERVAL_SECONDS")); App.Interval.Seconds == 0 || err != nil {
		ErrorLogger.Println("INTERVAL_SECONDS is not set or is not a number, using default value of 60")
		App.Interval.Seconds = 60
		err = nil // reset error
	}

	return err
}
