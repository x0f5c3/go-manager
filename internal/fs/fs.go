package fs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/x0f5c3/zerolog/log"
)

// FindExistingParent finds the first existing parent directory of the given path
func FindExistingParent(path string) string {
	exists := CheckExists(path)
	for !exists {
		msg := fmt.Sprintf("%s doesn't exist,", path)
		evt := log.Debug().Str("From", path)
		path = filepath.Dir(path)
		msg += fmt.Sprintf(" going to %s", path)
		evt.Str("To", path).Msg(msg)
		exists = CheckExists(path)
	}
	return path
}

// CheckExists checks if the given path exists
func CheckExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false
	} else if errors.Is(err, fs.ErrPermission) || errors.Is(err, fs.ErrExist) {
		return true
	}
	return false
}

func writeCheck(dir string) bool {
	f, err := os.CreateTemp(dir, "gomTest")
	if err != nil {
		log.Error().Err(err).Msgf("failed to create test file in %s", dir)
		return false
	}
	name := f.Name()
	err = os.Remove(name)
	if err != nil {
		log.Error().Err(err).Msgf("failed to remove the temp file %s", name)
		return false
	}
	return true
}

// CheckPerms checks if the given path is writable
func CheckPerms(path string) bool {
	origPath := path
	path = FindExistingParent(path)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrPermission) {
		log.Error().Err(err).Msgf("failed to access %s", origPath)
		return false
	} else if err != nil {
		log.Error().Err(err).Msgf("failed to access %s", origPath)
		return false
	}
	if path == origPath && CheckExists(path) {
		log.Debug().Str("Path", path).Msgf("checked %s and it exists and we have perms to stat, let's see if we have write", origPath)
	} else {
		log.Debug().Str("Found parent", path).Str("Original path", origPath).Msgf("checked %s and it doesn't exist, let's see if we have write in the existing parent", origPath)
	}
	return writeCheck(path)
}

func CheckExistsWritable(path string) (bool, bool) {
	originalPath := path
	exists := CheckExists(path)
	if !exists {
		path = FindExistingParent(path)
	}
	permsCheck := CheckPerms(path)
	log.Debug().Str("OriginalPath", originalPath).Str("Found parent", path).Bool("Exists", exists).Bool("Perms", permsCheck).Msgf("checked %s", originalPath)
	return exists, permsCheck
}

func MustCreateDir(dir string) {
	if CheckExists(dir) {
		return
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to get absolute path for %s", dir)
	}
	if err := os.MkdirAll(abs, 0755); err != nil {
		log.Fatal().Err(err).Msgf("failed to create %s directory", abs)
	}
}

func CreateDir(dir string) error {
	if CheckExists(dir) {
		return nil
	}
	log.Debug().Msgf("%s doesn't exist, creating...", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error().Err(err).Msgf("failed to create %s directory", dir)
		return errors.Wrapf(err, "failed to create %s directory", dir)
	}
	return nil
}
