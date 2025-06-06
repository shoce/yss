package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	NL = "\n"

	NameReString = "^[a-z][-a-z0-9]*$"

	ContentLengthLimit = 108 * 1024
)

var (
	DEBUG bool

	ListenAddr string = ":80"
	DataDir    string = "/yss/"

	NameRe *regexp.Regexp
)

func init() {
	NameRe = regexp.MustCompile(NameReString)

	if v := os.Getenv("DEBUG"); v != "" {
		DEBUG = true
	}

	if v := os.Getenv("ListenAddr"); v != "" {
		ListenAddr = v
	}

	if v := os.Getenv("DataDir"); v != "" {
		DataDir = v
	}
}

func main() {

	http.HandleFunc("/", yss)

	go func() {
		for {
			log("http: serving requests on %s.", ListenAddr)
			err := http.ListenAndServe(ListenAddr, nil)
			if err != nil {
				log("ERROR http: listen: %+v", err)
			}
			retryinterval := 3 * time.Second
			log("http: retrying listen in %s.", retryinterval)
			time.Sleep(retryinterval)
		}
	}()

	for {
		time.Sleep(11 * time.Second)
	}

}

func yss(rw http.ResponseWriter, req *http.Request) {
	fname := req.URL.Path
	fname = strings.Trim(fname, "/")
	if fname == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if !NameRe.MatchString(fname) {
		log("name `%s` does not match the allowed regexp", fname)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Method == http.MethodGet {

		if bb, err := ioutil.ReadFile(path.Join(DataDir, fname)); err != nil {
			log("ERROR read file %s: %v", fname, err)
			rw.WriteHeader(http.StatusNotFound)
			return
		} else {
			//rw.Header().Set("Content-Type", "application/x-yaml")
			rw.WriteHeader(http.StatusOK)
			if _, err := rw.Write(bb); err != nil {
				log("ERROR write response: %v", err)
			}
			if DEBUG {
				log("DEBUG get %s", fname)
			}
		}

	} else if req.Method == http.MethodPut {

		if req.ContentLength > ContentLengthLimit {
			log("WARNING request content length %d too big", req.ContentLength)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		if finfo, err := os.Stat(path.Join(DataDir, fname)); err != nil {
			log("ERROR stat file %s: %v", fname, err)
			if os.IsNotExist(err) {
				rw.WriteHeader(http.StatusNotFound)
			} else {
				rw.WriteHeader(http.StatusInternalServerError)
			}
			return
		} else if !finfo.Mode().IsRegular() {
			log("ERROR file %s is not a regular file", fname)
			rw.WriteHeader(http.StatusInternalServerError)
		}

		if bb, err := ioutil.ReadAll(req.Body); err != nil {
			log("ERROR read request body: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		} else {

			if err := os.WriteFile(path.Join(DataDir, fname), bb, 0644); err != nil {
				log("ERROR write file %s: %v", fname, err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			rw.WriteHeader(http.StatusOK)
			if DEBUG {
				log("DEBUG put %s: %d bytes", fname, req.ContentLength)
			}

		}

	} else {

		rw.WriteHeader(http.StatusBadRequest)
		return

	}

	return
}

func log(msg string, args ...interface{}) {
	tnow := time.Now().In(time.FixedZone("IST", 330*60))
	ts := fmt.Sprintf(
		"%d%02d%02d:%02d%02d",
		tnow.Year()%1000, tnow.Month(), tnow.Day(),
		tnow.Hour(), tnow.Minute(),
	)
	fmt.Fprintf(os.Stderr, ts+" "+msg+NL, args...)
}
