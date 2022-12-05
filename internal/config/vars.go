package config

import (
	"path/filepath"

	"github.com/x0f5c3/zerolog/log"

	"github.com/x0f5c3/go-manager/internal/fsutil"
)

var DefaultDataDir = fsutil.AppDataDir()
var dataDirOrig = DefaultDataDir
var DefaultEnvDir = filepath.Join(DefaultDataDir, "envs")
var DefaultConfigFilename = "gom.toml"
var DefaultConfigName = "gom"
var DefaultConfigPath = filepath.Join(DefaultDataDir, DefaultConfigFilename)

func refreshPaths() {
	installInfo, err := fsutil.GetInstallInformation()
	if err != nil {
		log.Error().Err(err).Msg("failed to get install information")
	} else {
		if modified() {
			return
		}
		DefaultDataDir = installInfo.InstallPath
		DefaultEnvDir = filepath.Join(DefaultDataDir, "envs")
		DefaultConfigPath = installInfo.ConfigPath
	}
	if DefaultDataDir != dataDirOrig {
	}
}

func modified() bool {
	if DefaultDataDir != dataDirOrig {
		log.Debug().Str("To", DefaultDataDir).Str("From", dataDirOrig).Msg("data dir changed")
		DefaultEnvDir = filepath.Join(DefaultDataDir, "envs")
		DefaultConfigPath = filepath.Join(DefaultDataDir, DefaultConfigFilename)
		return true
	}
	return false
}
