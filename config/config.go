package config

import (
	"strings"

	"github.com/Conflux-Chain/go-conflux-util/viper"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/scroll-tech/rpc-gateway/util/alert"
	"github.com/scroll-tech/rpc-gateway/util/metrics"
	"github.com/sirupsen/logrus"

	// For go-ethereum v1.0.15, node pkg imports internal/debug pkg which will inits log root
	// with `log.GlogHandler`. If we import node pkg from somewhere else, it will override our
	// custom handler defined within function `adaptGethLogger`.
	_ "github.com/ethereum/go-ethereum/node"
)

// Read system enviroment variables prefixed with "INFURA".
// eg., `INFURA_LOG_LEVEL` will override "log.level" config item from the config file.
const viperEnvPrefix = "infura"

func init() {
	// init viper
	viper.MustInit(viperEnvPrefix)
	// init logger
	initLogger()
	// init metrics
	metrics.Init()
	// init alert
	alert.InitDingRobot()
}

func initLogger() {
	var config struct {
		Level      string `default:"info"`
		ForceColor bool
	}
	viper.MustUnmarshalKey("log", &config)

	// set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		logrus.WithError(err).Fatalf("invalid log level configured: %v", config.Level)
	}
	logrus.SetLevel(level)

	if config.ForceColor {
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		})
	}

	// add alert hook for logrus fatal/warn/error level
	hookLevels := []logrus.Level{logrus.FatalLevel, logrus.WarnLevel, logrus.ErrorLevel}
	logrus.AddHook(alert.NewLogrusAlertHook(hookLevels))

	// customize logger here...
	adaptGethLogger()
}

// adaptGethLogger adapt geth logger (which is used by go sdk) to be attached to logrus.
func adaptGethLogger() {
	formatter := log.TerminalFormat(false)

	// geth level => logrus levl
	logrusLevelsMap := map[log.Lvl]logrus.Level{
		log.LvlCrit:  logrus.FatalLevel,
		log.LvlError: logrus.ErrorLevel,
		log.LvlWarn:  logrus.DebugLevel,
		log.LvlInfo:  logrus.DebugLevel,
		log.LvlDebug: logrus.DebugLevel,
		log.LvlTrace: logrus.TraceLevel,
	}

	log.Root().SetHandler(log.FuncHandler(func(r *log.Record) error {
		logLvl, ok := logrusLevelsMap[r.Lvl]
		if !ok {
			return errors.New("unsupported log level")
		}

		if logLvl <= logrus.GetLevel() {
			logStr := string(formatter.Format(r))
			abbrStr := logStr

			firstLineEnd := strings.IndexRune(logStr, '\n')
			if firstLineEnd > 0 { // extract first line as abstract
				abbrStr = logStr[:firstLineEnd]
			}

			logrus.WithField("gethWrappedLogs", logStr).Log(logLvl, abbrStr)
		}

		return nil
	}))
}
