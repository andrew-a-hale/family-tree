package data

import (
	"family-tree/server/database"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Provider struct {
	HttpHandler    http.Handler
	LogHandle      *slog.Logger
	DatabaseHandle *database.Db
}

func (p *Provider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	p.LogHandle.LogAttrs(
		r.Context(),
		slog.LevelInfo,
		"Start",
		slog.Attr{Key: "Method", Value: slog.StringValue(r.Method)},
		slog.Attr{Key: "Path", Value: slog.StringValue(r.URL.Path)},
	)
	p.HttpHandler.ServeHTTP(w, r)
	p.LogHandle.LogAttrs(
		r.Context(),
		slog.LevelInfo,
		"Finish",
		slog.Attr{Key: "Method", Value: slog.StringValue(r.Method)},
		slog.Attr{Key: "Path", Value: slog.StringValue(r.URL.Path)},
		slog.Attr{Key: "DurationMs", Value: slog.Int64Value(time.Since(start).Milliseconds())},
	)
}

func MakeHandler(p *Provider, f func(*Provider, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f(p, w, r)
	}
}

func HelloNeo4j(p *Provider, w http.ResponseWriter, r *http.Request) {
	res, err := p.DatabaseHandle.GetPerson(r)
	if err != nil {
		p.LogHandle.LogAttrs(
			r.Context(),
			slog.LevelError,
			"Failed to query neo4j",
			slog.Attr{Key: "Method", Value: slog.StringValue(r.Method)},
			slog.Attr{Key: "Path", Value: slog.StringValue(r.URL.Path)},
			slog.Attr{Key: "Error", Value: slog.StringValue(err.Error())},
		)
		return
	}

	for _, p := range res {
		fmt.Fprintf(w, "%v\n", p)
	}
}
