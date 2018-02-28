/*
  btcbot is a Bitcoin trading bot for HUOBI.com written
  in golang, it features multiple trading methods using
  technical analysis.

  Disclaimer:

  USE AT YOUR OWN RISK!

  The author of this project is NOT responsible for any damage or loss caused
  by this software. There can be bugs and the bot may not perform as expected
  or specified. Please consider testing it first with paper trading /
  backtesting on historical data. Also look at the code to see what how
  it's working.

  Weibo:http://weibo.com/bocaicfa
*/

package logger

import (
	"fmt"
	"gbird/config"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	colorRed = uint8(iota + 91)
	colorGreen
	colorYellow
	colorBlue
	colorMagenta //洋红
)

type logger struct {
	*log.Logger
}
type logType struct {
	Tag     string
	Logfile string
	Color   uint8
}

var logTypes = map[string]logType{
	"info":  logType{Tag: "INFO", Logfile: "info.log", Color: colorBlue},
	"debug": logType{Tag: "DEBUG", Logfile: "debug.log", Color: colorGreen},
	"error": logType{Tag: "ERROR", Logfile: "error.log", Color: colorRed},
	"fatal": logType{Tag: "FATAL", Logfile: "fatal.log", Color: colorMagenta},
}

func init() {
	os.Mkdir(config.ROOT+"/_log/", 0777)
}

func getFileName(source string) string {
	t := time.Now()
	path := fmt.Sprintf("%s/_log/%4d%0.2d%0.2d/", config.ROOT, t.Year(), t.Month(), t.Day())
	_, err := os.Stat(path)
	if err != nil {
		os.Mkdir(path, 0777)
	}
	if source == "trade.csv" || source == "error.log" || source == "fatal.log" {
		return fmt.Sprintf("%s%s", path, source)
	}
	return fmt.Sprintf("%s%2d_%s", path, t.Hour(), source)
}
func new(out io.Writer) *logger {
	return &logger{
		Logger: log.New(out, "", log.LstdFlags),
	}
}

//Infof infof
func Infof(format string, args ...interface{}) {
	corePrintf(logTypes["info"], format, args...)
}

//Infoln 一行
func Infoln(args ...interface{}) {

	corePrintln(logTypes["info"], args...)
}

//Errorf f
func Errorf(format string, args ...interface{}) {
	corePrintf(logTypes["error"], format, args...)
}

//Errorln ln
func Errorln(args ...interface{}) {
	corePrintln(logTypes["error"], args...)
}

//Fatalf f
func Fatalf(format string, args ...interface{}) {
	// 加上文件调用和行号
	_, callerFile, line, ok := runtime.Caller(1)
	if ok {
		args = append([]interface{}{"[", filepath.Base(callerFile), "]", line}, args...)
	}
	corePrintf(logTypes["fatal"], format, args...)
}

//Fatalln ln
func Fatalln(args ...interface{}) {
	// 加上文件调用和行号
	_, callerFile, line, ok := runtime.Caller(1)
	if ok {
		args = append([]interface{}{"[", filepath.Base(callerFile), "]", line}, args...)
	}
	corePrintln(logTypes["fatal"], args...)
}

//Debugf f
func Debugf(format string, args ...interface{}) {
	if config.Config["debug"] == "true" {
		corePrintf(logTypes["debug"], format, args...)
	}
}

//Debugln ln
func Debugln(args ...interface{}) {
	if config.Config["debug"] == "true" {
		corePrintln(logTypes["debug"], args...)
	}
}

func corePrintf(ty logType, format string, args ...interface{}) {
	file, err := os.OpenFile(getFileName(ty.Logfile), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	aa := append([]interface{}{"[" + ty.Tag + "]"}, args...)
	new(file).Printf(format, aa...)

	if config.Config["logconsole"] == "1" {
		format=time.Now().Format("2006/01/02 15:04:05")+" [" + ty.Tag + "] "+format
		lg := fmt.Sprintf("\x1b[%dm%s\x1b[0m", ty.Color, fmt.Sprintf(format, args...))
		fmt.Println(lg)
	}
}

func corePrintln(ty logType, args ...interface{}) {
	file, err := os.OpenFile(getFileName(ty.Logfile), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	aa := append([]interface{}{"[" + ty.Tag + "]"}, args...)
	new(file).Println(aa...)

	if config.Config["logconsole"] == "1" {
		as := []interface{}{time.Now().Format("2006/01/02 15:04:05") + " [" + ty.Tag + "] "}
		as = append(as, args...)
		lg := fmt.Sprintf("\x1b[%dm%s\x1b[0m", ty.Color, fmt.Sprint(as...))
		fmt.Println(lg)
	}
}
