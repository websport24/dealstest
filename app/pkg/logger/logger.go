package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// NewLogger создает новый настроенный логгер
func NewLogger() *logrus.Logger {
	logger := logrus.New()

	// Настройка вывода
	logger.SetOutput(os.Stdout)

	// Настройка формата
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Уровень логирования по умолчанию
	logger.SetLevel(logrus.InfoLevel)

	return logger
}

// NewDevelopmentLogger создает логгер для разработки с человекочитаемым форматом
func NewDevelopmentLogger() *logrus.Logger {
	logger := logrus.New()

	// Настройка вывода
	logger.SetOutput(os.Stdout)

	// Настройка формата для разработки
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     true,
	})

	// Уровень логирования для разработки
	logger.SetLevel(logrus.DebugLevel)

	return logger
}
