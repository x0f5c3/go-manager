package config

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/x0f5c3/zerolog/log"

	"github.com/x0f5c3/go-manager/internal/fsutil"
	"github.com/x0f5c3/go-manager/pkg/semver"
)


var initCmd = &cobra.Command{
	Use: "init [data directory]",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			args = append(args, fsutil.DefaultDataDir)
			return nil
		} else if len(args) > 1 {
			return errors.New("too many arguments")
		}
		return nil
	},
	Short: "Initialize config",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := args[0]
		err := setupDataDir(dataDir)

	}
}

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Args:  cobra.ExactArgs(2),
	Short: "Set config",
}

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Args:  cobra.ExactArgs(1),
	Short: "Get config",
}

var flagSet = configFlagSet()



func setupDataDir(datadir string) error {
	if !fsutil.CheckExists(datadir) {
		err := fsutil.CreateDir(datadir)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create data dir")
			return err
		}
	}
	if !fsutil.CheckExists(filepath.Join(datadir, "envs")) {
		err := fsutil.CreateDir(filepath.Join(datadir, "envs"))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create envs dir")
			return err
		}
	}
	if !fsutil.CheckExists(filepath.Join(datadir, "logs")) {
		err := fsutil.CreateDir(filepath.Join(datadir, "logs"))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create logs dir")
			return err
		}
	}
	err := setupTodayLogDir(filepath.Join(datadir, "logs"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create today log dir")
		return err
	}
	return nil
}

func todayLogDir(args ...string) string {
	args = append(args, time.Now().Local().Format("2006-01-02"))
	return filepath.Join(args...)
}

func setupTodayLogDir(logDir string) error {
	dirName := todayLogDir(logDir)
	if !fsutil.CheckExists(dirName) {
		err := fsutil.CreateDir(dirName)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create today log dir")
			return err
		}
	}
	return nil
}

func decoderHookSemver() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		v, ok := data.(string)
		if f.Kind() == reflect.String && t == reflect.TypeOf(&semver.Version{}) && ok {
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
		log.Error().Err(err).Msg("Failed to get go version")
		return nil, err
	}
	log.Debug().Str("out", string(out)).Msg("go version")
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
	Proxies    []string        `mapstructure:"proxies, omitempty"`
	LastUpdate time.Time       `mapstructure:"last_update, omitempty"`
	ConfigFile string          `mapstructure:"config_file, omitempty"`
	EnvsDir    string          `mapstructure:"envs_dir"`
	Current    *semver.Version `mapstructure:"current"`
	mod        bool            `mapstructure:"-"`
}

func (c *Config) SetProxies(Proxies []string) {
	c.mod = true
	c.Proxies = Proxies
}

func (c *Config) SetLastUpdate(LastUpdate time.Time) {
	c.mod = true
	c.LastUpdate = LastUpdate
}

func (c *Config) SetConfigFile(ConfigFile string) {
	c.mod = true
	c.ConfigFile = ConfigFile
}

func (c *Config) SetEnvsDir(EnvsDir string) {
	c.mod = true
	c.EnvsDir = EnvsDir
}

func (c *Config) SetCurrent(Current *semver.Version) {
	c.mod = true
	c.Current = Current
}

func checkFlagsExists(flags *pflag.FlagSet) []string {
	var flagNames []string
	if !flags.Changed("proxies") {
		flagNames = append(flagNames, "proxies")
	}
	if !flags.Changed("envs_dir") {
		flagNames = append(flagNames, "envs-dir")
	}
	if !flags.Changed("current") {
		flagNames = append(flagNames, "current")
	}
	if !flags.Changed("config") {
		flagNames = append(flagNames, "config")
	}
	return flagNames
}

var defaultFlagSet = configFlagSet()

func configFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("config", pflag.ContinueOnError)
	flags.StringSliceVar(&config.Proxies, "proxies", []string{}, "Proxies to use")
	flags.StringVar(&config.EnvsDir, "envs-dir", "", "Directory to store go envs")
	flags.Var(config.Current, "current", "Current go version")
	flags.StringVarP(&config.ConfigFile, "config", "c", "", "Config file")
	return flags
}

