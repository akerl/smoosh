package config

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/go-git/go-git/v5"
)

func cloneOrPullSource(path, url string) error {
	gitPath := filepath.Join(path, ".git", "objects")

	exists, err := pathExists(gitPath)
	if err != nil {
		return err
	}

	if exists {
		err = pullSource(path)
	} else {
		err = cloneSource(path, url)
	}
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

func copySource(path, root string, s *Source) error {
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

		exists, err := pathExists(targetItem)
		if err != nil {
			return err
		}

		if exists {
			destContents, err := os.ReadFile(targetItem)
			if err != nil {
				return err
			}
			newHash := sha256.Sum256(contents)
			oldHash := sha256.Sum256(destContents)
			if newHash == oldHash {
				return nil
			}
		}

		fmt.Printf("(%s) Installing %s\n", s.GetName(), item)
		//return os.WriteFile(targetItem, contents, info.Mode())
		return nil
	})
}

func shouldIgnore(item string, list []string) (bool, error) {
	defaultIgnore, err := regexp.Compie("(^|/)(.git|.gitignore|.gitmodules)(/|$)")
	if err != nil {
		return false, err
	}
	if defaultIgnore.MatchString(item) == true {
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
