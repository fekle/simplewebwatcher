// Package config provides configuration stuff
package config

import (
	"io"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

// Site is used to store information about the sites to check
type Site struct {
	Description string
	URL         string
	Username    string
	Password    string
	LastCheck   time.Time
	LastBytes   int
	LastHash    string
}

// Config is used to store all the sites in an array
type Config struct {
	Site []Site
}

var (
	// default configuration
	defaultConfig = &Config{
		Site: []Site{
			{
				Description: "First",
				URL:         "http://localhost",
				Username:    "user",
				Password:    "password",
				LastCheck:   time.Now(),
				LastBytes:   int(0),
				LastHash:    "",
			},
			{
				Description: "Second",
				URL:         "http://localhost",
				Username:    "user",
				Password:    "password",
				LastCheck:   time.Now(),
				LastBytes:   int(0),
				LastHash:    "",
			},
			{
				Description: "Third",
				URL:         "http://localhost",
				Username:    "user",
				Password:    "password",
				LastCheck:   time.Now(),
				LastBytes:   int(0),
				LastHash:    "",
			},
		},
	}
)

// NewDefaultConfig returns a new default config
func NewDefaultConfig() *Config {
	return defaultConfig
}

// ReadConfig reads a config from the given string
func ReadConfig(data string) (*Config, error) {
	var conf Config
	if _, err := toml.Decode(data, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

// WriteConfig writes the given configuration to the given Writer
func WriteConfig(conf *Config, w io.Writer) error {
	encoder := toml.NewEncoder(w)

	if err := encoder.Encode(conf); err != nil {
		return err
	}

	return nil
}

// ThreadSafeConfigWrapper is a wrapper for reading and editing a configuration struct from multiple goroutines
type ThreadSafeConfigWrapper struct {
	lock   sync.RWMutex
	config Config
}

// Get returns the stored config, threadsafe
func (t *ThreadSafeConfigWrapper) Get() Config {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.config
}

// Set sets the config, threadsafe
func (t *ThreadSafeConfigWrapper) Set(c Config) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.config = c
}

// SetSite sets a specific site, threadsafe
func (t *ThreadSafeConfigWrapper) SetSite(pos int, s Site) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.config.Site[pos] = s
}
