package main

import (
	"fmt"
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

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection:", "close")
				app.sendServerErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
