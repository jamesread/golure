package listenaddr

import (
	"net"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Resolve returns the address a production service should listen on. A valid
// $PORT wins, then configured, then fallback.
//
// $PORT may be a bare port, which keeps the host of configured (or of fallback
// if configured is empty), or a full address such as "127.0.0.1:8080". An
// invalid $PORT is logged and ignored. So is any $PORT when configured is set
// but malformed, because the host to keep cannot be determined.
//
// Unlike AvailableListenAddr, Resolve never tests or changes the port, so a
// port conflict fails at startup instead of moving the service.
func Resolve(configured, fallback string) string {
	configured = strings.TrimSpace(configured)

	base := configured

	if base == "" {
		base = fallback
	}

	raw := strings.TrimSpace(os.Getenv("PORT"))

	if raw == "" {
		return base
	}

	if strings.Contains(raw, ":") {
		if _, port, err := net.SplitHostPort(raw); err != nil || !validPort(port) {
			log.WithField("PORT", raw).Error("Ignoring invalid PORT environment variable")
			return base
		}

		return raw
	}

	if !validPort(raw) {
		log.WithField("PORT", raw).Error("Ignoring invalid PORT environment variable")
		return base
	}

	host, _, err := net.SplitHostPort(base)

	if err != nil {
		log.WithFields(log.Fields{
			"PORT":   raw,
			"listen": base,
		}).Error("Ignoring PORT environment variable because the listen address is invalid")

		return base
	}

	return net.JoinHostPort(host, raw)
}

func validPort(raw string) bool {
	port, err := strconv.Atoi(raw)

	return err == nil && port >= 1 && port <= 65535
}
