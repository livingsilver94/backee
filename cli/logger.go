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
		date := time.Now().Format("15:04:05")
		if prefix != "" {
			fmt.Printf("[%s] %s — %s\n", date, strings.ToUpper(prefix), args)
		} else {
			fmt.Printf("[%s] %s\n", date, args)
		}
	}
	Logger = funcr.New(f, funcr.Options{})
}
