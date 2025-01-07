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
	/*
		if err := os.MkdirAll(DataDir, 0755); err != nil {
			log("ERROR mkdir %s: %v", DataDir, err)
			os.Exit(1)
		}
	*/

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
	fname = strings.TrimLeft(fname, "/")
	fname = strings.TrimRight(fname, "/")
	if fname == "" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if !NameRe.MatchString(fname) {
		log("name `%s` does not match allowed regexp", fname)
		rw.WriteHeader(http.StatusBadRequest)
		//rw.Write([]byte("path must start with [a-z] and contain [-a-z0-9] chars only."))
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
			log("DEBUG get %s", fname)
		}

	} else if req.Method == http.MethodPut {

		if req.ContentLength > 12123 {
			log("WARNING request content length %d too big", req.ContentLength)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		if bb, err := ioutil.ReadAll(req.Body); err != nil {
			log("ERROR read request body: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			/*
				if strings.Contains(fname, "/") {
					fdir := path.Join(DataDir, path.Dir(fname))
					if err := os.MkdirAll(fdir, 0755); err != nil {
						log("ERROR mkdir %s: %v", fdir, err)
						rw.WriteHeader(http.StatusInternalServerError)
						return
					}
				}
			*/

			if err := os.WriteFile(path.Join(DataDir, fname), bb, 0644); err != nil {
				log("ERROR write file %s: %v", fname, err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			rw.WriteHeader(http.StatusOK)
			log("DEBUG put %s: %d bytes", fname, req.ContentLength)
		}

	} else {

		rw.WriteHeader(http.StatusBadRequest)
		//rw.Write([]byte("get or put methods only."))
		return

	}

	return
}

func log(msg string, args ...interface{}) {
	t := time.Now().Local()
	ts := fmt.Sprintf(
		"%03d%02d%02d:"+"%02d%02d",
		t.Year()%1000, t.Month(), t.Day(), t.Hour(), t.Minute(),
	)
	fmt.Fprintf(os.Stderr, ts+" "+msg+NL, args...)
}
