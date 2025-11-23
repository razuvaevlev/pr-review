package logging

import (
	"fmt"
	"log/slog"
	"os"
)

func Printf(format string, a ...any) {
	slog.Default().Info(fmt.Sprintf(format, a...))
}

func Fatalf(format string, a ...any) {
	slog.Default().Error(fmt.Sprintf(format, a...))
	os.Exit(1)
}
