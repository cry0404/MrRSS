package core

import (
	"net/http"
	"net/url"
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/feed"
	"MrRSS/internal/models"
)

func TestNewHandler_ConstructsHandler(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init failed: %v", err)
	}

	f := feed.NewFetcher(db)
	h := NewHandler(db, f, nil, nil)

	if h.DB == nil {
		t.Fatal("Handler DB is nil")
	}
	if h.Fetcher == nil {
		t.Fatal("Handler Fetcher is nil")
	}
	if h.DiscoveryService == nil {
		t.Fatal("DiscoveryService should be initialized")
	}
}

func TestCreateArticleHTTPClientUsesFeedProxy(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init failed: %v", err)
	}

	h := NewHandler(db, feed.NewFetcher(db), nil, nil)
	client, err := h.createArticleHTTPClient(&models.Feed{
		ProxyEnabled: true,
		ProxyURL:     "http://127.0.0.1:3128",
	})
	if err != nil {
		t.Fatalf("createArticleHTTPClient failed: %v", err)
	}

	proxyURL := proxyURLFromClient(t, client)
	if proxyURL != "http://127.0.0.1:3128" {
		t.Fatalf("proxy URL = %q", proxyURL)
	}
}

func TestCreateArticleHTTPClientUsesGlobalProxyWhenFeedRequestsIt(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init failed: %v", err)
	}
	_ = db.SetSetting("proxy_enabled", "true")
	_ = db.SetSetting("proxy_type", "http")
	_ = db.SetSetting("proxy_host", "127.0.0.1")
	_ = db.SetSetting("proxy_port", "8080")

	h := NewHandler(db, feed.NewFetcher(db), nil, nil)
	client, err := h.createArticleHTTPClient(&models.Feed{ProxyEnabled: true})
	if err != nil {
		t.Fatalf("createArticleHTTPClient failed: %v", err)
	}

	proxyURL := proxyURLFromClient(t, client)
	if proxyURL != "http://127.0.0.1:8080" {
		t.Fatalf("proxy URL = %q", proxyURL)
	}
}

func proxyURLFromClient(t *testing.T, client *http.Client) string {
	t.Helper()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type %T", client.Transport)
	}
	if transport.Proxy == nil {
		t.Fatalf("expected proxy to be configured")
	}
	reqURL, _ := url.Parse("https://example.com/article")
	req := &http.Request{URL: reqURL}
	proxy, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("proxy function returned error: %v", err)
	}
	if proxy == nil {
		t.Fatalf("proxy function returned nil")
	}
	return proxy.String()
}
