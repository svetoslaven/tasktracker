package main

import (
	"net/http"
)

type middleware func(http.Handler) http.Handler

func (app *application) newMiddlewareChain(middlewares ...middleware) middleware {
	return middleware(func(next http.Handler) http.Handler {
		for i := 0; i < len(middlewares); i++ {
			next = middlewares[i](next)
		}

		return next
	})
}
