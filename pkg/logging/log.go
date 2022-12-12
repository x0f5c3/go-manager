package logging

// type lvlPrefix pterm.Prefix
//
// var (
// 	debugLevelPrefix = pterm.Debug.Prefix
// 	infoLevelPrefix  = pterm.Info.Prefix
// 	warnLevelPrefix  = pterm.Warning.Prefix
// )
//
// type Level zerolog.Level
//
// type LevelPrinter struct {
// 	level Level
// 	*pterm.PrefixPrinter
// }
//
// func (p LevelPrinter) SetLevel(level Level) {
// 	p.level = level
// }
//
// func (p LevelPrinter) Level() Level {
// 	return p.level
// }

// func (p LevelPrinter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
// 	if level < p.level {
// 		return 0, nil
// 	}
//
// }

// type MultiWriter struct {
// 	baseLvl zerolog.Level
// 	writers map[zerolog.Level]zerolog.LevelWriter
// }
//
// func (m MultiWriter) Write(p []byte) (n int, err error) {
// 	for _, w := range m.writers {
// 		n, err = w.Write(p)
// 		if err != nil {
// 			return n, err
// 		}
// 	}
// 	return
// }

// func (m MultiWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
// 	if level < m.baseLvl {
// 		return 0, nil
// 	}
// 	for lvl, w := range m.writers {
// 		if lvl < level {
// 			continue
// 		}
// 		n, err = w.WriteLevel(level, p)
// 	}
// }
//
// type LevelLogger interface {
// 	SetLevel(level Level)
// 	Level() Level
// 	zerolog.LevelWriter
// }
//
// func NewMultiWriter(baseLvl Level, writers map[Level]LevelLogger) *MultiWriter {
// 	return &MultiWriter{
// 		baseLvl: baseLvl,
// 		writers: writers,
// 	}
// }
//
// var (
// 	defaultLogDir = func() string {
// 		home, err := os.UserHomeDir()
// 		if err != nil {
// 			log.Fatal().Err(err).Msg("failed to get home directory")
// 		}
// 		return filepath.Join(home, ".gom", "logs")
// 	}()
// )
//
// func todayLogDir(parent string) string {
// 	return filepath.Join(parent, time.Now().Local().Format("2006_01_02"))
// }
//
// func currentLogFile(logDir string) string {
// 	return filepath.Join(logDir, fmt.Sprintf("%s.log", time.Now().Format("15_04_05")))
// }
//
// func initLogging(consoleOut bool, logDirs ...string) error {
// 	var logDir string
// 	if len(logDirs) > 0 {
// 		logDir = logDirs[0]
// 	} else {
// 		logDir = defaultLogDir
// 	}
// 	todays := todayLogDir(logDir)
// 	if !fsutil.CheckExists(todays) {
// 		log.Warn().Msgf("failed to access %s", todays)
// 		err := fsutil.CreateDir(logDir)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	logFile := currentLogFile(todays)
// 	perms := fsutil.CheckPerms(logFile)
// 	if !perms {
// 		log.Error().Bool("Perms", perms).Msgf("failed to access %s", logFile)
// 		return errors.New("failed to access log file")
// 	}
// 	var w io.Writer
// 	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return errors.Wrapf(err, "failed to open %s", logFile)
// 	}
// 	if consoleOut {
// 		w = zerolog.MultiLevelWriter(f, zerolog.NewConsoleWriter())
// 	} else {
// 		w = f
// 	}
// 	log.Logger = zerolog.New(w).With().Timestamp().Caller().Stack().Logger()
// 	// log.Logger = log.Output(zerolog.MultiLevelWriter(zerolog.NewConsoleWriter(), zerolog.New(f).With().Timestamp().Caller().Stack().Logger()))
// 	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
// 	zerolog.TimeFieldFormat = time.RFC3339
// 	zerolog.TimestampFieldName = "time"
// 	zerolog.LevelFieldName = "level"
// 	zerolog.MessageFieldName = "msg"
// 	zerolog.ErrorFieldName = "error"
// 	zerolog.CallerFieldName = "caller"
// 	return nil
// }
