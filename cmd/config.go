package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/viper"
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
		return semver.MustParse("0.0.0")
	}
	sp := strings.Split(string(out), " ")
	if len(sp) < 3 {
		return semver.MustParse("0.0.0")
	}
	res, err := semver.NewVersion(strings.TrimPrefix(sp[2], "go"))
	if err != nil {
		return semver.MustParse("0.0.0")
	}
	return res
}

type Config struct {
	proxies []string        `mapstructure:"proxies"`
	envsDir string          `mapstructure:"envs_dir"`
	current *semver.Version `mapstructure:"current"`
}

func viperToConfig() (*Config, error) {
	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func initConfig() (*Config, error) {
	viper.SetConfigName("gom")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(filepath.Join(configDir, "gom"))
	viper.SetDefault("proxies", []string{})
	viper.SetDefault("envs_dir", envDir)
	viper.SetDefault("current", currentVersion())
	viper.SetEnvPrefix("GOM")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return viperToConfig()
}

var config *Config

func mustWriteConfig() {
	err := viper.WriteConfig()
	if err != nil {
		checkErr(err)
	}
}

func mustInitConfig() {
	conf, err := initConfig()
	if err != nil {
		config = defaultConfig()

	}
	config = conf
	err = viper.SafeWriteConfig()
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
