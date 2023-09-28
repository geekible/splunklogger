package splunklogger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/efimovalex/stackerr"
)

type SplunkLogger struct {
	splunkTokenValue string
	splunkEndpoint   string
	splunkPort       int
}

type LogLevel int

const (
	_ LogLevel = iota
	DEBUG
	INFORMATION
	WARNING
	ERROR
	FATAL
)

func logLevelToString(logLevel LogLevel) string {
	switch logLevel {
	case DEBUG:
		return "debug"
	case INFORMATION:
		return "information"
	case WARNING:
		return "warning"
	case ERROR:
		return "error"
	case FATAL:
		return "fatal"
	default:
		return "Not Set"
	}
}

type wrapper struct {
	Event LogMessage `json:"event"`
}

type LogMessage struct {
	Message             string    `json:"message"`
	StackTrace          string    `json:"stack_trace,omitempty"`
	LogLevel            LogLevel  `json:"log_level,omitempty"`
	LogLevelDescription string    `json:"log_level_description,omitempty"`
	EventTime           time.Time `json:"event_time"`
}

func NewSplunkLogger(splunkToken, splunkEndpoint string, splunkPort int) *SplunkLogger {
	return &SplunkLogger{
		splunkTokenValue: splunkToken,
		splunkEndpoint:   splunkEndpoint,
		splunkPort:       splunkPort,
	}
}

func (sp *SplunkLogger) logToSplunk(lm LogMessage) {
	url := fmt.Sprintf("%s:%d/services/collector/event", sp.splunkEndpoint, sp.splunkPort)

	payload := wrapper{Event: lm}
	payloadJson, err := json.Marshal(payload)
	if err != nil {

		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadJson))
	if err != nil {

		return
	}
	request.Header.Add("Authorization", fmt.Sprintf("Splunk %s", sp.splunkTokenValue))
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {

		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 299 {
		body, _ := io.ReadAll(response.Body)
		fmt.Println(string(body))
	}
}

func (sp *SplunkLogger) LogDebug(message, source_filename string, lineNo int) {
	lm := LogMessage{
		Message:             message,
		StackTrace:          fmt.Sprintf("Debug Stacktrace:\n-> %s", fmt.Sprintf("%s:%d", source_filename, lineNo)),
		LogLevel:            DEBUG,
		LogLevelDescription: logLevelToString(DEBUG),
		EventTime:           time.Now(),
	}
	sp.logToSplunk(lm)
}

func (sp *SplunkLogger) LogInformation(message, source_filename string, lineNo int) {
	lm := LogMessage{
		Message:             message,
		StackTrace:          fmt.Sprintf("Information Stacktrace:\n-> %s", fmt.Sprintf("%s:%d", source_filename, lineNo)),
		LogLevel:            INFORMATION,
		LogLevelDescription: logLevelToString(INFORMATION),
		EventTime:           time.Now(),
	}
	sp.logToSplunk(lm)
}

func (sp *SplunkLogger) LogWarning(message, source_filename string, lineNo int) {
	lm := LogMessage{
		Message:             message,
		StackTrace:          fmt.Sprintf("Warning Stacktrace:\n-> %s", fmt.Sprintf("%s:%d", source_filename, lineNo)),
		LogLevel:            WARNING,
		LogLevelDescription: logLevelToString(WARNING),
		EventTime:           time.Now(),
	}
	sp.logToSplunk(lm)
}

func (sp *SplunkLogger) LogError(err error) {
	e := stackerr.NewFromError(err)
	lm := LogMessage{
		Message:             err.Error(),
		StackTrace:          e.Stack().Sprint(),
		LogLevel:            ERROR,
		LogLevelDescription: logLevelToString(ERROR),
		EventTime:           time.Now(),
	}
	sp.logToSplunk(lm)
}

func (sp *SplunkLogger) LogFatal(err error) {
	e := stackerr.NewFromError(err)
	lm := LogMessage{
		Message:             err.Error(),
		StackTrace:          e.Stack().Sprint(),
		LogLevel:            FATAL,
		LogLevelDescription: logLevelToString(FATAL),
		EventTime:           time.Now(),
	}
	sp.logToSplunk(lm)
}
