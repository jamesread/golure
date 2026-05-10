package listenaddr

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"strconv"
	"strings"
)

// AvailableListenAddr returns a TCP address string suitable for net/http
// (for example ":8080"). It tries $PORT when set and valid, then 8080, then
// random ports in 8000–8999 until one accepts a listen.
func AvailableListenAddr() (string, error) {
	tried := make(map[int]struct{})

	try := func(port int) (string, bool) {
		if port < 1 || port > 65535 {
			return "", false
		}
		if _, ok := tried[port]; ok {
			return "", false
		}
		tried[port] = struct{}{}
		addr := fmt.Sprintf(":%d", port)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return "", false
		}
		_ = ln.Close()
		return addr, true
	}

	if p := portFromEnv(); p != 0 {
		if addr, ok := try(p); ok {
			return addr, nil
		}
	}

	if addr, ok := try(8080); ok {
		return addr, nil
	}

	ports := make([]int, 1000)
	for i := range ports {
		ports[i] = 8000 + i
	}
	rand.Shuffle(len(ports), func(i, j int) { ports[i], ports[j] = ports[j], ports[i] })
	for _, p := range ports {
		if addr, ok := try(p); ok {
			return addr, nil
		}
	}

	return "", errors.New("no free TCP port found in 8000–8999 after trying PORT and 8080")
}

func portFromEnv() int {
	raw := strings.TrimSpace(os.Getenv("PORT"))
	if raw == "" {
		return 0
	}
	p, err := strconv.Atoi(raw)
	if err != nil || p < 1 || p > 65535 {
		return 0
	}
	return p
}
