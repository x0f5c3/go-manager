package pkg

import (
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
)

type GOPATH struct {
	Dir string
	GOROOT
}

type GOROOT struct {
	Dir   string
	paths map[string]*GOPATH
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
