package logging

import (
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/x0f5c3/zerolog"
	"github.com/x0f5c3/zerolog/log"
	"github.com/x0f5c3/zerolog/pkgerrors"

	"github.com/x0f5c3/go-manager/internal/fsutil"
)

func initlogger(consoleOut bool, logFile string) error {
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

func initLogging(consoleOut bool, logDirs ...string) error {
	var logDir string
	if len(logDirs) > 0 {
		logDir = logDirs[0]
	} else {
		return initlogger(consoleOut, fsutil.CurrentLogFile)
	}
	todays := fsutil.TodayLogDirPath(logDir)
	if !fsutil.CheckExists(todays) {
		log.Warn().Msgf("failed to access %s", todays)
		err := fsutil.CreateDir(logDir)
		if err != nil {
			return err
		}
	}
	logFile := fsutil.CurrentLogPath(todays)
	perms := fsutil.CheckPerms(logFile)
	if !perms {
		log.Error().Bool("Perms", perms).Msgf("failed to access %s", logFile)
		return errors.New("failed to access log file")
	}
	return initlogger(consoleOut, logFile)
}
