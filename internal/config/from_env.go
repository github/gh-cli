package config

import (
	"fmt"
	"os"

	"github.com/cli/cli/internal/ghinstance"
)

const (
	GH_TOKEN                = "GH_TOKEN"
	GITHUB_TOKEN            = "GITHUB_TOKEN"
	GH_ENTERPRISE_TOKEN     = "GH_ENTERPRISE_TOKEN"
	GITHUB_ENTERPRISE_TOKEN = "GITHUB_ENTERPRISE_TOKEN"
)

func InheritEnv(c Config) Config {
	return &envConfig{Config: c}
}

type envConfig struct {
	Config
}

func (c *envConfig) Hosts() ([]string, error) {
	hasDefault := false
	hosts, err := c.Config.Hosts()
	for _, h := range hosts {
		if h == ghinstance.Default() {
			hasDefault = true
		}
	}
	token, _ := AuthTokenFromEnv(ghinstance.Default())
	if (err != nil || !hasDefault) && token != "" {
		hosts = append([]string{ghinstance.Default()}, hosts...)
		return hosts, nil
	}
	return hosts, err
}

func (c *envConfig) Get(hostname, key string) (string, error) {
	val, _, err := c.GetWithSource(hostname, key)
	return val, err
}

func (c *envConfig) GetWithSource(hostname, key string) (string, string, error) {
	if hostname != "" && key == "oauth_token" {
		if token, env := AuthTokenFromEnv(hostname); token != "" {
			return token, env, nil
		}
	}

	return c.Config.GetWithSource(hostname, key)
}

func (c *envConfig) CheckWriteable(hostname, key string) error {
	if hostname != "" && key == "oauth_token" {
		if token, env := AuthTokenFromEnv(hostname); token != "" {
			return fmt.Errorf("The '%s' environment variable is not empty, please clear it to save authentication data", env)
		}
	}

	return c.Config.CheckWriteable(hostname, key)
}

func AuthTokenFromEnv(hostname string) (string, string) {
	if ghinstance.IsEnterprise(hostname) {
		if token := os.Getenv(GH_ENTERPRISE_TOKEN); token != "" {
			return token, GH_ENTERPRISE_TOKEN
		}

		return os.Getenv(GITHUB_ENTERPRISE_TOKEN), GITHUB_ENTERPRISE_TOKEN
	}

	if token := os.Getenv(GH_TOKEN); token != "" {
		return token, GH_TOKEN
	}

	return os.Getenv(GITHUB_TOKEN), GITHUB_TOKEN
}

func AuthTokenProvidedFromEnv() bool {
	return os.Getenv(GH_ENTERPRISE_TOKEN) != "" ||
		os.Getenv(GITHUB_ENTERPRISE_TOKEN) != "" ||
		os.Getenv(GH_TOKEN) != "" ||
		os.Getenv(GITHUB_TOKEN) != ""
}
