package pkg

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/x0f5c3/zerolog/log"
)

func findExistingParent(targetPath string) string {
	path := targetPath
	exists := checkExists(path)
	log.Debug().Str("Path", path).Bool("Exists", exists).Msgf("checking %s", path)
	for !exists {
		path = filepath.Dir(path)
		exists = checkExists(path)
	}
	return path
}

func checkExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false
	} else if errors.Is(err, fs.ErrPermission) || errors.Is(err, fs.ErrExist) {
		return true
	}
	return false
}

func checkPerms(path string) bool {
	origPath := path
	if !checkExists(path) {
		path = findExistingParent(path)
	}
	if _, err := os.Stat(path); errors.Is(err, fs.ErrPermission) {
		log.Error().Err(err).Msgf("failed to access %s", origPath)
		return false
	} else if err != nil {
		log.Error().Err(err).Msgf("failed to access %s", origPath)
		return false
	}
	f, err := os.CreateTemp(filepath.Dir(origPath), "gomTest")
	if errors.Is(err, fs.ErrPermission) {
		log.Error().Err(err).Msgf("failed to create %s, no perms", origPath)
		return false
	} else if err != nil {
		log.Error().Err(err).Msgf("failed to create %s, something's wrong", origPath)
		return false
	}
	name := f.Name()
	err = os.Remove(name)
	if err != nil {
		log.Error().Err(err).Msgf("failed to remove the temp file %s", name)
	}
	return true
}

func checkExistsWritable(path string) (bool, bool) {
	originalPath := path
	exists := checkExists(path)
	for !checkExists(path) {
		newPath := filepath.Dir(path)
		log.Debug().Str("Child", path).Str("Parent", newPath).Msgf("going from %s to %s", path, newPath)
		path = newPath
	}
	log.Debug().Str("Path", path).Msgf("found existing parent, checking perms for %s", path)
	permsCheck := checkPerms(path)
	log.Debug().Str("Path", originalPath).Bool("Exists", exists).Bool("Perms", permsCheck).Msgf("checked %s", originalPath)
	return exists, permsCheck
}

func createDir(dir string) error {
	exists, perms := checkExistsWritable(dir)
	if !exists {
		if perms {
			log.Error().Msgf("failed to create %s, no perms", dir)
			return errors.New("failed to create " + dir + ", no perms")
		}
		log.Debug().Msgf("%s doesn't exist, creating...", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Error().Err(err).Msgf("failed to create %s directory", dir)
			return errors.Wrapf(err, "failed to create %s directory", dir)
		}
	}
	return nil
}

type GOPATH struct {
	EnvName string
	Dir     string
	BinDir  string
	GOROOT
}

type GOROOT struct {
	Dir     string
	paths   map[string]*GOPATH
	current *GOPATH
}

func (g *GOPATH) GetEnv() string {
	return g.EnvName
}

func (g *GOROOT) GetCurrent() *GOPATH {
	return g.current
}

func (g *GOROOT) NewEnv(name string, path string) error {
	if g.paths == nil {
		g.paths = make(map[string]*GOPATH)
	}
	if v, ok := g.paths[name]; ok && v != nil {
		return errors.New("env already exists")
	}
	g.paths[name] = &GOPATH{
		EnvName: name,
		Dir:     path,
		BinDir:  filepath.Join(path, "bin"),
		GOROOT:  *g,
	}
	return nil
}

func (g *GOROOT) SetCurrent(envName string) error {
	if g.paths == nil {
		g.paths = make(map[string]*GOPATH)
	}
	if g.paths[envName] == nil {
		g.paths[envName] = &GOPATH{
			EnvName: envName,
			Dir:     filepath.Join(EnvsDir, envName),
			GOROOT:  *g,
		}
	}
	g.current = g.paths[envName]
	return nil
}

type GoPath struct {
	Dir    string
	BinDir string
	Root   GoRoot
}

type GoRoot struct {
	Dir    string
	BinDir string
}

func WinPathSetter(path string) string {
	return fmt.Sprintf("if (-NOT $env:PATH.Split(';').Contains('%s')) "+
		"{ "+
		"[Environment]::SetEnvironmentVariable("+
		"'Path',"+
		"[Environment]::GetEnvironmentVariable('Path', [EnvironmentVariableTarget]::User) + '%s',"+
		"[EnvironmentVariableTarget]::User)"+
		"}", path, path)
}

func WinPathUnsetter(path string) string {
	return fmt.Sprintf("[Environment]::SetEnvironmentVariable("+
		"'Path',"+
		"[Environment]::GetEnvironmentVariable('Path', [EnvironmentVariableTarget]::User).Replace('%s', '')"+
		",[EnvironmentVariableTarget]::User)", path)
}

func UnixPathSetter(path string) string {
	return fmt.Sprintf("if [[ ! $PATH =~ (^|:)%s(:|$) ]]; then export PATH=$PATH:%s; fi", path, path)
}

func UnixPathUnsetter(path string) string {
	return fmt.Sprintf("export PATH=${PATH//:%s/}", path)
}

var EnvsDir = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		pterm.Fatal.Println(err)
	}
	return filepath.Join(home, ".goenvs")
}()

var AliasDir = filepath.Join(EnvsDir, "bin")

//goland:noinspection ALL
func SwitchEnv(envName string) error {
	binDir := filepath.Join(EnvsDir, envName, "bin")
	bins, err := os.ReadDir(binDir)
	if err != nil {
		return err
	}
	for _, v := range bins {
		err = os.Symlink(filepath.Join(binDir, v.Name()), filepath.Join(AliasDir, v.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}
