package fsutil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/x0f5c3/zerolog/log"

	"github.com/x0f5c3/go-manager/internal/config"
)

// FindExistingParent finds the first existing parent directory of the given path
func FindExistingParent(path string) string {
	exists := CheckExists(path)
	for !exists {
		msg := fmt.Sprintf("%s doesn't exist,", path)
		evt := log.Debug().Str("From", path)
		path = filepath.Dir(path)
		msg += fmt.Sprintf(" going to %s", path)
		evt.Str("To", path).Msg(msg)
		exists = CheckExists(path)
	}
	return path
}

var InstallCmd = &cobra.Command{
	Use:   "install [directory]",
	Short: "Install gom to the given directory",
	Long:  `Install gom to the given directory`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			args = append(args, DefaultDataDir)
		} else if len(args) > 1 {
			return fmt.Errorf("too many arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := args[0]
		exists, perms := CheckExistsWritable(dataDir)
		if !exists {
			log.Info().Str("Path", dataDir).Msg("data directory doesn't exist")
		}
		if !perms {
			log.Error().Err(fmt.Errorf("no write permissions")).Str("Path", dataDir).Msg("data directory is not writable")
			return fmt.Errorf("data directory is not writable")
		}
		installInfo, err := CopyAppToDataDir(dataDir)
		if err != nil {
			return err
		}
		b, err := toml.Marshal(config.Conf)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal config")
			return errors.Wrap(err, "failed to marshal config")
		}
		if err := os.WriteFile(installInfo.ConfigPath, b, 0644); err != nil {
			log.Error().Err(err).Msgf("failed to write %s", installInfo.ConfigPath)
			return errors.Wrap(err, "failed to write config")
		}
		b, err = toml.Marshal(installInfo)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal install info")
			return errors.Wrap(err, "failed to marshal install info")
		}
		if err := os.WriteFile(filepath.Join(dataDir, "install.toml"), b, 0644); err != nil {
			log.Error().Err(err).Msgf("failed to write %s", filepath.Join(dataDir, "install.toml"))
			return errors.Wrap(err, "failed to write install info")
		}
		return nil
	},
	Hidden: true,
}

func CopyAppToDataDir(dataDir string) (*InstallInformation, error) {
	binDir := filepath.Join(dataDir, "bin")
	selfPath, err := FindMyself()
	if err != nil {
		log.Error().Err(err).Msg("failed to find myself")
		return nil, errors.Wrap(err, "failed to find myself")
	}
	selfBytes, err := os.ReadFile(selfPath)
	if err != nil {
		log.Error().Err(err).Msg("failed to read myself")
		return nil, errors.Wrap(err, "failed to read myself")
	}
	selfName := filepath.Base(selfPath)
	newSelfPath := filepath.Join(binDir, selfName)
	if !CheckExists(binDir) {
		if err := CreateDir(binDir); err != nil {
			log.Error().Err(err).Msgf("failed to create %s directory", binDir)
			return nil, errors.Wrap(err, "failed to create bin dir")
		}
	} else {
		if CheckExists(newSelfPath) {
			if err := os.Remove(newSelfPath); err != nil {
				log.Error().Err(err).Msgf("failed to remove %s", newSelfPath)
				return nil, errors.Wrap(err, "failed to remove old self")
			}
		}
	}
	err = os.WriteFile(newSelfPath, selfBytes, 0755) //nolint:gosec
	if err != nil {
		log.Error().Err(err).Msgf("failed to write %s", newSelfPath)
		return nil, errors.Wrap(err, "failed to write new self")
	}
	return &InstallInformation{
		InstallPath: dataDir,
		ConfigPath:  filepath.Join(dataDir, "gom.toml"),
		ExePath:     newSelfPath,
	}, nil
}

type InstallInformation struct {
	InstallPath string
	ConfigPath  string
	ExePath     string
}

// GetInstallInformation returns the install information for the given executable
func GetInstallInformation() (*InstallInformation, error) {
	exe, err := FindMyself()
	if err != nil {
		log.Error().Err(err).Msgf("failed to get absolute path for %s", exe)
		return nil, errors.Wrapf(err, "failed to get absolute path for %s", exe)
	}
	exeDir := filepath.Dir(exe)
	installPath := filepath.Dir(exeDir)
	configPath := filepath.Join(installPath, "gom.toml")
	return &InstallInformation{
		InstallPath: installPath,
		ConfigPath:  configPath,
		ExePath:     exe,
	}, nil
}

func FindMyself() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		log.Error().Err(err).Msg("failed to get executable path")
		return "", errors.Wrap(err, "failed to find executable")
	}
	return filepath.EvalSymlinks(ex)
}

// CheckExists checks if the given path exists
func CheckExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false
	} else if errors.Is(err, fs.ErrPermission) || errors.Is(err, fs.ErrExist) {
		return true
	}
	return false
}

func writeCheck(dir string) bool {
	f, err := os.CreateTemp(dir, "gomTest")
	if err != nil {
		log.Error().Err(err).Msgf("failed to create test file in %s", dir)
		return false
	}
	name := f.Name()
	err = os.Remove(name)
	if err != nil {
		log.Error().Err(err).Msgf("failed to remove the temp file %s", name)
		return false
	}
	return true
}

// CheckPerms checks if the given path is writable
func CheckPerms(path string) bool {
	origPath := path
	path = FindExistingParent(path)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrPermission) {
		log.Error().Err(err).Msgf("failed to access %s", origPath)
		return false
	} else if err != nil {
		log.Error().Err(err).Msgf("failed to access %s", origPath)
		return false
	}
	if path == origPath && CheckExists(path) {
		log.Debug().Str("Path", path).Msgf("checked %s and it exists and we have perms to stat, let's see if we have write", origPath)
	} else {
		log.Debug().Str("Found parent", path).Str("Original path", origPath).Msgf("checked %s and it doesn't exist, let's see if we have write in the existing parent", origPath)
	}
	return writeCheck(path)
}

func CheckExistsWritable(path string) (bool, bool) {
	originalPath := path
	exists := CheckExists(path)
	if !exists {
		path = FindExistingParent(path)
	}
	permsCheck := CheckPerms(path)
	log.Debug().Str("OriginalPath", originalPath).Str("Found parent", path).Bool("Exists", exists).Bool("Perms", permsCheck).Msgf("checked %s", originalPath)
	return exists, permsCheck
}

func MustCreateDir(dir string) {
	if CheckExists(dir) {
		return
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to get absolute path for %s", dir)
	}
	if err := os.MkdirAll(abs, 0755); err != nil {
		log.Fatal().Err(err).Msgf("failed to create %s directory", abs)
	}
}

func CreateDir(dir string) error {
	if CheckExists(dir) {
		return nil
	}
	log.Debug().Msgf("%s doesn't exist, creating...", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Error().Err(err).Msgf("failed to create %s directory", dir)
		return errors.Wrapf(err, "failed to create %s directory", dir)
	}
	return nil
}
