package commands

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/urfave/cli/v2"
)

var ServerURL string

var serverURLFlag = &cli.StringFlag{
	Name:        "url",
	Usage:       "The jetbridge url to connect to",
	Aliases:     []string{"u"},
	Value:       "http://localhost:8080",
	Destination: &ServerURL,
	Action: func(c *cli.Context, v string) error {
		u, err := url.Parse(v)
		if err != nil {
			return err
		}

		if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("failed to parse server url, invalid scheme: %q", u.Scheme)
		}

		if u.Host == "" {
			return fmt.Errorf("failed to parse server url, invalid host: %q", u.Host)
		}

		if i, err := strconv.Atoi(u.Port()); err != nil || i < 1 || i > 65535 {
			return fmt.Errorf("failed to parse server url, invalid port: %q", u.Port())
		}

		return nil
	},
}
