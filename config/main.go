package config

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/go-git/go-git/v5"
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
	Noop    bool     `json:"noop"`
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
	err := s.cloneOrPull(c)
	if err != nil {
		return err
	}

	return s.install(c)
}

func (s *Source) cloneOrPull(c *Config) error {
	tmpdir, err := c.GetTmpDir()
	if err != nil {
		return err
	}

	path := filepath.Join(tmpdir, s.GetName())
	gitPath := filepath.Join(path, ".git", "objects")

	exists, err := pathExists(gitPath)
	if err != nil {
		return err
	}

	if exists {
		err = pullSource(path)
	} else {
		err = cloneSource(path, s.URL)
	}
	return err
}

func (s *Source) install(c *Config) error { //revive:disable-line cyclomatic
	tmpdir, err := c.GetTmpDir()
	if err != nil {
		return err
	}
	path := filepath.Join(tmpdir, s.GetName())

	root, err := c.GetRoot()
	if err != nil {
		return err
	}

	return filepath.Walk(path, func(itemAbs string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		item, err := filepath.Rel(path, itemAbs)
		if err != nil {
			return err
		}
		ignore, err := shouldIgnore(item, s.Ignore)
		if err != nil {
			return err
		}
		if ignore {
			return nil
		}

		targetItem := filepath.Join(root, item)

		if info.IsDir() {
			return os.MkdirAll(targetItem, 0750)
		}

		contents, err := os.ReadFile(itemAbs)
		if err != nil {
			return err
		}

		uptodate, err := isUpToDate(contents, targetItem)
		if err != nil {
			return err
		}
		if uptodate {
			return nil
		}

		if c.Noop {
			fmt.Printf("(%s) Would install %s\n", s.GetName(), item)
		} else {
			fmt.Printf("(%s) Installing %s\n", s.GetName(), item)
			err = os.WriteFile(targetItem, contents, info.Mode())
		}
		return err
	})
}

func isUpToDate(contents []byte, file string) (bool, error) {
	exists, err := pathExists(file)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	destContents, err := os.ReadFile(file)
	if err != nil {
		return false, err
	}
	newHash := sha256.Sum256(contents)
	oldHash := sha256.Sum256(destContents)
	return newHash == oldHash, nil
}

func shouldIgnore(item string, list []string) (bool, error) {
	defaultIgnore, err := regexp.Compile("(^|/)(.git|.gitignore|.gitmodules)(/|$)")
	if err != nil {
		return false, err
	}
	if defaultIgnore.MatchString(item) {
		return true, nil
	}
	for _, ignoreString := range list {
		pattern, err := regexp.Compile(ignoreString)
		if err != nil {
			return false, err
		}
		if pattern.MatchString(item) {
			return true, nil
		}
	}
	return false, nil
}

func pullSource(path string) error {
	r, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Force:             true,
	})
	if err == git.NoErrAlreadyUpToDate {
		err = nil
	}
	return err
}

func cloneSource(path, url string) error {
	err := os.MkdirAll(path, 0750)
	if err != nil {
		return err
	}

	_, err = git.PlainClone(path, false, &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	return err
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
