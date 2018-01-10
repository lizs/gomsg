package gomsg

import (
	"log"
	"os"
	"path"
	"strings"
)

var (
	// Logger logger instance
	Logger *log.Logger
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

// NewLogger create logger
func NewLogger(name string) {
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
	file, err := os.Create(fullName)
	if err != nil {
		log.Fatalln(err)
	}

	Logger = log.New(file, "", log.LstdFlags|log.Llongfile)
}

// Recover recover tool function
func Recover() {
	if e := recover(); e != nil {
		log.Println(e)
		if Logger != nil {
			Logger.Println(e)
		}
	}
}
