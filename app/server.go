package app

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type DeeplinkCache struct {
	mu       sync.Mutex
	lastLink string
	lastTime time.Time
	window   time.Duration
}

func newDeeplinkCache(window time.Duration) *DeeplinkCache {
	return &DeeplinkCache{window: window}
}

func (c *DeeplinkCache) IsDuplicate(deeplink string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if c.lastLink == deeplink && now.Sub(c.lastTime) < c.window {
		return true
	}

	c.lastLink = deeplink
	c.lastTime = now
	return false
}

func StartServer(cfg AppConfig) error {
	server, err := initServer(cfg)
	if err != nil {
		return err
	}

	slog.Info("Starting server", "addr", server.Addr)
	return server.ListenAndServe()
}

func initServer(cfg AppConfig) (*http.Server, error) {
	mux := http.NewServeMux()
	dp := cfg.AppDeepLink
	key := []byte(cfg.Key)

	dedup := newDeeplinkCache(5 * time.Second)

	mux.HandleFunc("POST /dl/{deeplink...}", requireAPIKey(key, rateLimit(time.Second, func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, dp, dedup)
	})))

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return server, nil
}

func requireAPIKey(expectedKey []byte, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientKey := r.Header.Get("X-API-Key")

		if subtle.ConstantTimeCompare([]byte(clientKey), expectedKey) != 1 {
			http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func rateLimit(interval time.Duration, next http.HandlerFunc) http.HandlerFunc {
	var lastReq atomic.Int64
	minInterval := interval.Nanoseconds()

	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now().UnixNano()
		last := lastReq.Load()

		if now-last < minInterval {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		if !lastReq.CompareAndSwap(last, now) {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}

func handler(w http.ResponseWriter, r *http.Request, dp string, dedup *DeeplinkCache) {
	deeplink := r.PathValue("deeplink")
	if deeplink == "" || !strings.HasPrefix(deeplink, dp) {
		http.Error(w, "Invalid deeplink prefix", http.StatusBadRequest)
		return
	}

	if dedup.IsDuplicate(deeplink) {
		http.Error(w, "Duplicate deeplink", http.StatusForbidden)
		return
	}

	slog.Debug("[Handler] Request from reader. Reading:", "deeplink", deeplink)
	go func() {
		if err := ExecuteDeepLink(deeplink); err != nil {
			slog.Error("[DeepLink] Failed to execute deeplink", "deeplink", deeplink, "error", err)
		}
	}()
	w.WriteHeader(http.StatusOK)
}
