package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

var Logger logr.Logger

func init() {
	f := func(prefix, args string) {
		fmt.Printf(
			"[%s] %s — %s",
			time.Now().Format("15:04:05"), strings.ToUpper(prefix), args,
		)
	}
	Logger = funcr.New(f, funcr.Options{})
}
