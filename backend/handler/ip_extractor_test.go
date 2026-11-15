package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/handler"
)

func TestNewIPExtractor_正常系(t *testing.T) {
	tests := []struct {
		name              string
		trustedProxyCIDRs []string
		remoteAddr        string
		xForwardedFor     []string
		want              string
	}{
		{
			name:              "信用するCIDRが空のとき、X-Forwarded-Forに偽のIPが付いていても、接続元IPが返ること",
			trustedProxyCIDRs: []string{},
			remoteAddr:        "203.0.113.9:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "203.0.113.9",
		},
		{
			name:              "信用するCIDRが空のとき、接続元がプライベートIPでもX-Forwarded-Forを無視して、接続元IPが返ること",
			trustedProxyCIDRs: []string{},
			remoteAddr:        "172.18.0.5:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "172.18.0.5",
		},
		{
			name:              "接続元が信用するCIDR内のとき、X-Forwarded-ForのIPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "172.28.0.5:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "198.51.100.1",
		},
		{
			name:              "接続元が信用するCIDR外のグローバルIPのとき、X-Forwarded-Forを無視して接続元IPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "203.0.113.9:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "203.0.113.9",
		},
		{
			name:              "接続元が信用するCIDR外のプライベートIPのとき、接続元IPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "10.0.0.1:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "10.0.0.1",
		},
		{
			name:              "接続元が信用するCIDR外のループバックIPのとき、接続元IPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "127.0.0.1:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "127.0.0.1",
		},
		{
			name:              "接続元が信用するCIDR外のリンクローカルIPのとき、接続元IPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "169.254.0.1:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "169.254.0.1",
		},
		{
			name:              "接続元が信用するCIDRの直前のとき、接続元IPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "172.27.255.255:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "172.27.255.255",
		},
		{
			name:              "接続元が信用するCIDRの先頭のとき、X-Forwarded-ForのIPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "172.28.0.0:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "198.51.100.1",
		},
		{
			name:              "接続元が信用するCIDRの末尾のとき、X-Forwarded-ForのIPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "172.28.0.255:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "198.51.100.1",
		},
		{
			name:              "接続元が信用するCIDRの直後のとき、接続元IPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "172.28.1.0:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "172.28.1.0",
		},
		{
			name:              "プレフィックス長が上限ちょうどのCIDRのみを指定し、接続元がそのIPのとき、X-Forwarded-ForのIPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.5/32"},
			remoteAddr:        "172.28.0.5:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "198.51.100.1",
		},
		{
			name:              "接続元が信用するCIDR内で、X-Forwarded-Forが攻撃者の偽のIPとBFFが付けた実IPの順に並ぶとき、実IPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "172.28.0.5:12345",
			xForwardedFor:     []string{"6.6.6.6, 198.51.100.1"},
			want:              "198.51.100.1",
		},
		{
			name:              "信用するCIDRを複数指定し、接続元が2つ目のCIDR内のとき、X-Forwarded-ForのIPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24", "10.0.1.0/24"},
			remoteAddr:        "10.0.1.5:12345",
			xForwardedFor:     []string{"198.51.100.1"},
			want:              "198.51.100.1",
		},
		{
			name:              "接続元が信用するCIDR内でX-Forwarded-Forが付いていないとき、接続元IPが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24"},
			remoteAddr:        "172.28.0.5:12345",
			xForwardedFor:     nil,
			want:              "172.28.0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header[echo.HeaderXForwardedFor] = tt.xForwardedFor

			sut, err := handler.NewIPExtractor(tt.trustedProxyCIDRs)
			require.NoError(t, err)
			got := sut(req)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewIPExtractor_異常系(t *testing.T) {
	tests := []struct {
		name              string
		trustedProxyCIDRs []string
	}{
		{
			name:              "範囲外のオクテットを含むCIDRが含まれるとき、エラーが返ること",
			trustedProxyCIDRs: []string{"172.28.0.999/32"},
		},
		{
			name:              "プレフィックス長が範囲外のCIDRが含まれるとき、エラーが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/33"},
		},
		{
			name:              "プレフィックス長のないIPが含まれるとき、エラーが返ること",
			trustedProxyCIDRs: []string{"172.28.0.5"},
		},
		{
			name:              "空文字の要素が含まれるとき、エラーが返ること",
			trustedProxyCIDRs: []string{"172.28.0.0/24", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := handler.NewIPExtractor(tt.trustedProxyCIDRs)

			require.Error(t, err)
		})
	}
}
