package config

import (
	"encoding/json"
	"ftp/common"
	"os"
	"regexp"
)

type config struct {
	InitDir              string   `json:"initDir"`
	RestrictedPathsSlice []string `json:"restrictedPaths"`
	restrictedPathsRegex *regexp.Regexp
}

func (c *config) parse(path string) error {
	cf, err := os.Open(path)
	if err != nil {
		return err
	}
	defer cf.Close()
	if err = json.NewDecoder(cf).Decode(c); err != nil {
		return err
	}

	if c.InitDir, err = common.ToAbsToSlash(c.InitDir); err != nil {
		return err
	}

	var regexStr string
	for _, p := range c.RestrictedPathsSlice {
		if p, err = common.ToAbsToSlash(p); err != nil {
			return err
		}
		regexStr += p + "/*|"
	}
	if len(regexStr) > 0 {
		regexStr = regexStr[:len(regexStr)-1]
		if c.restrictedPathsRegex, err = regexp.Compile(regexStr); err != nil {
			return err
		}
	}
	c.RestrictedPathsSlice = nil
	return nil
}

func (c *config) GetInitDir() string {
	return c.InitDir
}

func (c *config) IsRestricted(path string) bool {
	if len(path) > 0 && path[len(path)-1] != '/' {
		path += "/"
	}
	return c.restrictedPathsRegex != nil && c.restrictedPathsRegex.MatchString(path)
}

func (c *config) IsInBounds(path string) bool {
	return false
}
