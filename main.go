package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hirochachacha/go-smb2"
)

type application struct {
	endpoint struct {
		username string
		password string
		url      string
	}
	smb struct {
		servername string
		sharename  string
		username   string
		password   string
		domain     string
	}
	interval struct {
		seconds int
	}
}

func main() {

	app := new(application)

	app.endpoint.username = os.Getenv("ENDPOINT_USERNAME")
	app.endpoint.password = os.Getenv("ENDPOINT_PASSWORD")
	app.endpoint.url = os.Getenv("ENDPOINT_URL")
	app.smb.servername = os.Getenv("SMB_SERVERNAME")
	app.smb.sharename = os.Getenv("SMB_SHARENAME")
	app.smb.username = os.Getenv("SMB_USERNAME")
	app.smb.password = os.Getenv("SMB_PASSWORD")
	app.smb.domain = os.Getenv("SMB_DOMAIN")

	app.interval.seconds = getInterval(os.Getenv("INTERVAL_SECONDS"))

	checkEnvs(app)

	for {

		err := sendSMBFiles(app, app.smb.servername, app.smb.sharename, app.smb.username, app.smb.password, app.smb.domain)
		if err != nil && !strings.Contains(err.Error(), "EOF") {
			log.Println(err)
		}
		time.Sleep(time.Duration(app.interval.seconds) * time.Second)

	}

}

func checkEnvs(app *application) {
	if app.endpoint.username == "" {
		log.Fatal("endpoint username must be provided")
	}
	if app.endpoint.password == "" {
		log.Fatal("endpoint username must be provided")
	}
	if app.endpoint.url == "" {
		log.Fatal("endpoint username must be provided")
	}

	if app.smb.servername == "" {
		log.Fatal("smb servername must be provided")
	}
	if app.smb.sharename == "" {
		log.Fatal("smb sharename must be provided")
	}
	if app.smb.username == "" {
		log.Fatal("smb username must be provided")
	}
	if app.smb.password == "" {
		log.Fatal("smb password must be provided")
	}
	if app.smb.domain == "" {
		log.Fatal("smb domain must be provided")
	}
}

func getInterval(interval_seconds string) int {
	if interval_seconds == "" {
		return 60
	} else {
		interval, err := strconv.Atoi(os.Getenv("INTERVAL_SECONDS"))
		if err != nil {
			log.Fatal("interval seconds must be a number")
		}
		if interval < 30 {
			log.Println("the interval must at least be 30s -> automatically set to this value now")
			interval = 30
		}
		return interval
	}
}

func sendSMBFiles(app *application, servername string, sharename string, username string, password string, domain string) error {
	conn, err := net.Dial("tcp", servername+":445")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     username,
			Password: password,
			Domain:   domain,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		panic(err)
	}
	defer s.Logoff()

	fs, err := s.Mount("\\\\" + servername + "\\" + sharename)
	if err != nil {
		panic(err)
	}
	defer fs.Umount()

	// List all the files
	dir, err := fs.Open("")
	if err != nil {
		return err
	}
	fis, err := dir.Readdir(10)
	if err != nil {
		return err
	}
	for i := range fis {
		log.Println("found file: " + fis[i].Name())

		f, err := fs.Open(fis[i].Name())
		if err != nil {
			log.Println("file not found")
		}
		defer fs.Remove(fis[i].Name())
		defer f.Close()

		bs, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("file not readable")
		}

		status := postData(bs, app.endpoint.url, fis[i].Name(), app)

		if status != "200 OK" {
			log.Println(status + ": file not sent")
			continue
		} else {
			log.Println(status + ": " + fis[i].Name() + " successfully sent")
		}

		log.Println("deleted " + fis[i].Name())
	}

	return nil
}

func postData(content []byte, url string, filename string, app *application) string {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(content))
	req.Header.Set("X-Custom-Header", "post-requester")
	req.Header.Set("Content-Type", "text/xml")
	req.SetBasicAuth(app.endpoint.username, app.endpoint.password)
	if err != nil {
		log.Println(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	return resp.Status
}
