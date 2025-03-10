package main

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"path"
	"strconv"
	"time"
)

func (app *App) GetIndex(w http.ResponseWriter, _ *http.Request) {
	document := Doc(
		Ela("html", []Attr{{"lang", "en"}},
			El("head",
				Sela("meta", []Attr{{"charset", "utf-8"}}),
				Sela("meta", []Attr{{"name", "viewport"}, {"content", "width=device-width, initial-scale=1"}}),
				Sela("meta", []Attr{{"name", "description"}, {"content", "Anonymously share code snippets."}}),
				El("title", "Paste"),
				Sela("link", []Attr{{"rel", "stylesheet"}, {"href", "/css/toastify.min.css"}}),
				Sela("link", []Attr{{"rel", "stylesheet"}, {"href", "/css/style.css"}}),
				Ela("script", []Attr{{"src", "/js/toastify.min.js"}}),
				Ela("script", []Attr{{"src", "/js/index.js"}, {"defer", ""}}),
			),
			El("body",
				Ela("nav", []Attr{{"id", "nav"}},
					El("div",
						Ela("h1", []Attr{{"id", "title"}},
							Ela("a", []Attr{{"href", "/"}}, "Paste"),
						),
					),
					El("div",
						Ela("button", []Attr{{"id", "save"}}, "Save"),
					),
				),
				Ela("textarea", []Attr{
					{"id", "content"},
					{"name", "content"},
					{"placeholder", "Enter text here..."},
					{"wrap", "off"},
					{"autofocus", ""},
				}),
			),
		),
	)

	app.SendHtml(w, document)
}

func (app *App) PostSave() http.HandlerFunc {
	limiter := NewAddrRateLimiter(rate.Every(5*time.Second), 10)

	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 512*1024)

		data, err := io.ReadAll(r.Body)

		if err != nil {
			if errors.As(err, new(*http.MaxBytesError)) {
				app.SendApiError(w, r, http.StatusRequestEntityTooLarge, "Your paste is too large.")
			} else {
				app.SendApiError(w, r, http.StatusBadRequest, "Something went wrong. Please try again.")
			}
			return
		}

		if len(data) == 0 {
			app.SendApiError(w, r, http.StatusBadRequest, "Please enter something to save.")
			return
		}

		addr := netip.MustParseAddr(r.RemoteAddr)

		if !limiter.Allow(addr) {
			app.SendApiError(w, r, http.StatusTooManyRequests, "You have been rate limited.")
			return
		}

		id, err := GenerateId()

		if err != nil {
			app.HandleApiError(w, r, err)
			return
		}

		milli := time.Now().UTC().UnixMilli()
		countryCode := r.Header.Get("CF-IPCountry")

		_, err = app.db.Exec("INSERT INTO pastes (id, created_at, country_code, content) VALUES ($1, $2, $3, $4)", id, milli, countryCode, data)

		if err != nil {
			app.HandleApiError(w, r, err)
			return
		}

		type Response struct {
			Id string `json:"id"`
		}

		app.SendJson(w, r, http.StatusOK, Response{Id: id})
	}
}

