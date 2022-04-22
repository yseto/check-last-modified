package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/mackerelio/checkers"
)

func main() {
	ckr := run(os.Args[1:])
	ckr.Name = "LastModified"
	ckr.Exit()
}

var opts struct {
	Url      string `short:"u" long:"url" required:"true" description:"monitor url"`
	Warning  int64  `short:"w" long:"warning" default:"3600" description:"warning if more old than"`
	Critical int64  `short:"c" long:"critical" default:"86400" description:"critical if more old than"`
}

func run(args []string) *checkers.Checker {
	_, err := flags.ParseArgs(&opts, args)
	if err != nil {
		os.Exit(1)
	}

	resp, err := http.Get(opts.Url)
	if err != nil {
		return checkers.Unknown(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return checkers.Critical(fmt.Sprintf("Response Code is %d", resp.StatusCode))
	}

	t, err := time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified"))
	if err != nil {
		return checkers.Unknown(err.Error())
	}

	var d interface{}
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return checkers.Critical(fmt.Sprintf("Cannot parse %s", err.Error()))
	}

	result := checkers.OK

	age := time.Now().Unix() - t.Unix()

	if opts.Warning < age {
		result = checkers.WARNING
	}

	if opts.Critical < age {
		result = checkers.CRITICAL
	}

	msg := fmt.Sprintf("%s is %d seconds old (%02d:%02d:%02d)", opts.Url, age, t.Hour(), t.Minute(), t.Second())
	return checkers.NewChecker(result, msg)
}
