package main

func (app *application) sendEmail(recipient, templateFile string, data any) {
	app.runInBackground(func() {
		if err := app.mailer.Send(recipient, templateFile, data); err != nil {
			app.logger.LogError(err, nil)
		}
	})
}
