package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"regexp"
)

type Config struct {
	Server   string `yaml:"server" json:"server"`
	UpsName  string `yaml:"upsName" json:"upsName"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	Port     uint16 `yaml:"port" json:"port"`
	Refresh  int    `yaml:"refresh" json:"refresh"`
}

var (
	showConfig    = kingpin.Flag("config.show", "Show actual configuration and ends").Default("false").Bool()
	configFile    = kingpin.Flag("config.file", "Configuration file default is \"nut.yml\".").PlaceHolder("cfg.yml").Default("nut.yml").String()
	server        = kingpin.Flag("nut.server", "NUT server FQDn or IP address").PlaceHolder("server").Default("").String()
	user          = kingpin.Flag("nut.user", "NUT user for read data").PlaceHolder("user").Default("").String()
	pwd           = kingpin.Flag("nut.pwd", "NUT user password").PlaceHolder("pwd").Default("").String()
	upsName       = kingpin.Flag("nut.ups", "name of UPS on NUT server").PlaceHolder("ups").Default("ups").String()
	listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":8100").String()
	config        = &Config{
		Server:   "",
		UpsName:  "ups",
		User:     "",
		Password: "",
		Port:     3493,
		Refresh:  10,
	}
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (c *Config) LoadFile(filename string) error {
	if fileExists(filename) {
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		err = yaml.UnmarshalStrict(content, c)
		if err != nil {
			err = json.Unmarshal(content, c)
			if err != nil {
				return err
			}
		}
	}
	if len(*server) > 0 {
		c.Server = *server
	}
	if len(*user) > 0 {
		c.User = *user
	}
	if len(*pwd) > 0 {
		c.Password = *pwd
	}
	if len(*upsName) > 0 {
		c.UpsName = *upsName
	}

	match, err := regexp.MatchString("^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$", c.Server)
	if !match || err != nil {
		match, err = regexp.MatchString("^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\\-]*[A-Za-z0-9])$", c.Server)
		if !match || err != nil {
			return errors.New("NUT server address isn't valid FQDN or IP address")
		}
	}
	if len(c.User) < 1 {
		return errors.New("NUT User must be defined")
	}
	if len(c.Password) < 1 {
		return errors.New("NUT User password must be defined")
	}
	if len(c.UpsName) < 1 {
		return errors.New("UPS name must be defined")
	}
	if c.Port < 1024 || c.Port > 65535 {
		return errors.New("defined port not valid")
	}
	if c.Refresh < 5 || c.Refresh > 300 {
		return errors.New("refresh time is out of range (5-300 sec)")
	}
	return nil
}

func (c *Config) getServer() string {
	return fmt.Sprintf("%s:%d", c.Server, c.Port)
}

func (c *Config) print() string {
	p := "Not set!"
	if len(c.Password) > 0 {
		p = "****"
	}
	a := fmt.Sprintf("\r\n%s\r\nActual configuration:\r\n", applicationName)
	a = fmt.Sprintf("%sUPS name:     [%s]\r\n", a, c.UpsName)
	a = fmt.Sprintf("%sNUT Server :  [%s:%d]\r\n", a, c.Server, c.Port)
	a = fmt.Sprintf("%sUser:         [%s]\r\n", a, c.User)
	a = fmt.Sprintf("%sPassword:     [%s]\r\n", a, p)
	return a
}
