package handler

import (
	"fmt"
	"net"

	"github.com/labstack/echo/v4"
)

// NewIPExtractor builds Echo's IPExtractor, which decides how c.RealIP() finds the client IP that
// user IDs are keyed on. X-Forwarded-For is trusted only when the direct peer is in trustedProxyCIDRs
// (e.g. the BFF's network); Echo's default loopback/link-local/private-net trust is disabled so a
// client can't spoof another user's IP through the header. An empty trustedProxyCIDRs means no proxy
// is trusted and the connection's RemoteAddr is used as-is. It returns an error if any entry is not
// a valid CIDR.
func NewIPExtractor(trustedProxyCIDRs []string) (echo.IPExtractor, error) {
	if len(trustedProxyCIDRs) == 0 {
		return echo.ExtractIPDirect(), nil
	}

	proxyTrustOptions := []echo.TrustOption{
		echo.TrustLoopback(false),
		echo.TrustLinkLocal(false),
		echo.TrustPrivateNet(false),
	}
	for _, trustedProxyCIDR := range trustedProxyCIDRs {
		_, trustedProxyIPRange, err := net.ParseCIDR(trustedProxyCIDR)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", trustedProxyCIDR, err)
		}
		proxyTrustOptions = append(proxyTrustOptions, echo.TrustIPRange(trustedProxyIPRange))
	}
	return echo.ExtractIPFromXFFHeader(proxyTrustOptions...), nil
}
