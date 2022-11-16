package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/coreos/go-semver/semver"
)

var envDir = func() string {
	home, err := os.UserConfigDir()
	if err != nil {
		checkErr(err)
	}
	return filepath.Join(home, "goenvs")
}()

func currentVersion() *semver.Version {
	c := exec.Command("go", "version")
	out, err := c.Output()
	if err != nil {
		return semver.New("0.0.0")
	}
	sp := strings.Split(string(out), " ")
	if len(sp) < 3 {
		return semver.New("0.0.0")
	}
	res, err := semver.NewVersion(strings.TrimPrefix(sp[2], "go"))
	if err != nil {
		return semver.New("0.0.0")
	}
	return res
}

type Config struct {
	proxies []string
	envsDir string
	current *semver.Version
}

func defaultConfig() *Config {
	return &Config{
		proxies: []string{},
		envsDir: envDir,
		current: currentVersion(),
	}
}

// func (c *Config) ToFlags() *pflag.FlagSet {
// 	res := pflag.NewFlagSet("config", pflag.ExitOnError)
// 	res.StringSliceVar(&c.proxies, "proxies", []string{}, "Proxies to use")
// 	res.StringVar(&c.envsDir, "envs-dir", envDir, "Directory to store environments in")
// 	res.Var
// }
