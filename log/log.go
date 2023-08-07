package log

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
	"golang.org/x/exp/slog"
)

// Handler is a slog handler with a focus on readability and aesthetics,
// although it sacrifices parsability a little.
type Handler struct {
	dest *bufio.Writer
	// group is a string idenfying a particular context while logging.
	group string
	// attribs is a collection of default attributes to be logged.
	attribs []slog.Attr

	opts Options
}

func NewHandler(dest io.Writer, op *Options) Handler {
	var opts Options
	if op != nil {
		opts = *op
	} else {
		opts = DefaultOptions()
	}
	return Handler{
		dest: bufio.NewWriter(dest),
		opts: opts,
	}
}

// Enabled implements slog.Handler's Enabled function.
func (h Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level
}

// Handle implements slog.Handler's Handle function.
func (h Handler) Handle(_ context.Context, rec slog.Record) error {
	err := h.printPrefix(rec.Time)
	if err != nil {
		return err
	}
	err = h.printMessage(rec.Level, rec.Message)
	if err != nil {
		return err
	}
	err = h.printAttributes(rec)
	if err != nil {
		return err
	}
	err = h.dest.WriteByte('\n')
	if err != nil {
		return err
	}
	return h.dest.Flush()
}

// WithAttrs implements slog.Handler's WithAttrs function.
func (h Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return Handler{
		dest:    h.dest,
		group:   h.group,
		attribs: append(h.attribs, attrs...),
		opts:    h.opts,
	}
}

// WithGroup implements slog.Handler's WithGroup function.
func (h Handler) WithGroup(group string) slog.Handler {
	return Handler{
		dest:    h.dest,
		group:   group,
		attribs: h.attribs,
		opts:    h.opts,
	}
}

func (h Handler) printPrefix(t time.Time) error {
	var err error
	tim := t.Format("15:04:05")
	if h.group != "" {
		_, err = fmt.Fprintf(h.dest, "[%s] %s — ", tim, h.color(color.Bold).Sprint(h.group))
	} else {
		_, err = fmt.Fprintf(h.dest, "[%s] ", tim)
	}
	return err
}

func (h Handler) printMessage(level slog.Level, message string) error {
	if message == "" {
		return nil
	}
	var err error
	if level == slog.LevelError {
		_, err = h.color(color.FgRed).Fprintf(h.dest, "%s: %s", level, message)
	} else {
		_, err = fmt.Fprint(h.dest, message)
	}
	return err
}

func (h Handler) printAttributes(r slog.Record) error {
	if r.NumAttrs() == 0 {
		return nil
	}
	err := h.printSeparator()
	if err != nil {
		return err
	}
	r.Attrs(func(attr slog.Attr) bool {
		if attr.Equal(slog.Attr{}) {
			return true
		}
		_, err = fmt.Fprintf(h.dest, "%s: %s ", attr.Key, attr.Value)
		return err == nil
	})
	return err
}

// printSeparator prints a separator between pieces of information.
func (h Handler) printSeparator() error {
	_, err := fmt.Fprint(h.dest, " | ")
	return err
}

func (h Handler) color(attrs ...color.Attribute) *color.Color {
	var col *color.Color
	if h.opts.Colored {
		col = color.New(attrs...)
	} else {
		col = color.New()
		col.DisableColor()
	}
	return col
}

// Options sets the behavior of Handler.
type Options struct {
	// Level is the granularity of Handler.
	Level slog.Level
	// Colored specifies whether Handler will output ANSI colors.
	// true actually means "auto", in that colors are disabled when
	// the output is not a terminal.
	Colored bool
}

func DefaultOptions() Options {
	return Options{
		Level:   slog.LevelInfo,
		Colored: true,
	}
}
