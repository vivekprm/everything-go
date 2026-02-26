package chatapp

import (
	"net/http"
	"ultimate-software-design/chat/foundation/logger"
	"ultimate-software-design/chat/foundation/web"
)

func Routes(app *web.App, log *logger.Logger) {
	api := newApp(log)

	app.HandlerFunc(http.MethodGet, "", "/connect", api.connect)
}