func BindFlags(flags *pflag.FlagSet) error {
	flags.StringSliceVar(&config.Proxies, "proxies", []string{}, "Proxies to use")
	err := viper.BindPFlag("proxies", flags.Lookup("proxies"))
	if err != nil {
		return err
	}
	flags.StringVar(&config.EnvsDir, "envs-dir", "", "Directory to store go envs")
	err = viper.BindPFlag("envs_dir", flags.Lookup("envs-dir"))
	if err != nil {
		return err
	}
	flags.Var(config.Current, "current", "Current go version")
	err = viper.BindPFlag("current", flags.Lookup("current"))
	if err != nil {
		return err
	}
	flags.StringVarP(&config.ConfigFile, "config", "c", "", "Config file")
	err = viper.BindPFlag("config_file", flags.Lookup("config"))
	if err != nil {
		return err
	}
	return nil
}

func ReadInConfig(confDir ...string) (*Config, error) {
	return tryRead(confDir...)
}

func (c *Config) Save() error {
	c.LastUpdate = time.Now()
	viper.Set("last_update", c.LastUpdate)
	viper.Set("proxies", c.Proxies)
	viper.Set("envs_dir", c.EnvsDir)
	viper.Set("current", c.Current)
	return viper.WriteConfigAs(c.ConfigFile)
}

func tryRead(toTry ...string) (*Config, error) {
	if len(toTry) == 0 {
		toTry = append(toTry, fsutil.DefaultConfigPath)
	}
	var foundFiles []string
	for _, v := range toTry {
		if !fsutil.CheckExists(v) {
			log.Error().Str("path", v).Msg("File does not exist")
			continue
		}
		foundFiles = append(foundFiles, v)
	}
	var latestConf string
	var latestTime time.Time
	rest := make(map[string]fs.FileInfo)
	restKeys := make([]string, 0)
	for _, v := range foundFiles {
		info, err := os.Stat(v)
		if err != nil {
			log.Error().Err(err).Str("path", v).Msg("Failed to read config")
			continue
		}
		rest[v] = info
		if latestTime == (time.Time{}) {
			latestTime = info.ModTime()
			latestConf = v
			continue
		}
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestConf = v
		}
	}
	delete(rest, latestConf)
	for k := range rest {
		restKeys = append(restKeys, k)
	}
	sort.Slice(restKeys, func(i, j int) bool {
		return rest[restKeys[j]].ModTime().After(rest[restKeys[i]].ModTime())
	})
	if fsutil.CheckExists(latestConf) {
		conf, err := tryReadFile(latestConf)
		if err != nil {
			log.Error().Err(err).Str("path", latestConf).Msg("Failed to read config")
			if len(restKeys) > 0 {
				for _, v := range restKeys {
					_, ok := rest[v]
					if !ok {
						continue
					}
					conf, err = tryReadFile(v)
					if err != nil {
						log.Error().Err(err).Str("path", v).Msg("Failed to read config")
						continue
					}
					return conf, nil
				}
			} else {
				return nil, err
			}
			if conf == nil {
				return nil, errors.New("failed to read config")
			}
		}
		return conf, nil
	}
	return nil, errors.New("default config file does not exist")
}

func tryReadFile(path string) (*Config, error) {
	if !fsutil.CheckExists(path) {
		return nil, errors.New("file does not exist")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close file")
		}
	}(f)
	var conf Config
	err = toml.NewDecoder(f).Decode(&conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

var config = defaultConfig()

// func init() {
// 	create, err := configDirCreate()
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to create config dir")
// 		return
// 	}
// 	res, err := InitConfig(create)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to init config")
// 		return
// 	}
// 	config = res
// 	err = config.Save()
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to save config")
// 		return
// 	}
// }

func configDirCreate() (string, error) {
	if fsutil.CheckExists(fsutil.DefaultDataDir) {
		return fsutil.DefaultDataDir, nil
	}
	// Create config dir
	err := fsutil.CreateDir(fsutil.DefaultDataDir)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create config dir")
		return "", err
	}
	return fsutil.DefaultDataDir, nil
}

func mustWriteConfig() {
	err := Vip.WriteConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to write config")
	}
}

func mustInitConfig() {
	conf, err := InitConfig(fsutil.DefaultDataDir, WithDefaults())
	if err != nil {
		config = defaultConfig()

	}
	config = conf
	err = config.Save()
	if err != nil {
		log.Error().Err(err).Msg("Failed to save config")
	}
}

func defaultConfig() *Config {
	curr, err := currentVersion()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get current version, using nil")
		curr = nil
	} else {
		log.Debug().Str("Version", curr.String()).Msg("Got current version")
	}
	return &Config{
		Proxies:    nil,
		EnvsDir:    fsutil.DefaultEnvDir,
		ConfigFile: fsutil.DefaultConfigPath,
		LastUpdate: time.Now(),
		Current:    curr,
		mod:        false,
	}
}
