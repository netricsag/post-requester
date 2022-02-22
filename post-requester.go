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
		err := sendSMBFiles()
		if err != nil && !strings.Contains(err.Error(), "EOF") {
			util.ErrorLogger.Println(err)
		}
		time.Sleep(time.Duration(util.App.Interval.Seconds) * time.Second)
	}
}

func sendSMBFiles() error {
	conn, err := net.Dial("tcp", util.App.SMB.Servername+":445")
	if err != nil {
		return err
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     util.App.SMB.Username,
			Password: util.App.SMB.Password,
			Domain:   util.App.SMB.Domain,
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

	fs, err := s.Mount("\\\\" + util.App.SMB.Servername + "\\" + util.App.SMB.Sharename)
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

		response, err := postData(bs)
		if err != nil {
			util.ErrorLogger.Printf("failed to post data: %s", fis[i].Name())
		} else {
			util.InfoLogger.Printf("posted data: %s", fis[i].Name())
		}

		if response != nil {
			util.InfoLogger.Printf("status code: %d", response.StatusCode)
		} else {
			util.InfoLogger.Printf("no response")
		}

		// close file handle
		if err := f.Close(); err != nil {
			util.ErrorLogger.Printf("failed to close file: %s; error: %s", fis[i].Name(), err)
			continue
		}

		// delete file if status code is 200
		if response.StatusCode == 200 {
			if err := fs.Remove(fis[i].Name()); err != nil {
				util.ErrorLogger.Printf("failed to remove file: %s; error: %s", fis[i].Name(), err)
			}

			util.InfoLogger.Printf("removed file: %s", fis[i].Name())
		}
	}

	return nil
}

func postData(content []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", util.App.Endpoint.URL, bytes.NewBuffer(content))
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

	return resp, nil
}
