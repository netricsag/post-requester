package main

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hirochachacha/go-smb2"
	"github.com/natron-io/post-requester/util"
)

func init() {
	util.InitLoggers()

	if err := util.LoadEnv(); err != nil {
		util.ErrorLogger.Fatal(err)
	}

	util.InfoLogger.Println("Loaded config")
}

func main() {
	util.InfoLogger.Println("Starting post-requester")
	for {
		err := sendSMBFiles(util.App, util.App.SMB.Servername, util.App.SMB.Sharename, util.App.SMB.Username, util.App.SMB.Password, util.App.SMB.Domain)
		if err != nil && !strings.Contains(err.Error(), "EOF") {
			util.ErrorLogger.Println(err)
		}
		time.Sleep(time.Duration(util.App.Interval.Seconds) * time.Second)
	}
}

func sendSMBFiles(app *util.Application, servername string, sharename string, username string, password string, domain string) error {
	conn, err := net.Dial("tcp", servername+":445")
	if err != nil {
		return err
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
		return err
	}

	defer func() {
		if err := s.Logoff(); err != nil {
			util.ErrorLogger.Println(err)
		}
	}()

	fs, err := s.Mount("\\\\" + servername + "\\" + sharename)
	if err != nil {
		return err
	}

	defer func() {
		if err := fs.Umount(); err != nil {
			util.ErrorLogger.Println(err)
		}
	}()

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
		util.InfoLogger.Printf("found file: %s", fis[i].Name())

		f, err := fs.Open(fis[i].Name())
		if err != nil {
			util.ErrorLogger.Printf("failed to open file: %s", fis[i].Name())
		}

		// wait until return of surrounding function and then remove the file and close the file
		defer func() {
			var err error
			err = fs.Remove(fis[i].Name())
			if err != nil {
				util.ErrorLogger.Printf("failed to remove file: %s; %s", fis[i].Name(), err)
			}

			err = f.Close()
			if err != nil {
				util.ErrorLogger.Printf("failed to close file: %s; %s", fis[i].Name(), err)
			}
		}()

		bs, err := ioutil.ReadAll(f)
		if err != nil {
			util.ErrorLogger.Printf("failed to read file: %s", fis[i].Name())
		}

		status, err := postData(bs, util.App.Endpoint.URL, fis[i].Name(), app)
		if err != nil {
			util.ErrorLogger.Printf("failed to post file: %s; %s", fis[i].Name(), err)
		}

		util.InfoLogger.Printf("posted file: %s; status: %s", fis[i].Name(), status)

		if status == "200 OK" || status == "201 Created" {
			return nil
		}
	}

	return nil
}

func postData(content []byte, url string, filename string, app *util.Application) (string, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(content))
	req.Header.Set("X-Custom-Header", "post-requester")
	req.Header.Set("Content-Type", "text/xml")
	req.SetBasicAuth(util.App.Endpoint.Username, util.App.Endpoint.Password)
	if err != nil {
		util.ErrorLogger.Println(err)
		return "", err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		util.ErrorLogger.Println(err)
		return "", err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			util.ErrorLogger.Println(err)
		}
	}()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		util.ErrorLogger.Printf("failed to post file: %s", filename)
		return resp.Status, err
	}
	return resp.Status, nil
}
