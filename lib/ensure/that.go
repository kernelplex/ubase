package ensure

import (
	"log/slog"
)

func That(condition bool, message string) {
	if !condition {
		slog.Error("Assertion failed", "message", message)
		panic(message)
	}
}
