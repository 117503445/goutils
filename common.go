package goutils

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// TimeStrSec returns the time format string, like 20240915.221219
func TimeStrSec() string {
	return time.Now().Format("20060102.150405")
}

// TimeStrMilliSec returns the time format string with millisecond, like 20240915.221219.123
func TimeStrMilliSec() string {
	return time.Now().Format("20060102.150405.000")
}

func DurationToStr(d time.Duration) string {
	sec := d.Seconds()

	if sec < 1 {
		// 对于小于 1s 的时间，保留一位小数
		// 例子: 994.2ms, 5.4ms
		// 但是不要出现 .0，例子: 10ms，不要出现 10.0ms
		ms := sec * 1000
		s := fmt.Sprintf("%.1fms", ms)
		return strings.Replace(s, ".0ms", "ms", 1)
	} else if sec < 60 {
		// 对于小于 1m 的时间，保留一位小数
		// 例子: 994.2s, 5.4s
		// 但是不要出现 .0，例子: 10s，不要出现 10.0s
		s := fmt.Sprintf("%.1fs", sec)
		return strings.Replace(s, ".0s", "s", 1)
	} else if sec < 3600 {
		// 对于小于 1h 的时间，保留一位小数
		// 例子: 12m14.4s，59m59.5s
		// 但是不要出现 .0s，例子: 10m4s, 不要出现 10m4.0s
		// 不要出现 0s，例子: 10m，不要出现 10m0s
		m := int(sec / 60)
		sRemaining := sec - float64(m)*60
		parts := []string{fmt.Sprintf("%dm", m)}
		if sRemaining > 0 {
			s := fmt.Sprintf("%.1fs", sRemaining)
			s = strings.Replace(s, ".0s", "s", 1)
			parts = append(parts, s)
		}
		return strings.Join(parts, "")
	} else {
		// 对于大于 1h 的时间，保留一位小数
		// 例子: 12h14m14.4s，59h59m59.5s
		// 但是不要出现 .0s，例子: 10h4m4s, 不要出现 10h4m4.0s
		// 不要出现 0s，例子: 12h10m，不要出现 12h10m0s
		// 不要出现 0m0s，例子: 12h，不要 出现 12h0m0s
		// 但是 12h0m12s 是合法的
		h := int(sec / 3600)
		remaining := sec - float64(h)*3600
		m := int(remaining / 60)
		sRemaining := remaining - float64(m)*60
		parts := []string{fmt.Sprintf("%dh", h)}
		if m > 0 || sRemaining > 0 {
			parts = append(parts, fmt.Sprintf("%dm", m))
			if sRemaining > 0 {
				s := fmt.Sprintf("%.1fs", sRemaining)
				s = strings.Replace(s, ".0s", "s", 1)
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, "")
	}
}

func UUID4() string {
	return uuid.New().String()
}

func UUID7() string {
	uuid, err := uuid.NewV7()
	if err != nil {
		log.Fatal().Err(err).Msg("UUID7 failed")
	}
	return uuid.String()
}
