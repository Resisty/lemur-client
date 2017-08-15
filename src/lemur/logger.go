package main

import (
    "io"
    "io/ioutil"
    "os"
    "fmt"
    "github.com/sirupsen/logrus"
)

type logger struct {
    Trace   *logrus.Logger
    Info    *logrus.Logger
    Warning *logrus.Logger
    Error   *logrus.Logger
}

func GetLoggers(traceHandle io.Writer,
                infoHandle io.Writer,
                warningHandle io.Writer,
                errorHandle io.Writer) *(logger) {
    formatter := &logrus.JSONFormatter{
        FieldMap: logrus.FieldMap{
            logrus.FieldKeyTime: "timestamp",
            logrus.FieldKeyLevel: "level",
            logrus.FieldKeyMsg: "message",
        },
    }
	custLogger := logger{
        Trace: &logrus.Logger{
            Out: traceHandle,
            Formatter: formatter,
            Hooks: make(logrus.LevelHooks),
            Level: logrus.DebugLevel,
          },
         Info: &logrus.Logger{
            Out: infoHandle,
            Formatter: formatter,
            Hooks: make(logrus.LevelHooks),
            Level: logrus.InfoLevel,
          },
         Warning: &logrus.Logger{
            Out: warningHandle,
            Formatter: formatter,
            Hooks: make(logrus.LevelHooks),
            Level: logrus.WarnLevel,
          },
         Error: &logrus.Logger{
            Out: errorHandle,
            Formatter: formatter,
            Hooks: make(logrus.LevelHooks),
            Level: logrus.ErrorLevel,
          },
    }
    return &custLogger
}

var Logs = GetLoggers(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

func (log *logger) Tracef(format string, v ...interface{}) {
    defer func() {
        // Rescue ourselves in case Flags hasn't been set up yet but is trying
        // to log
        if r := recover(); r != nil {
            log.Trace.Debug(fmt.Sprintf(format, v...))
        }
    }()
    log.
    Trace.
    WithFields(logrus.
               Fields{"service_name": *Flags.MakoServiceId,
                      "service_environment": *Flags.MakoEnv,
                      "service_version": *Flags.MakoVer}).
    Debug(fmt.Sprintf(format, v...))
}
func (log *logger) Infof(format string, v ...interface{}) {
    defer func() {
        // Rescue ourselves in case Flags hasn't been set up yet but is trying
        // to log
        if r := recover(); r != nil {
            log.Info.Info(fmt.Sprintf(format, v...))
        }
    }()
    log.
    Info.
    WithFields(logrus.
               Fields{"service_name": *Flags.MakoServiceId,
                      "service_environment": *Flags.MakoEnv,
                      "service_version": *Flags.MakoVer}).
    Info(fmt.Sprintf(format, v...))
}
func (log *logger) Warningf(format string, v ...interface{}) {
    defer func() {
        // Rescue ourselves in case Flags hasn't been set up yet but is trying
        // to log
        if r := recover(); r != nil {
            log.Warning.Warning(fmt.Sprintf(format, v...))
        }
    }()
    log.
    Warning.
    WithFields(logrus.
               Fields{"service_name": *Flags.MakoServiceId,
                      "service_environment": *Flags.MakoEnv,
                      "service_version": *Flags.MakoVer}).
    Warning(fmt.Sprintf(format, v...))
}
func (log *logger) Errorf(format string, v ...interface{}) {
    defer func() {
        // Rescue ourselves in case Flags hasn't been set up yet but is trying
        // to log
        if r := recover(); r != nil {
            log.Error.Error(fmt.Sprintf(format, v...))
        }
    }()
    log.
    Error.
    WithFields(logrus.
               Fields{"service_name": *Flags.MakoServiceId,
                      "service_environment": *Flags.MakoEnv,
                      "service_version": *Flags.MakoVer}).
    Error(fmt.Sprintf(format, v...))
}
