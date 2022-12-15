package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
)

const (
	defaultConfigName = ".smoosh/config.yml"
	defaultTmpDirName = ".smoosh/cache"
)

// Source defines a repo to pull files from
type Source struct {
	URL    string   `json:"url"`
	Name   string   `json:"name"`
	Ignore []string `json:"ignore"`
}

// Config defines a set of sources and associated metadata
type Config struct {
	Root    string   `json:"root"`
	Sources []Source `json:"sources"`
	TmpDir  string   `json:"tmpdir"`
}

// NewConfig loads a config from a given file or the default location
func NewConfig(file string) (Config, error) {
	var c Config
	var err error

	path := file
	if path == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return c, err
		}
		path = filepath.Join(homedir, defaultConfigName)
	}

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(contents, &c)
	return c, err
}

// GetName parses the name for a source
func (s *Source) GetName() string {
	if s.Name == "" {
		parts := strings.Split(s.URL, "/")
		s.Name = parts[len(parts)-1]
	}
	return s.Name
}

// GetRoot parses the root location for copying
func (c *Config) GetRoot() (string, error) {
	if c.Root == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		c.Root = homedir
	}
	return c.Root, nil
}

// GetTmpDir parses the tmpdir for managing source repos
func (c *Config) GetTmpDir() (string, error) {
	if c.TmpDir == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		c.TmpDir = filepath.Join(homedir, defaultTmpDirName)
	}
	return c.TmpDir, nil
}

// Sync updates a config set of sources
func (c *Config) Sync() error {
	for _, s := range c.Sources {
		err := s.Sync(c)
		if err != nil {
			return err
		}
	}
	return nil
}

// Sync updates a single source
func (s *Source) Sync(c *Config) error {
	tmpdir, err := c.GetTmpDir()
	if err != nil {
		return err
	}
	root, err := c.GetRoot()
	if err != nil {
		return nil
	}

	path := filepath.Join(tmpdir, s.GetName())

	err = cloneOrPullSource(path, s.URL)
	if err != nil {
		return err
	}

	return copySource(path, root, s)
}
