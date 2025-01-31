package config

import (
	"net/http"

	"github.com/michaeljs1990/sqlitestore"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func ConfigureSessions(e *echo.Echo, dbPath string, secret string) error {
    store, err := sqlitestore.NewSqliteStore(
        dbPath,
        "sessions",
        "/",
        3600, // 1 week
        []byte(secret),
    )
    if err != nil {
        return err
    }

    store.Options.Secure = true
    store.Options.HttpOnly = true
    store.Options.SameSite = http.SameSiteStrictMode

    e.Use(session.Middleware(store))
    return nil
}