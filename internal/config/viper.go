package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/x0f5c3/zerolog/log"

	"github.com/x0f5c3/go-manager/internal/fsutil"
)

type FactoryCommand struct {
	ctx     context.Context
	factory *AppFactory
	*cobra.Command
}

func (f *FactoryCommand) AddCommand(cmd *FactoryCommand) {
	f.Command.AddCommand(cmd.Command)
}

func (f *FactoryCommand) Execute() error {
	return f.Command.Execute()
}

func NewCommand(ctx context.Context, factory *AppFactory, cmd *cobra.Command) *FactoryCommand {
	return &FactoryCommand{
		ctx:     ctx,
		factory: factory,
		Command: cmd,
	}
}

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

type AppFactory struct {
	ctx       context.Context
	conf      *Config
	err       error
	loaded    bool
	opts      []Option
	exactPath string
	direct    bool
	v         *viper.Viper
	cmd       *cobra.Command
}

func NewConfigFactory() *AppFactory {
	f := &AppFactory{
		v: commonViper(),
		cmd: &cobra.Command{
			Use:   "config",
			Short: "Manage config",
		},
		loaded: false,
	}
	f.cmd.PersistentFlags().AddFlagSet(flagSet)
	f.cmd.AddCommand(initCmd)
	f.cmd.AddCommand(setCmd)
	f.cmd.AddCommand(getCmd)
	return f
}

func DefaultConfigFactory() *AppFactory {
	return NewConfigFactory().WithDefaults().WithExactPath(fsutil.DefaultConfigPath)
}

func (c *AppFactory) GetCommand() *cobra.Command {
	c.cmd.PersistentFlags().AddFlagSet(configFlagSet())
	_ = c.v.BindPFlags(configFlagSet())
	return c.cmd
}

func (c *AppFactory) WithConfigName(name string) *AppFactory {
	c.opts = append(c.opts, WithConfigName(name))
	return c
}

// WithExactPath sets the config file path,
// and causes the factory to first try to load the config by:
// - Reading the file and deserializing from toml
// - If that fails, it will try to load the config from viper
// This is the default behavior of the factory
// It takes less time to deserialize from toml than to load from viper
// This function does not modify the viper configuration, to allow it to be used as fallback in case the exact path fails
func (c *AppFactory) WithExactPath(path string) *AppFactory {
	c.exactPath = path
	c.direct = true
	return c
}

func (c *AppFactory) WithContext(ctx context.Context) *AppFactory {
	c.ctx = ctx
	c.cmd.SetContext(c.ctx)
	return c
}

func (c *AppFactory) WithViper(v *viper.Viper) *AppFactory {
	c.opts = append(c.opts, WithViper(v))
	return c
}

func (c *AppFactory) WithDefaults() *AppFactory {
	c.opts = append(c.opts, WithDefaults())
	return c
}

func (c *AppFactory) WithConfigType(t string) *AppFactory {
	c.opts = append(c.opts, WithConfigType(t))
	return c
}

func (c *AppFactory) Save() error {
	if c.conf == nil || !c.loaded {
		return errors.New("config not loaded")
	}
	if c.err != nil {
		return c.err
	}
	err := c.conf.Save()
	if err != nil {
		c.err = err
		return err
	}
	return nil
}

func LoadConfig(path string) (*Config, error) {
	pb, err := pterm.DefaultSpinner.WithText("Loading config").WithShowTimer(true).WithTimerRoundingFactor(time.Second).Start()
	if err != nil {
		log.Error().Err(err).Msg("Error starting spinner")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		pb.Fail(pterm.Red(fmt.Sprintf("Failed to read config from %s", path)))
		return nil, err
	}
	pb.Success(pterm.Green(fmt.Sprintf("Read config from %s", path)))
	log.Info().Str("path", path).Msg("Read config from path, trying to deserialize")
	var conf Config
	err = toml.Unmarshal(b, &conf)
	if err != nil {
		pb.Fail(pterm.Red("Failed to deserialize config"))
		log.Error().Err(err).Msg("Failed to deserialize config")
		return nil, err
	}
	pb.Success(pterm.Green("Deserialized config"))
	return &conf, nil
}

