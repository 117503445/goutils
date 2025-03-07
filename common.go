package goutils

import (
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
