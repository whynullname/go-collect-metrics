package cidrmiddleware

import (
	"net"
	"net/http"

	config "github.com/whynullname/go-collect-metrics/internal/configs/serverconfig"
	"github.com/whynullname/go-collect-metrics/internal/logger"
)

const headerKey = "X-Real-IP"

func CheckCIDR(config *config.ServerConfig) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return checkCIDR(h, config)
	}
}

func checkCIDR(next http.Handler, config *config.ServerConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.TrustedSubnet == nil {
			next.ServeHTTP(w, r)
			return
		}

		ipAdress := r.Header.Get(headerKey)
		if ipAdress == "" {
			logger.Log.Error("Ip adress in header empty, but trusted subnet not nil, return")
			return
		}

		if !isTrustedIP(ipAdress, config.TrustedSubnet) {
			logger.Log.Warnf("attempt to connect from untrusted IP address: %v\n", ipAdress)
			w.WriteHeader(http.StatusForbidden)
			return
		}
	})
}

func isTrustedIP(ipStr string, trustedNet *net.IPNet) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return trustedNet.Contains(ip)
}
