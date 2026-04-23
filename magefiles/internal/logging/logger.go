package logging

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func New() *slog.Logger {
	handler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelInfo,
		TimeFormat: time.DateTime,
		NoColor:    os.Getenv("NO_COLOR") != "" || os.Getenv("CI") != "",
		//AddSource:  true,
	})

	return slog.New(handler)
}
