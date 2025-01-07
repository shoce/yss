package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	yaml "gopkg.in/yaml.v3"
)

type YssConfig struct {
	YssUrl string `yaml:"-"`
	DEBUG  bool   `yaml:"DEBUG"`
}

func (config *YssConfig) Get() error {
	if config.DEBUG {
		log("DEBUG Config.Get %s", config.YssUrl)
	}

	req, err := http.NewRequest(http.MethodGet, config.YssUrl, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("yss response status %s", resp.Status)
	}

	rbb, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if config.DEBUG {
		//log("DEBUG Config.Get: %s", string(rbb))
	}

	if err := yaml.Unmarshal(rbb, config); err != nil {
		return err
	}

	if config.DEBUG {
		log("DEBUG Config.Get: %+v", config)
	}

	return nil
}

func (config *YssConfig) Put() error {
	if config.DEBUG {
		log("DEBUG Config.Put %s %+v", config.YssUrl, config)
	}

	rbb, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	if config.DEBUG {
		//log("DEBUG Config.Put %s", string(rbb))
	}

	req, err := http.NewRequest(http.MethodPut, config.YssUrl, bytes.NewBuffer(rbb))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("yss response status %s", resp.Status)
	}
	if config.DEBUG {
		//log("DEBUG Config.Put response status code %s", resp.Status)
	}

	return nil
}
