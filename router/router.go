package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/huangxuantao/screego/auth"
	"github.com/huangxuantao/screego/config"
	"github.com/huangxuantao/screego/ui"
	"github.com/huangxuantao/screego/ws"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

type UIConfig struct {
	AuthMode                 string `json:"authMode"`
	User                     string `json:"user"`
	LoggedIn                 bool   `json:"loggedIn"`
	Version                  string `json:"version"`
	CloseRoomWhenOwnerLeaves bool   `json:"closeRoomWhenOwnerLeaves"`
}

func Router(conf config.Config, rooms *ws.Rooms, users *auth.Users, version string) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// https://github.com/gorilla/mux/issues/416
		accessLogger(r, 404, 0, 0)
	})
	router.Use(hlog.AccessHandler(accessLogger))
	router.Use(handlers.CORS(handlers.AllowedMethods([]string{"GET", "POST"}), handlers.AllowedOriginValidator(conf.CheckOrigin)))
	router.HandleFunc("/stream", rooms.Upgrade)
	router.Methods("POST").Path("/login").HandlerFunc(users.Authenticate)
	router.Methods("POST").Path("/logout").HandlerFunc(users.Logout)
	router.Methods("GET").Path("/config").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, loggedIn := users.CurrentUser(r)
		_ = json.NewEncoder(w).Encode(&UIConfig{
			AuthMode:                 conf.AuthMode,
			LoggedIn:                 loggedIn,
			User:                     user,
			Version:                  version,
			CloseRoomWhenOwnerLeaves: conf.CloseRoomWhenOwnerLeaves,
		})
	})
	if conf.Prometheus {
		log.Info().Msg("Prometheus enabled")
		router.Methods("GET").Path("/metrics").Handler(basicAuth(promhttp.Handler(), users))
	}

	ui.Register(router)

	return router
}

func accessLogger(r *http.Request, status, size int, dur time.Duration) {
	log.Debug().
		Str("host", r.Host).
		Int("status", status).
		Int("size", size).
		Str("ip", r.RemoteAddr).
		Str("path", r.URL.Path).
		Str("duration", dur.String()).
		Msg("HTTP")
}

func basicAuth(handler http.Handler, users *auth.Users) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || !users.Validate(user, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="screego"`)
			w.WriteHeader(401)
			_, _ = w.Write([]byte("Unauthorized.\n"))
			return
		}

		handler.ServeHTTP(w, r)
	}
}
