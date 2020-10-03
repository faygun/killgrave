package killgrave

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

// Config representation of config file yaml
type Config struct {
	ImpostersPath string      `yaml:"imposters_path"`
	Port          int         `yaml:"port"`
	Host          string      `yaml:"host"`
	CORS          ConfigCORS  `yaml:"cors"`
	Proxy         ConfigProxy `yaml:"proxy"`
}

// ConfigCORS representation of section CORS of the yaml
type ConfigCORS struct {
	Methods          []string `yaml:"methods"`
	Headers          []string `yaml:"headers"`
	Origins          []string `yaml:"origins"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
}

// ConfigProxy is a representation of section proxy of the yaml
type ConfigProxy struct {
	Url  string    `yaml:"url"`
	Mode ProxyMode `yaml:"mode"`
}

// ProxyMode is enumeration of proxy server modes
type ProxyMode uint8

const (
	// ProxyNone server is off
	ProxyNone ProxyMode = iota
	// ProxyMissing handle only missing requests are proxied
	ProxyMissing
	// ProxyAll all requests are proxied
	ProxyAll
)

var (
	ErrInvalidConfigPath = errors.New("invalid config file")
)

func (p ProxyMode) String() string {
	m := map[ProxyMode]string{
		ProxyNone:    "none",
		ProxyMissing: "missing",
		ProxyAll:     "all",
	}

	s, ok := m[p]
	if !ok {
		return "none"
	}
	return s
}

// StringToProxyMode convert string into a ProxyMode if not exists return a none mode and an error
func StringToProxyMode(t string) (ProxyMode, error) {
	m := map[string]ProxyMode{
		"none":    ProxyNone,
		"missing": ProxyMissing,
		"all":     ProxyAll,
	}

	p, ok := m[t]
	if !ok {
		return ProxyNone, fmt.Errorf("unknown proxy mode: %s", t)
	}

	return p, nil
}

// UnmarshalYAML implementation of yaml.Unmarshaler interface
func (p *ProxyMode) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var proxyMode string
	if err := unmarshal(&proxyMode); err != nil {
		return err
	}

	m, err := StringToProxyMode(proxyMode)
	if err != nil {
		return err
	}

	*p = m
	return nil
}

// ConfigureProxy preparing the server with the proxy configuration that the user has indicated
func (cfg *Config) ConfigureProxy(proxyMode ProxyMode, proxyURL string) {
	cfg.Proxy.Mode = proxyMode
	cfg.Proxy.Url = proxyURL
}

// ConfigOpt function to encapsulate optional parameters
type ConfigOpt func(cfg *Config) error

// NewConfig initialize< the config
func NewConfig(impostersPath, host string, port int, opts ...ConfigOpt) (Config, error) {
	cfg := Config{
		ImpostersPath: impostersPath,
		Host:          host,
		Port:          port,
	}

	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}

// NewConfigFromFile  unmarshal content of config file to initialize a Config struct
func NewConfigFromFile(cfgPath string) (Config, error) {
	if cfgPath == "" {
		return Config{}, ErrInvalidConfigPath
	}
	configFile, err := os.Open(cfgPath)
	if err != nil {
		return Config{}, fmt.Errorf("%w: error trying to read config file: %s, using default configuration instead", err, cfgPath)
	}
	defer configFile.Close()

	var cfg Config
	bytes, _ := ioutil.ReadAll(configFile)
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return Config{}, fmt.Errorf("%w: error while unmarshall configFile file %s, using default configuration instead", err, cfgPath)
	}

	cfg.ImpostersPath = path.Join(path.Dir(cfgPath), cfg.ImpostersPath)

	return cfg, nil
}
