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

		bs, err := ioutil.ReadAll(f)
		if err != nil {
			util.ErrorLogger.Printf("failed to read file: %s", fis[i].Name())
		}

		response, err := postData(bs, util.App.Endpoint.URL, fis[i].Name(), app)
		if err != nil {
			util.ErrorLogger.Printf("failed to post data: %s", fis[i].Name())
		}

		// log response body and status code
		util.InfoLogger.Printf("status code: %d", response.StatusCode)
		util.InfoLogger.Printf("response body: %s", response.Body)
		util.InfoLogger.Printf("response headers: %s", response.Header)

		// return status code is 200 OK remove file
		if response.StatusCode == 200 {
			if err := fs.Remove(fis[i].Name()); err != nil {
				util.ErrorLogger.Printf("failed to remove file: %s", fis[i].Name())
			}

			util.InfoLogger.Printf("removed file: %s", fis[i].Name())

			// return status code is not 200 OK
		} else {
			util.ErrorLogger.Printf("failed to post data: %s", fis[i].Name())
		}

		// close file
		if err := f.Close(); err != nil {
			util.ErrorLogger.Printf("failed to close file: %s", fis[i].Name())
		}
	}

	return nil
}

func postData(content []byte, url string, filename string, app *util.Application) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(content))
	req.Header.Set("X-Custom-Header", "post-requester")
	req.Header.Set("Content-Type", "text/xml")
	req.SetBasicAuth(util.App.Endpoint.Username, util.App.Endpoint.Password)
	if err != nil {
		util.ErrorLogger.Println(err)
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		util.ErrorLogger.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	return resp, nil
}