func (c *AppFactory) Load(opts ...Option) (*Config, error) {
	if c.loaded {
		log.Info().Msg("Config already loaded")
		return c.conf, c.err
	}
	if c.direct && c.exactPath != "" {
		log.Info().Str("path", c.exactPath).Msg("Loading config from exact path")
		c.conf, c.err = LoadConfig(c.exactPath)
		if c.err != nil {
			log.Error().Err(c.err).Msg("Error loading config from exact path")
		} else {
			c.loaded = true
			return c.conf, c.err
		}
	}
	for _, opt := range opts {
		opt(c.v)
	}
	c.err = c.v.ReadInConfig()
	if c.err != nil {
		return nil, c.err
	}
	c.conf, c.err = viperToConfig(c.v)
	if c.err == nil {
		log.Info().Str("config", c.v.ConfigFileUsed()).Msg("Loaded config")
		c.loaded = true
		return c.conf, c.err
	}
	log.Error().Err(c.err).Msg("Failed to load config, fallback to default")
	c.conf = config
	c.loaded = true
	return c.conf, nil
}

func (c *AppFactory) Config() (*Config, error) {
	if c.conf == nil || !c.loaded {
		return nil, errors.New("config not loaded")
	}
	if c.err != nil {
		log.Error().Err(c.err).Msg("Config error")
	}
	return c.conf, c.err
}

var Vip = commonViper()

func setViperDefaults(v *viper.Viper, conf *Config) {
	if conf == nil {
		conf = defaultConfig()
	}
	v.SetDefault("last_update", time.Now())
	v.SetDefault("proxies", conf.Proxies)
	v.SetDefault("envs_dir", conf.EnvsDir)
	v.SetDefault("current", conf.Current)
	// current, err := currentVersion()
	// if err != nil {
	// 	v.SetDefault("current", nil)
	// } else {
	// 	v.SetDefault("current", current)
	// }
}

func InitDefaultConfig(opts ...Option) (*Config, error) {
	for _, opt := range opts {
		opt(Vip)
	}
	Vip.SetConfigFile(fsutil.DefaultConfigPath)
	err := Vip.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return viperToConfig(Vip)
}

func InitConfig(confDir string, opts ...Option) (*Config, error) {
	for _, opt := range opts {
		opt(Vip)
	}
	return InitConfigManual(confDir, Vip)
}

func commonViper() *viper.Viper {
	v := viper.New()
	v.OnConfigChange(func(in fsnotify.Event) {
		log.Info().Str("event", in.String()).Msg("Config file changed")
		config.LastUpdate = time.Now()
	})
	v.WatchConfig()
	v.SetConfigName("gom")
	v.AddConfigPath(fsutil.DefaultDataDir)
	v.AddConfigPath(filepath.Join(".", "gom"))
	setViperDefaults(v, config)
	v.SetEnvPrefix("GOM")
	v.AutomaticEnv()
	return v
}

func InitConfigManual(confDir string, v *viper.Viper) (*Config, error) {
	v.AddConfigPath(filepath.Join(confDir, "gom"))
	v.AddConfigPath(confDir)
	// v.SetDefault("envs_dir", fsutil.DefaultEnvDir)
	current, err := currentVersion()
	if err == nil {
		v.SetDefault("current", current)
	}
	err = v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return viperToConfig(v)
}

func viperToConfig(v *viper.Viper) (*Config, error) {
	if v.GetTime("last_update").IsZero() {
		conf := defaultConfig()
		conf.LastUpdate = time.Now()
		err := conf.Save()
		if err != nil {
			log.Error().Err(err).Msg("Failed to save c")
			return nil, err
		}
		return conf, nil
	}
	var c Config
	err := v.Unmarshal(&c, viper.DecodeHook(decoderHookSemver()))
	if err != nil {
		return nil, err
	}
	return &c, nil
}
