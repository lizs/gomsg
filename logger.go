package gomsg

import (
	"log"
	"os"
	"path"
	"runtime/debug"
	"strings"
)

var (
	// Logger logger instance
	Logger *log.Logger
	_file  *os.File
)

func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// NewLog create logger
func NewLog(name string) {
	if len(name) == 0 {
		log.Fatalln("logger name not specified.")
	}

	if !strings.HasSuffix(name, ".log") {
		name += ".log"
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	logPath := path.Join(wd, "log")
	if exists, _ := pathExist(logPath); !exists {
		os.MkdirAll(logPath, os.ModePerm)
	}

	fullName := path.Join(logPath, name)
	_file, err := os.Create(fullName)
	if err != nil {
		log.Fatalln(err)
	}

	Logger = log.New(_file, "", log.LstdFlags|log.Llongfile)
	log.Printf("Logger [%s] created", fullName)
}

// CloseLog close underline log file
func CloseLog() {
	_file.Close()
}

// Recover recover tool function
func Recover() {
	if e := recover(); e != nil {
		log.Println(debug.Stack())
		if Logger != nil {
			Logger.Println(debug.Stack())
		}
	}
}
