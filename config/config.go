package config

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"time"

	"gopkg.in/yaml.v2"
)

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

type Config struct {
	CMD          CMD           `yaml:"cmd"`
	Environments []Environment `yaml:"environments"`
	SSHUser      string        `yaml:"ssh-user"`
	DPSetupPath  string        `yaml:"dp-setup-path"`
}

type CMD struct {
	MongoURL    string   `yaml:"mongo-url"`
	Neo4jURL    string   `yaml:"neo4j-url"`
	MongoDBs    []string `yaml:"mongo-dbs"`
	Hierarchies []string `yaml:"hierarchies"`
	Codelists   []string `yaml:"codelists"`
}

// Environment represents an environment
type Environment struct {
	Name    string `yaml:"name"`
	Profile string `yaml:"profile"`
}

func Get() (*Config, error) {
	path := os.Getenv("DP_CLI_CONFIG")
	if len(path) == 0 {
		return nil, errors.New("no DP_CLI_CONFIG config file specified")
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Dump() ([]byte, error) {
	c, err := Get()
	if err != nil {
		return nil, err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetMyIP fetches your external IP address
func GetMyIP() (string, error) {
	if ip := os.Getenv("MY_IP"); len(ip) > 0 {
		isIP, err := regexp.Match(`^\d{1,3}(?:\.\d{1,3}){3}(?:/\d{1,2})?$`, []byte(ip))
		if err != nil {
			return "", err
		}
		if !isIP {
			return "", errors.New("unexpected format for var MY_IP")
		}
		return ip, nil
	}

	res, err := httpClient.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}

	defer func() {
		io.Copy(ioutil.Discard, res.Body)
	}()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status code fetching IP: %d", res.StatusCode)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
