package logger

import (
	"log"
)

func Info(msg string) {

	levelColor := "\x1b[32m"
	resetColor := "\x1b[0m"
	log.Printf("%s[I]%s %s", levelColor, resetColor, msg)
}

func Warning(msg string) {

	levelColor := "\x1b[33m"
	resetColor := "\x1b[0m"
	log.Printf("%s[W]%s %s", levelColor, resetColor, msg)
}

func Error(msg string) {

	levelColor := "\x1b[31m"
	resetColor := "\x1b[0m"
	log.Printf("%s[E]%s %s", levelColor, resetColor, msg)
}

func Fatal(msg string) {
	levelColor := "\x1b[31m"
	resetColor := "\x1b[0m"
	log.Printf("%s[F]%s %s", levelColor, resetColor, msg)
}
