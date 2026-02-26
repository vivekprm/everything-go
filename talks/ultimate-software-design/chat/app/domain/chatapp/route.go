package chatapp

import (
	"net/http"
	"ultimate-software-design/chat/foundation/web"
)

func Routes(app *web.App) {
	api := newApp()

	app.HandlerFunc(http.MethodGet, "", "/test", api.test)
}
