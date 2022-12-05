package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/x0f5c3/zerolog"
	"github.com/x0f5c3/zerolog/log"
	"github.com/x0f5c3/zerolog/pkgerrors"

	. "github.com/x0f5c3/go-manager/internal/fsutil"
)

type Level zerolog.Level

type MultiWriter struct {
	baseLvl zerolog.Level
	writers map[zerolog.Level]io.Writer
}

type LevelLogger interface {
	SetLevel()
	Level() Level
	zerolog.LevelWriter
}

func NewMultiWriter(baseLvl zerolog.Level, writers map[zerolog.Level]io.Writer) *MultiWriter {
	return &MultiWriter{
		baseLvl: baseLvl,
		writers: writers,
	}
}

var (
	defaultLogDir = func() string {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get home directory")
		}
		return filepath.Join(home, ".gom", "logs")
	}()
)

func todayLogDir(parent string) string {
	return filepath.Join(parent, time.Now().Local().Format("2006_01_02"))
}

func currentLogFile(logDir string) string {
	return filepath.Join(todayLogDir(logDir), fmt.Sprintf("%s.log", time.Now().Format("15_04_05")))
}

func initLogging(consoleOut bool, logDirs ...string) error {
	var logDir string
	if len(logDirs) > 0 {
		logDir = logDirs[0]
	} else {
		logDir = defaultLogDir
	}
	todays := todayLogDir(logDir)
	if !CheckExists(logDir) {
		log.Warn().Msgf("failed to access %s", logDir)
		err := CreateDir(logDir)
		if err != nil {
			return err
		}
	}
	logFile := currentLogFile(logDir)
	exists, perms := CheckExistsWritable(logFile)
	if !exists && perms {
		log.Warn().Bool("Exists", exists).Bool("Perms", perms).Msgf("failed to access %s", logFile)
		err := CreateDir(todays)
		if err != nil {
			return err
		}
	}
	var w io.Writer
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s", logFile)
	}
	if consoleOut {
		w = zerolog.MultiLevelWriter(f, zerolog.NewConsoleWriter())
	} else {
		w = f
	}
	log.Logger = zerolog.New(w).With().Timestamp().Caller().Stack().Logger()
	// log.Logger = log.Output(zerolog.MultiLevelWriter(zerolog.NewConsoleWriter(), zerolog.New(f).With().Timestamp().Caller().Stack().Logger()))
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFieldName = "time"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "msg"
	zerolog.ErrorFieldName = "error"
	zerolog.CallerFieldName = "caller"
	return nil
}
