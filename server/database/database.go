package database

import (
	"context"
	"net/http"
)

type Database interface {
	GetPerson(r *http.Request) ([]Person, error)
	Context() context.Context
}
