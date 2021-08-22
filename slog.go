package slog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fatih/color"
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
		ExtraData:     make(map[string]interface{}, 0),
		Events:        make([]SloggerEvent, 0),
		writeLogsHook: writeLogsHook,
	}
}

type PrettyDuration time.Duration

func (pd PrettyDuration) MarshalText() (text []byte, err error) {
	return []byte(pd.String()), nil
}

func (pd PrettyDuration) String() string {
	return fmt.Sprintf("%v", time.Duration(pd))
}

type SloggerRequest struct {
	s               *Slogger
	ClientHost      string
	HttpMethod      string
	HttpPath        string
	HttpStatusCode  int
	ResponseSize    int64
	RequestTime     time.Time
	RequestDuration PrettyDuration
	ExtraData       map[string]interface{}
	Events          []SloggerEvent
	writeLogsHook   func(sr *SloggerRequest)
}

func (sr *SloggerRequest) appendEvent(severity SloggerEventSeverity, msg string) {
	sr.Events = append(sr.Events, SloggerEvent{
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

var (
	prettyGET     string
	prettyHEAD    string
	prettyPOST    string
	prettyPUT     string
	prettyDELETE  string
	prettyOPTIONS string
)

func coloriseText(text string, attrs ...color.Attribute) string {
	return color.New(attrs...).Sprint(text)
}

func init() {
	prettyGET = coloriseText("(GET)", color.BgGreen, color.FgBlack)
	prettyHEAD = coloriseText("(HEAD)", color.BgHiGreen, color.FgBlack)
	prettyPOST = coloriseText("(POST)", color.BgMagenta, color.FgBlack)
	prettyPUT = coloriseText("(PUT)", color.BgYellow, color.FgBlack)
	prettyDELETE = coloriseText("(DELETE)", color.BgRed, color.FgBlack)
	prettyOPTIONS = coloriseText("(OPTIONS)", color.BgCyan, color.FgBlack)
}

func colouriseMethod(method string) (output string) {
	method = strings.ToUpper(method)
	switch method {
	case "GET":
		return prettyGET
	case "HEAD":
		return prettyHEAD
	case "POST":
		return prettyPOST
	case "PUT":
		return prettyPUT
	case "DELETE":
		return prettyDELETE
	case "OPTIONS":
		return prettyOPTIONS
	default:
		return coloriseText(method, color.BgWhite, color.FgBlack)
	}
}

var (
	pretty1xx func(int) string
	pretty2xx func(int) string
	pretty3xx func(int) string
	pretty4xx func(int) string
	pretty5xx func(int) string
)

func init() {
	color1xx := color.New(color.BgBlue, color.FgBlack)
	pretty1xx = func(statusCode int) string {
		return color1xx.Sprintf("[%d]", statusCode)
	}
	color2xx := color.New(color.BgGreen, color.FgBlack)
	pretty2xx = func(statusCode int) string {
		return color2xx.Sprintf("[%d]", statusCode)
	}
	color3xx := color.New(color.BgMagenta, color.FgBlack)
	pretty3xx = func(statusCode int) string {
		return color3xx.Sprintf("[%d]", statusCode)
	}
	color4xx := color.New(color.BgYellow, color.FgBlack)
	pretty4xx = func(statusCode int) string {
		return color4xx.Sprintf("[%d]", statusCode)
	}
	color5xx := color.New(color.BgRed, color.FgBlack)
	pretty5xx = func(statusCode int) string {
		return color5xx.Sprintf("[%d]", statusCode)
	}
}

func prettifyStatusCode(statusCode int) string {
	if statusCode >= 100 && statusCode <= 199 {
		return pretty1xx(statusCode)
	} else if statusCode >= 200 && statusCode <= 299 {
		return pretty2xx(statusCode)
	} else if statusCode >= 300 && statusCode <= 399 {
		return pretty3xx(statusCode)
	} else if statusCode >= 400 && statusCode <= 499 {
		return pretty4xx(statusCode)
	} else if statusCode >= 500 && statusCode <= 599 {
		return pretty5xx(statusCode)
	} else {
		return fmt.Sprintf("[%d]", statusCode)
	}
}

var prettySeverity map[SloggerEventSeverity]string

func init() {
	prettySeverity = make(map[SloggerEventSeverity]string)
	prettySeverity[SeverityInfo] = color.New(color.BgCyan, color.FgBlack).Sprint("<INFO>")
	prettySeverity[SeverityWarning] = color.New(color.BgYellow, color.FgBlack).Sprint("<WARNING>")
	prettySeverity[SeverityError] = color.New(color.BgRed, color.FgBlack).Sprint("<ERROR>")
	prettySeverity[SeverityWTF] = color.New(color.BgMagenta, color.FgBlack).Sprint("<WTF>")
}

func formatFileSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func (sr *SloggerRequest) writeDevelopmentLogs() {
	buf := bytes.NewBufferString(fmt.Sprintf("%s | \"%s\" | %s | %s | %s | %s | %s | %s\n",
		colouriseMethod(sr.HttpMethod),
		sr.HttpPath,
		sr.ClientHost,
		prettifyStatusCode(sr.HttpStatusCode),
		formatFileSize(sr.ResponseSize),
		sr.RequestTime.Format("2 Jan 2006 3:04:05 PM"),
		sr.RequestDuration,
		fmt.Sprintf("%v", sr.ExtraData)[3:],
	))
	for _, ev := range sr.Events {
		buf.WriteString(fmt.Sprintf(" |-> %s: %s\n", prettySeverity[ev.Severity], ev.Message))
	}

	sr.s.Writer.Write(buf.Bytes())
}

func (sr *SloggerRequest) writeJsonLogs() {
	enc := json.NewEncoder(sr.s.Writer)
	enc.Encode(sr)
}

func (sr *SloggerRequest) WriteLogs() {
	sr.RequestDuration = PrettyDuration(time.Now().Sub(sr.RequestTime))
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

type SloggerEvent struct {
	Severity SloggerEventSeverity
	Message  string
}
