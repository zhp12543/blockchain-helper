package logger

import (
	"fmt"
	"go.uber.org/zap"
)

var (
	logInstance *zap.Logger
)

func init(){
	logInstance = zapLogger()
}

func zapLogger() *zap.Logger {
	logger, err := zap.NewDevelopment(zap.AddCallerSkip(1))
	if err != nil {
		panic(fmt.Sprintf("failed to initialize zap logger: %v", err))
	}
	return logger
}

func GetLogger() *zap.Logger{
	return logInstance
}

func Info(info string){
	logInstance.Info(info)
}

func Debug(info string){
	logInstance.Debug(info)
}

func Error(info string){
	logInstance.Error(info)
}

func Warn(info string){
	logInstance.Warn(info)
}

func Infof(format string, args ...interface{}){
	logInstance.Info(fmt.Sprintf(format, args...))
}

func Debugf(format string, args ...interface{}){
	logInstance.Debug(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...interface{}){
	logInstance.Error(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...interface{}){
	logInstance.Warn(fmt.Sprintf(format, args...))
}

func Fatalln(format string, args ...interface{}) {
	logInstance.Fatal(fmt.Sprintf(format, args...))
}

func Sync(){
	logInstance.Sync()
}