package main

import "fmt"

func (app *application) runInBackground(fn func()) {
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				app.logger.LogError(fmt.Errorf("%s", err), nil)
			}
		}()

		fn()
	}()
}
