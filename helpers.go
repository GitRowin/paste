package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/time/rate"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"
)

func (app *App) LogError(r *http.Request, err error) {
	log.Println(r.Method, r.RequestURI, err)
}

func (app *App) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	app.LogError(r, err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *App) HandleApiError(w http.ResponseWriter, r *http.Request, err error) {
	app.LogError(r, err)
	app.SendApiError(w, r, http.StatusInternalServerError, "Something went wrong. Please try again.")
}

func (app *App) SendJson(w http.ResponseWriter, r *http.Request, code int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		app.LogError(r, err)
	}
}

func (app *App) SendApiError(w http.ResponseWriter, r *http.Request, code int, message string) {
	type Response struct {
		Error string `json:"error"`
	}

	app.SendJson(w, r, code, &Response{Error: message})
}

func (app *App) SendHtml(w http.ResponseWriter, s string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, s)
}

func (app *App) SendText(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write(data)
}

const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var charactersLength = big.NewInt(int64(len(characters)))

func GenerateId() (string, error) {
	result := make([]byte, 8)

	for i := 0; i < len(result); i++ {
		n, err := rand.Int(rand.Reader, charactersLength)

		if err != nil {
			return "", err
		}

		result[i] = characters[n.Int64()]
	}

	return string(result), nil
}

func TimeAgo(t time.Time) string {
	ago := time.Now().Sub(t)
	seconds := int(ago.Seconds())

	switch {
	case seconds < 60:
		return fmt.Sprintf("%ds ago", seconds)
	case seconds < 3600:
		return fmt.Sprintf("%dm ago", seconds/60)
	case seconds < 86400:
		return fmt.Sprintf("%dh ago", seconds/3600)
	default:
		return fmt.Sprintf("%dd ago", seconds/86400)
	}
}

func ToString(input []byte) string {
	builder := strings.Builder{}
	builder.Write(input)
	return builder.String()
}

type AddrRateLimiter struct {
	lock  sync.Mutex
	cache *ttlcache.Cache[netip.Prefix, *rate.Limiter]
	r     rate.Limit
	b     int
}

func NewAddrRateLimiter(r rate.Limit, b int) *AddrRateLimiter {
	cache := ttlcache.New[netip.Prefix, *rate.Limiter](
		ttlcache.WithTTL[netip.Prefix, *rate.Limiter](time.Minute),
	)

	go cache.Start()

	return &AddrRateLimiter{
		cache: cache,
		r:     r,
		b:     b,
	}
}

func (l *AddrRateLimiter) Allow(addr netip.Addr) bool {
	var prefix netip.Prefix

	if addr.Is4() {
		prefix, _ = addr.Prefix(32)
	} else {
		prefix, _ = addr.Prefix(64)
	}

	var limiter *rate.Limiter

	l.lock.Lock()

	if item := l.cache.Get(prefix); item == nil {
		limiter = rate.NewLimiter(l.r, l.b)
		l.cache.Set(prefix, limiter, ttlcache.DefaultTTL)
	} else {
		limiter = item.Value()
	}

	l.lock.Unlock()

	return limiter.Allow()
}
