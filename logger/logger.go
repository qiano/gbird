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
	. "gbird/config"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	info_file     = "info.log"
	debug_file    = "debug.log"
	trace_file    = "trace.log"
	error_file    = "error.log"
	fatal_file    = "fatal.log"
	backtest_file string
)

func GetFileName(source string) string {
	t := time.Now()
	path := fmt.Sprintf("%s/_log/%4d%0.2d%0.2d/", ROOT, t.Year(), t.Month(), t.Day())
	_, err := os.Stat(path)
	if err != nil {
		os.Mkdir(path, 0777)
	}
	if source == "trade.csv" || source == "error.log" || source == "fatal.log" {
		return fmt.Sprintf("%s%s", path, source)
	}
	return fmt.Sprintf("%s%2d_%s", path, t.Hour(), source)
}

func init() {
	os.Mkdir(ROOT+"/_log/", 0777)
}

type logger struct {
	*log.Logger
}

func New(out io.Writer) *logger {
	return &logger{
		Logger: log.New(out, "", log.LstdFlags),
	}
}

func NewReport(out io.Writer) *logger {
	return &logger{
		Logger: log.New(out, "", log.LstdFlags),
	}
}

func Infof(format string, args ...interface{}) {
	file, err := os.OpenFile(GetFileName(info_file), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	New(file).Printf(format, args...)
	if Config["infoconsole"] == "1" {
		log.Printf(format, args...)
	}
}

func Infoln(args ...interface{}) {
	file, err := os.OpenFile(GetFileName(info_file), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	New(file).Println(args...)
	if Config["infoconsole"] == "1" {
		as := []interface{}{"[INFO]"}
		for _, v := range args {
			as = append(as, v)
		}
		log.Println(as...)
	}
}

func Errorf(format string, args ...interface{}) {
	file, err := os.OpenFile(GetFileName(error_file), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	New(file).Printf(format, args...)
	if Config["errorconsole"] == "1" {
		log.Printf(format, args...)
	}
}

func Errorln(args ...interface{}) {
	file, err := os.OpenFile(GetFileName(error_file), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	// 加上文件调用和行号
	_, callerFile, line, ok := runtime.Caller(1)
	if ok {
		args = append([]interface{}{"[", filepath.Base(callerFile), "]", line}, args...)
	}
	New(file).Println(args...)
	if Config["infoconsole"] == "1" {
		as := []interface{}{"[ERROR]"}
		for _, v := range args {
			as = append(as, v)
		}
		log.Println(as...)
	}
}

func Fatalf(format string, args ...interface{}) {
	file, err := os.OpenFile(GetFileName(fatal_file), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	New(file).Printf(format, args...)
	if Config["fatalconsole"] == "1" {
		log.Printf(format, args...)
	}
}

func Fatalln(args ...interface{}) {
	file, err := os.OpenFile(GetFileName(fatal_file), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	// 加上文件调用和行号
	_, callerFile, line, ok := runtime.Caller(1)
	if ok {
		args = append([]interface{}{"[", filepath.Base(callerFile), "]", line}, args...)
	}
	New(file).Println(args...)
	if Config["infoconsole"] == "1" {
		as := []interface{}{"[FATAL]"}
		for _, v := range args {
			as = append(as, v)
		}
		log.Println(as...)
	}
}

func Fatal(args ...interface{}) {
	file, err := os.OpenFile(GetFileName(fatal_file), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	// 加上文件调用和行号
	_, callerFile, line, ok := runtime.Caller(1)
	if ok {
		args = append([]interface{}{"[", filepath.Base(callerFile), "]", line}, args...)
	}
	New(file).Println(args...)
	if Config["fatalconsole"] == "1" {
		log.Println(args...)
	}
}

func Debugf(format string, args ...interface{}) {
	if Config["debug"] == "1" {
		file, err := os.OpenFile(GetFileName(debug_file), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return
		}
		defer file.Close()
		New(file).Printf(format, args...)

		if Config["debugconsole"] == "1" {
			log.Printf(format, args...)
		}
	}
}

func Debugln(args ...interface{}) {
	if Config["debug"] == "1" {
		file, err := os.OpenFile(GetFileName(debug_file), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return
		}
		defer file.Close()
		// 加上文件调用和行号
		_, callerFile, line, ok := runtime.Caller(1)
		if ok {
			args = append([]interface{}{"[", filepath.Base(callerFile), "]", line}, args...)
		}
		New(file).Println(args...)
		if Config["debugconsole"] == "1" {
			log.Println(args...)
		}
	}
}

func Tracef(format string, args ...interface{}) {
	file, err := os.OpenFile(GetFileName(trace_file), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	New(file).Printf(format, args...)
}

func Traceln(args ...interface{}) {
	file, err := os.OpenFile(GetFileName(trace_file), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	// 加上文件调用和行号
	_, callerFile, line, ok := runtime.Caller(1)
	if ok {
		args = append([]interface{}{"[", filepath.Base(callerFile), "]", line}, args...)
	}
	New(file).Println(args...)
}
