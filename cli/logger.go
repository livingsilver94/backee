package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/go-logr/logr"
)

// LogLevel sets the verbosity of a logger.
type LogLevel int

const (
	// LogError is the least verbose level of a logger. It only logs errors.
	LogError LogLevel = 1
	// LogInfo logs information and error messages.
	LogInfo LogLevel = 2
)

// NewLogger creates a new structured logger.
// colored regulates whether the logger will use colored output.
func NewLogger(level LogLevel, colored bool) logr.Logger {
	sink := logSink{
		writer:  bufio.NewWriter(os.Stdout),
		level:   level,
		colored: colored,
	}
	return logr.New(sink)
}

// logSink is our custom logr.LogSink implementation.
// It aims to be a structured logger with decent aesthetics to human readers,
// although it sacrifices parsability a little.
type logSink struct {
	name    string
	keyVals map[string]interface{}
	writer  *bufio.Writer
	level   LogLevel
	colored bool
}

// Init implements logr.LogSink's Init method.
func (l logSink) Init(info logr.RuntimeInfo) {}

// Enabled implements logr.LogSink's Enabled method.
func (l logSink) Enabled(level int) bool {
	return l.level >= LogLevel(level)
}

// Info implements logr.LogSink's Info method.
func (l logSink) Info(level int, msg string, keyVals ...interface{}) {
	if !l.Enabled(level) {
		return
	}
	validateKeyVals(keyVals...)
	l.printLine(nil, msg, keyVals...)
}

// Error implements logr.LogSink's Error method.
func (l logSink) Error(err error, msg string, keyVals ...interface{}) {
	validateKeyVals(keyVals...)
	l.printLine(err, msg, keyVals...)
}

// WithValues implements logr.LogSink's WithValues method.
func (l logSink) WithValues(keyVals ...interface{}) logr.LogSink {
	validateKeyVals(keyVals...)
	kvMap := make(map[string]interface{}, len(l.keyVals)+len(keyVals)/2)
	for k, v := range l.keyVals {
		kvMap[k] = v
	}
	for i := 0; i < len(keyVals); i += 2 {
		kvMap[keyVals[i].(string)] = keyVals[i+1]
	}
	copy := l
	copy.keyVals = kvMap
	return copy
}

// WithName implements logr.LogSink's WithName method.
func (l logSink) WithName(name string) logr.LogSink {
	var newName string
	if l.name == "" {
		newName = strings.ToUpper(name)
	} else {
		newName = l.name + "." + strings.ToUpper(name)
	}
	copy := l
	copy.name = newName
	return copy
}

func validateKeyVals(keyVals ...interface{}) {
	if len(keyVals)%2 != 0 {
		panic("odd number of key/value arguments passed")
	}
}

func (l logSink) printLine(err error, msg string, keyVals ...interface{}) {
	l.printPrefix()
	if err != nil {
		l.color(color.FgRed).Fprintf(l.writer, "ERROR: %s", err)
		if msg != "" {
			l.printSeparator()
		}
	}
	if msg != "" {
		fmt.Fprint(l.writer, msg)
		if len(l.keyVals) != 0 || len(keyVals) != 0 {
			l.printSeparator()
		}
	}
	for k, v := range l.keyVals {
		fmt.Fprintf(l.writer, "%s: %+v  ", k, v)
	}
	for i := 0; i < len(keyVals); i += 2 {
		fmt.Fprintf(l.writer, "%s: %+v  ", keyVals[i], keyVals[i+1])
	}
	fmt.Fprint(l.writer, "\n")
	l.writer.Flush()
}

func (l logSink) printPrefix() {
	date := time.Now().Format("15:04:05")
	if l.name != "" {
		fmt.Fprintf(l.writer, "[%s] %s ??? ", date, l.color(color.Bold).Sprint(l.name))
	} else {
		fmt.Fprintf(l.writer, "[%s] ", date)
	}
}

// printSeparator prints a separator between pieces of information.
func (l logSink) printSeparator() {
	fmt.Fprint(l.writer, " | ")
}

func (l logSink) color(attrs ...color.Attribute) *color.Color {
	c := color.New(attrs...)
	if !l.colored {
		c.DisableColor()
	}
	return c
}
