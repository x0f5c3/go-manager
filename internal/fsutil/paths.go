package fsutil

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/x0f5c3/zerolog/log"
)

func UserConfigDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return os.UserHomeDir()
	default:
		return os.UserConfigDir()
	}
}

var UserName = func() string {
	u, err := user.Current()
	if err != nil {
		log.Error().Err(err).Msg("failed to get current user")
		return "gom"
	}
	return u.Username
}()

func AppDataDir(parent ...string) string {
	var path string
	root, err := func() (string, error) {
		if len(parent) > 0 {
			return parent[0], nil
		}
		return UserConfigDir()
	}()
	if err != nil {
		log.Error().Err(err).Msg("failed to get user config dir")
		path = "gom"
	} else {
		path = filepath.Join(root, "gom")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		log.Error().Err(err).Msg("failed to get absolute path")
		return path
	}
	return abs
}

func TodayLogDirPath(parent string) string {
	return filepath.Join(parent, time.Now().Local().Format("2006_01_02"))
}

func CurrentLogPath(logDir string) string {
	return filepath.Join(logDir, fmt.Sprintf("%s.log", time.Now().Format("15_04_05")))
}

func InitializeDirStructure() error {
	err := CreateDir(EnvDir)
	if err != nil {
		return err
	}
	err = CreateDir(LogDir)
	if err != nil {
		return err
	}
	return nil
}

var DataDir = DefaultDataDir
var EnvDir = DefaultEnvDir
var LogDir = DefaultLogDir
var TodayLogDir = DefaultTodayLogDir
var CurrentLogFile = DefaultCurrentLogFile
var DefaultDataDir = AppDataDir()
var DefaultEnvDir = filepath.Join(DefaultDataDir, "envs")
var DefaultLogDir = filepath.Join(DefaultDataDir, "logs")
var DefaultTodayLogDir = TodayLogDirPath(DefaultLogDir)
var DefaultCurrentLogFile = CurrentLogPath(DefaultTodayLogDir)
var DefaultConfigFilename = "gom.toml"
var DefaultConfigName = "gom"
var DefaultConfigPath = filepath.Join(DefaultDataDir, DefaultConfigFilename)
