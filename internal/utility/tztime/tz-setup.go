package tztime

import (
	"go-api-boilerplate/config"
	"os"
	"strings"
	"time"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

// SetupTimezone
// ini init function untuk setup timezone service golang globally
// prioritas timezone ::
// 1. env variable TZ
// 2. baca settingan dari etc/loadlocation
// 3. appsettings.ini Timezone key
// 4. kalo suram, kembali ke UTC ~ :(
func SetupTimezone(logger *log.Logger, cfg *config.Config) {
	tzName, tzSource := func() (string, string) {
		if cfg.IsDevelopment {
			return cfg.Timezone, "appsettings.ini [Timezone] (development mode)"
		}
		return resolveTimezone(cfg)
	}()

	if tzName == "" {
		logger.Warn("No timezone configured — using UTC as default")
		return
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil {
		logger.Warnf("Invalid timezone %q (from %s): %v — falling back to UTC", tzName, tzSource, err)
		return
	}

	time.Local = loc
	logger.Infof("Timezone set to %q (source: %s)", tzName, tzSource)
}

func resolveTimezone(cfg *config.Config) (tzName, tzSource string) {
	if tz := os.Getenv("TZ"); tz != "" {
		return tz, "TZ environment variable"
	}

	if tz := readLinuxLocaltimeSymlink(); tz != "" {
		return tz, "/etc/localtime symlink"
	}

	if cfg.Timezone != "" {
		return cfg.Timezone, "appsettings.ini [Timezone]"
	}

	return "", ""
}

func readLinuxLocaltimeSymlink() string {
	symlinkTarget, err := os.Readlink("/etc/localtime")
	if err != nil {
		return ""
	}

	parts := strings.SplitN(symlinkTarget, "zoneinfo/", 2)
	if len(parts) < 2 {
		return ""
	}

	return parts[1]
}
