package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

func Logger() logr.Logger {
	f := func(prefix, args string) {
		fmt.Printf(
			"[%s] %s — %s",
			time.Now().Format("15:04:05"), strings.ToUpper(prefix), args,
		)
	}
	return funcr.New(f, funcr.Options{})
}
