package cmd

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/x0f5c3/zerolog/log"

	"github.com/spf13/viper"

	"github.com/x0f5c3/go-manager/pkg/semver"
)

var envDir = func() string {
	home, err := os.UserConfigDir()
	if err != nil {
		checkErr(err)
	}
	return filepath.Join(home, "goenvs")
}()

func decoderHookSemver() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		v, ok := data.(string)
		if f.Kind() == reflect.String && t == reflect.TypeOf(semver.Version{}) && ok {
			if semver.IsValid(v) {
				return semver.Parse(v), nil
			}
		}
		return data, nil
	}
}

func currentVersion() (*semver.Version, error) {
	c := exec.Command("go", "version")
	out, err := c.Output()
	if err != nil {
		return nil, err
	}
	sp := strings.Split(string(out), " ")
	if len(sp) < 3 {
		return nil, err
	}
	res := strings.ReplaceAll(sp[2], "go", "v")
	if semver.IsValid(res) {
		return semver.Parse(res)
	} else {
		return nil, errors.New("invalid version")
	}
}

type Config struct {
	proxies []string        `mapstructure:"proxies, omitempty"`
	envsDir string          `mapstructure:"envs_dir"`
	current *semver.Version `mapstructure:"current"`
}

func viperToConfig() (*Config, error) {
	var config Config
	err := viper.Unmarshal(&config, viper.DecodeHook(decoderHookSemver()))
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func initConfig() (*Config, error) {
	viper.SetConfigName("gom")
	viper.AddConfigPath(".")
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(filepath.Join(configDir, "gom"))
	viper.SetDefault("envs_dir", envDir)
	current, err := currentVersion()
	if err != nil {
		viper.SetDefault("current", nil)
	} else {
		viper.SetDefault("current", current)
	}
	viper.SetEnvPrefix("GOM")
	viper.AutomaticEnv()
	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return viperToConfig()
}

var config *Config

func initconfig() {
	configDirCreate()
	// Set config
	// v := viper.New()
	// v.SetEnvPrefix("gom")
	// v.AutomaticEnv()
	// v.SetDefault("current", currentVersion())
	// confDir, err := os.UserConfigDir()
	// checkErr(err)
	// v.AddConfigPath(confDir)
	// v.AddConfigPath(filepath.Join(confDir, "gom"))
	// var maybeConf Config
	// err = v.Unmarshal(&maybeConf)
	res, err := initConfig()
	if err != nil {
		log.Error().Err(err).Msg("Failed to init config")
		config = &Config{}
		return
	}
	// if maybeConf.current == "" {
	//	maybeConf.current = currentVersion()
	// }
	config = res
}

func configDirCreate() (string, error) {
	// Create config dir
	confDir, err := os.UserConfigDir()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user config dir")
		confDir, err = filepath.Abs("gom")
		if err != nil {
			log.Error().Err(err).Msg("Failed to get current dir")
			confDir = "gom"
		}
		if _, err = os.Stat(confDir); errors.Is(err, fs.ErrNotExist) {
			err = os.MkdirAll(confDir, 0755)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create config dir")
			}
		}
	} else {
		confDir = filepath.Join(confDir, "gom")
		if _, err = os.Stat(confDir); errors.Is(err, fs.ErrNotExist) {
			err = os.MkdirAll(confDir, 0755)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create config dir")
			}
		}
	}
	if confDir == "" {
		log.Error().Err(err).Msg("Failed to get config dir")
	}
}
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
		proxies: nil,
		envsDir: envDir,
		current: nil,
	}
}

// func (c *Config) ToFlags() *pflag.FlagSet {
// 	res := pflag.NewFlagSet("config", pflag.ExitOnError)
// 	res.StringSliceVar(&c.proxies, "proxies", []string{}, "Proxies to use")
// 	res.StringVar(&c.envsDir, "envs-dir", envDir, "Directory to store environments in")
// 	res.Var
// }
