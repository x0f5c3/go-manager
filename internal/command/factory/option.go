package config

import (
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/x0f5c3/go-manager/internal/fsutil"
)

type Option func(*viper.Viper)

func WithConfigName(name string) Option {
	return func(v *viper.Viper) {
		v.SetConfigName(name)
	}
}

func WithConfigType(t string) Option {
	return func(v *viper.Viper) {
		v.SetConfigType(t)
	}
}

func WithExactPath(path string) Option {
	return func(v *viper.Viper) {
		v.SetConfigFile(path)
	}
}

func WithViper(v *viper.Viper) Option {
	return func(v *viper.Viper) {
		Vip = v
	}
}

func WithDefaults() Option {
	def := defaultConfig()
	return func(v *viper.Viper) {
		setViperDefaults(v, def)
		v.SetConfigName("gom")
		v.AddConfigPath(fsutil.DefaultDataDir)
		v.AddConfigPath(filepath.Join(".", "gom"))
		v.AddConfigPath(".")
		v.SetEnvPrefix("GOM")
		v.AutomaticEnv()
	}
}

func WithConfigDir(dir string) Option {
	return func(v *viper.Viper) {
		v.AddConfigPath(dir)
	}
}

func WithEnvPrefix(prefix string) Option {
	return func(v *viper.Viper) {
		v.SetEnvPrefix(prefix)
	}
}
