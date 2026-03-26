/*
Yaml Storage System
*/

// GoGet GoFmt GoBuildNull

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
	SP = " "
	NL = "\n"

	NameReString = "^[a-z][-a-z0-9]*$"

	ContentLengthLimit = 108 * 1024

	ListenAddrDef = ":80"
	DataDirDef    = "/yss/"
)

var (
	DEBUG bool

	ListenAddr string = ListenAddrDef
	DataDir    string = DataDirDef

	NameRe *regexp.Regexp

	TZIST = time.FixedZone("IST", 330*60)
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
			perr("serving http requests on [%s]", ListenAddr)
			err := http.ListenAndServe(ListenAddr, nil)
			if err != nil {
				perr("ERROR http listen %+v", err)
			}
			retryinterval := 3 * time.Second
			perr("retrying http listen in <%s>", retryinterval.Truncate(time.Second))
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
		perr("name [%s] does not match the allowed regexp", fname)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodGet:

		if bb, err := ioutil.ReadFile(path.Join(DataDir, fname)); err != nil {
			perr("ERROR read file [%s] %v", fname, err)
			rw.WriteHeader(http.StatusNotFound)
			return
		} else {
			perr("DEBUG get [%s]", fname)
			//rw.Header().Set("Content-Type", "application/x-yaml")
			rw.WriteHeader(http.StatusOK)
			if _, err := rw.Write(bb); err != nil {
				perr("ERROR write response %v", err)
			}
		}

	case http.MethodPut:

		if req.ContentLength > ContentLengthLimit {
			perr("WARNING request content length <%d> is more than allowed <%d>", req.ContentLength, ContentLengthLimit)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		if finfo, err := os.Stat(path.Join(DataDir, fname)); err != nil {
			perr("ERROR stat file [%s] %v", fname, err)
			if os.IsNotExist(err) {
				rw.WriteHeader(http.StatusNotFound)
			} else {
				rw.WriteHeader(http.StatusInternalServerError)
			}
			return
		} else if !finfo.Mode().IsRegular() {
			perr("ERROR file [%s] is not a regular file", fname)
			rw.WriteHeader(http.StatusInternalServerError)
		}

		if bb, err := ioutil.ReadAll(req.Body); err != nil {
			perr("ERROR read request body %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		} else {

			perr("DEBUG put [%s] <%d> bytes", fname, req.ContentLength)

			if err := os.WriteFile(path.Join(DataDir, fname), bb, 0644); err != nil {
				perr("ERROR write file [%s] %v", fname, err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			rw.WriteHeader(http.StatusOK)

		}

	default:

		rw.WriteHeader(http.StatusBadRequest)
		return

	}

	return
}

func perr(msg string, args ...interface{}) {
	if strings.HasSuffix(msg, "DEBUG ") && !DEBUG {
		return
	}
	tnow := time.Now().In(TZIST)
	ts := fmt.Sprintf(
		"<%d:%02d%02d:%02d%02dॐ>",
		tnow.Year()%1000, tnow.Month(), tnow.Day(),
		tnow.Hour(), tnow.Minute(),
	)
	msgtext := msg
	if len(args) > 0 {
		msgtext = fmt.Sprintf(msg, args...)
	}
	fmt.Fprint(os.Stderr, ts+SP+msgtext+NL)
}
