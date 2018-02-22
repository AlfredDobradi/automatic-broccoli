package config

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

type Options struct {
	Port    uint16 `json:"port"`
	Host    string `json:"host"`
	Backend string `json:"backend"`

	TLS struct {
		Cert string `json:"cert"`
		Key  string `json:"key"`
	} `json:"tls"`

	Timescale struct {
		URI string `json:"uri"`
	} `json:"timescale"`
}

func Load(file string) (Options, error) {
	options := Options{
		Port:    9001,
		Host:    "0.0.0.0",
		Backend: "stdout",
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			return options, errors.Wrap(err, "config file doesn't exist")
		}
		return options, errors.Wrap(err, "unknown error when opening file")
	}

	err = json.Unmarshal(data, &options)
	if err != nil {
		return options, errors.Wrap(err, "error unmarshaling config")
	}

	return options, nil
}
