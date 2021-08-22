package slog

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type SloggerMode int

const (
	ModeDevelopment SloggerMode = iota
	ModeProduction
)

type Slogger struct {
	Writer io.Writer
	Mode   SloggerMode
}

func (s *Slogger) NewRequest(writeLogsHook func(sr *SloggerRequest)) (sr *SloggerRequest) {
	return &SloggerRequest{
		s:             s,
		RequestTime:   time.Now(),
		ExtraData:     make([]interface{}, 0),
		Events:        make([]sloggerEvent, 0),
		writeLogsHook: writeLogsHook,
	}
}

type prettyDuration time.Duration

func (pd prettyDuration) MarshalText() (text []byte, err error) {
	t := fmt.Sprintf("%v", time.Duration(pd))
	return []byte(t), nil
}

type SloggerRequest struct {
	s               *Slogger
	ClientHost      string
	HttpMethod      string
	HttpPath        string
	HttpStatusCode  int
	RequestTime     time.Time
	RequestDuration prettyDuration
	ExtraData       []interface{}
	Events          []sloggerEvent
	writeLogsHook   func(sr *SloggerRequest)
}

func (sr *SloggerRequest) appendEvent(severity SloggerEventSeverity, msg string) {
	sr.Events = append(sr.Events, sloggerEvent{
		Severity: severity,
		Message:  msg,
	})
}

func (sr *SloggerRequest) Info(msg string) {
	sr.appendEvent(SeverityInfo, msg)
}

func (sr *SloggerRequest) Warning(msg string) {
	sr.appendEvent(SeverityWarning, msg)
}

func (sr *SloggerRequest) Error(msg string) {
	sr.appendEvent(SeverityError, msg)
}

func (sr *SloggerRequest) WTF(msg string) {
	sr.appendEvent(SeverityWTF, msg)
}

func (sr *SloggerRequest) FormatLog(format string, values ...interface{}) (msg string) {
	return fmt.Sprintf(format, values...)
}

func (sr *SloggerRequest) writeDevelopmentLogs() {
	// TODO: implement a pretty dev logger
	sr.writeJsonLogs()
}

func (sr *SloggerRequest) writeJsonLogs() {
	enc := json.NewEncoder(sr.s.Writer)
	enc.Encode(sr)
}

func (sr *SloggerRequest) WriteLogs() {
	sr.RequestDuration = prettyDuration(time.Now().Sub(sr.RequestTime))
	sr.writeLogsHook(sr)
	switch sr.s.Mode {
	case ModeDevelopment:
		sr.writeDevelopmentLogs()
	case ModeProduction:
		sr.writeJsonLogs()
	}
}

type SloggerEventSeverity string

const (
	SeverityInfo    SloggerEventSeverity = "info"
	SeverityWarning SloggerEventSeverity = "warning"
	SeverityError   SloggerEventSeverity = "error"
	SeverityWTF     SloggerEventSeverity = "wtf"
)

type sloggerEvent struct {
	Severity SloggerEventSeverity
	Message  string
}
