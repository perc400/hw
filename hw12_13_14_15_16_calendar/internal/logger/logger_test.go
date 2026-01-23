package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {
	t.Run("debug_level", func(t *testing.T) {
		logg, err := New("INFO")
		require.NoError(t, err)
		require.Equal(t, zapcore.InfoLevel, logg.logger.Level())

		logg.Error("error message")
		logg.Warn("warn message")
		logg.Info("info message")

		logg.SetLevel("DEBUG")
		logg.Debug("debug message")

		require.Equal(t, zapcore.DebugLevel, logg.logger.Level())
	})

	t.Run("change_level", func(t *testing.T) {
		logg, err := New("INFO")
		require.NoError(t, err)

		logg.SetLevel("ERROR")
		require.Equal(t, zapcore.ErrorLevel, logg.logger.Level())
	})
}