func (app *App) GetPaste(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	row := app.db.QueryRow("SELECT created_at, content FROM pastes WHERE id = $1", id)

	var createdAt int64
	var content []byte

	if err := row.Scan(&createdAt, &content); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.NotFound(w, r)
		} else {
			app.HandleError(w, r, err)
		}
		return
	}

	if err := app.IncrementViews(id); err != nil {
		app.HandleError(w, r, err)
		return
	}

	createdAtTime := time.UnixMilli(createdAt).UTC().Round(time.Second)
	createdAtAgo := TimeAgo(createdAtTime)
	createdAtRfc := createdAtTime.Format(time.RFC3339)

	lines := bytes.Count(content, []byte("\n")) + 1

	document := Doc(
		Ela("html", []Attr{{"lang", "en"}},
			El("head",
				Sela("meta", []Attr{{"charset", "utf-8"}}),
				Sela("meta", []Attr{{"name", "viewport"}, {"content", "width=device-width, initial-scale=1"}}),
				El("title", "Paste"),
				Sela("link", []Attr{
					{"rel", "stylesheet"},
					{"href", "/css/vs.min.css"},
					{"media", "(prefers-color-scheme: light), (prefers-color-scheme: no-preference)"},
				}),
				Sela("link", []Attr{
					{"rel", "stylesheet"},
					{"href", "/css/vs2015.min.css"},
					{"media", "(prefers-color-scheme: dark)"},
				}),
				Sela("link", []Attr{{"rel", "stylesheet"}, {"href", "/css/style.css"}}),
				Ela("script", []Attr{{"src", "/js/highlight.min.js"}}),
				Ela("script", []Attr{{"src", "/js/view.js"}, {"defer", ""}}),
			),
			El("body",
				Ela("nav", []Attr{{"id", "nav"}},
					El("div",
						Ela("h1", []Attr{{"id", "title"}},
							Ela("a", []Attr{{"href", "/"}}, "Paste"),
						),
					),
					Ela("div", []Attr{{"id", "metadata"}},
						Ela("a", []Attr{{"id", "new"}, {"href", "/"}}, "New"),
						Ela("a", []Attr{{"id", "raw"}, {"href", "/raw/" + id}}, "Raw"),
						Ela("time", []Attr{
							{"id", "created-at"},
							{"datetime", createdAtRfc},
							{"title", createdAtRfc},
						}, createdAtAgo),
						" - ",
						strconv.Itoa(lines)+" line", IfElse(lines == 1, "", "s"),
					),
				),
				Ela("pre", []Attr{{"id", "content"}}, Esc(ToString(content))),
			),
		),
	)

	app.SendHtml(w, document)
}

func (app *App) GetRawPaste(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	row := app.db.QueryRow("SELECT content FROM pastes WHERE id = $1", id)

	var content []byte

	if err := row.Scan(&content); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.NotFound(w, r)
		} else {
			app.HandleError(w, r, err)
		}
		return
	}

	if err := app.IncrementViews(id); err != nil {
		app.HandleError(w, r, err)
		return
	}

	app.SendText(w, content)
}

func (app *App) IncrementViews(id string) error {
	_, err := app.db.Exec("UPDATE pastes SET views = views + 1, last_view = $1 WHERE id = $2", time.Now().UTC().UnixMilli(), id)
	return err
}

func (app *App) ServeAssets(w http.ResponseWriter, r *http.Request) {
	file, err := assets.Open(path.Join("assets", r.URL.Path))

	if err != nil {
		app.NotFound(w, r)
		return
	}

	info, err := file.Stat()

	if err != nil || info.IsDir() {
		app.NotFound(w, r)
		return
	}

	// Tell Cloudflare to cache assets for 7 days
	w.Header().Set("Cloudflare-CDN-Cache-Control", "max-age=604800")

	// Tell browsers to cache assets for 5 minutes
	w.Header().Set("Cache-Control", "max-age=300")

	http.ServeContent(w, r, info.Name(), info.ModTime(), file.(io.ReadSeeker))
}

func (app *App) NotFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func (app *App) MethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func (app *App) RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
			r.RemoteAddr = ip
		} else {
			r.RemoteAddr, _, _ = net.SplitHostPort(r.RemoteAddr)
		}

		next.ServeHTTP(w, r)
	})
}

func (app *App) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			log.Printf("%s %d %dB %s %s %s %s %s\n",
				time.Since(now), ww.Status(), ww.BytesWritten(), r.Header.Get("CF-IPCountry"), r.RemoteAddr, r.Method, r.RequestURI, r.UserAgent())
		}()

		next.ServeHTTP(ww, r)
	})
}

func (app *App) RateLimit(next http.Handler) http.Handler {
	limiter := NewAddrRateLimiter(rate.Limit(5), 50)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		addr := netip.MustParseAddr(r.RemoteAddr)

		if !limiter.Allow(addr) {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
