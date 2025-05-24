package goutils_test

import (
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"

	"github.com/117503445/goutils"
)

func TestCommon(t *testing.T) {
	goutils.InitZeroLog(goutils.WithNoColor{})

	ast := assert.New(t)
	ast.NotEmpty(goutils.TimeStrSec())

	log.Debug().Str("TimeStrSec", goutils.TimeStrSec()).Str("TimeStrMilliSec", goutils.TimeStrMilliSec()).Msg("Time")

	log.Debug().Str("UUID4", goutils.UUID4()).Send()
	log.Debug().Str("UUID7", goutils.UUID7()).Send()

	dir, err := goutils.FindGitRepoRoot()
	ast.NoError(err)
	log.Debug().Str("GitRepoRoot", dir).Msg("GitRepoRoot")

}
func TestDurationToStr(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		// 毫秒测试（<1秒）
		{
			name:     "500ms",
			input:    500 * time.Millisecond,
			expected: "500ms",
		},
		{
			name:     "999.9ms",
			input:    999*time.Millisecond + 900*time.Microsecond,
			expected: "999.9ms",
		},
		{
			name:     "0.5ms",
			input:    500 * time.Microsecond, // 0.5ms
			expected: "0.5ms",
		},

		// 秒测试（1秒~1分钟）
		{
			name:     "1s",
			input:    1 * time.Second,
			expected: "1s",
		},
		{
			name:     "59.9s",
			input:    59*time.Second + 900*time.Millisecond,
			expected: "59.9s",
		},
		{
			name:     "30s",
			input:    30 * time.Second,
			expected: "30s",
		},

		// 分钟测试（1分钟~1小时）
		{
			name:     "1m",
			input:    1 * time.Minute,
			expected: "1m",
		},
		{
			name:     "1m5s",
			input:    65 * time.Second,
			expected: "1m5s",
		},
		{
			name:     "59m59.9s",
			input:    (59*60+59)*time.Second + 900*time.Millisecond,
			expected: "59m59.9s",
		},
		{
			name:     "30m0s", // 确保秒部分为0时被省略
			input:    30 * time.Minute,
			expected: "30m",
		},

		// 小时测试（≥1小时）
		{
			name:     "1h",
			input:    1 * time.Hour,
			expected: "1h",
		},
		{
			name:     "1h5m",
			input:    1*time.Hour + 5*time.Minute,
			expected: "1h5m",
		},
		{
			name:     "1h0m5s",
			input:    1*time.Hour + 5*time.Second,
			expected: "1h0m5s",
		},
		{
			name:     "12h10m",
			input:    12*time.Hour + 10*time.Minute,
			expected: "12h10m",
		},
		{
			name:     "12h0m0s", // 分钟和秒都为0时只保留小时
			input:    12 * time.Hour,
			expected: "12h",
		},
		{
			name:     "12h0m5s",
			input:    12*time.Hour + 5*time.Second,
			expected: "12h0m5s",
		},
		{
			name:     "25h30m15.5s", // 复杂组合测试
			input:    25*time.Hour + 30*time.Minute + 15*time.Second + 500*time.Millisecond,
			expected: "25h30m15.5s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, goutils.DurationToStr(tt.input))
		})
	}
}
