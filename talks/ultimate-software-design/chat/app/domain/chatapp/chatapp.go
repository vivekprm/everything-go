package chatapp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	errs "ultimate-software-design/chat/app/sdk/err"
	"ultimate-software-design/chat/foundation/logger"
	"ultimate-software-design/chat/foundation/web"

	"github.com/gorilla/websocket"
)

type app struct {
	log *logger.Logger
	WS  *websocket.Upgrader
}

func newApp(log *logger.Logger) *app {
	return &app{
		log: log,
		WS: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (a *app) connect(ctx context.Context, r *http.Request) web.Encoder {
	// Web socket implemented here.
	// Just perform basic echo for now.
	// Make sure we are handling connection drops/issues (context)
	// How we will map connection to a user.
	c, err := a.WS.Upgrade(web.GetWriter(ctx), r, nil)
	if err != nil {
		return errs.Errorf(errs.FailedPrecondition, "unable to upgrade to websocket: %v", err)
	}
	defer c.Close()

	usr, err := a.handshake(ctx, c)

	if err != nil {
		return errs.Errorf(errs.FailedPrecondition, "handshake failed: %v", err)
	}

	a.log.Info(ctx, "handshake complete", "user", usr.Name)

	// var wg sync.WaitGroup
	// wg.Add(3)

	// ticker := time.NewTicker(time.Second)
	// defer ticker.Stop()

	// // Ping Goroutine
	// go func() {
	// 	wg.Done()
	// 	select {
	// 	case <-ticker.C:
	// 		if err := c.WriteMessage(websocket.TextMessage, []byte("ping")); err != nil {
	// 			return
	// 		}
	// 	}
	// }()

	// // Read Goroutine
	// go func() {
	// 	wg.Done()
	// 	for {
	// 		_, msg, err := c.ReadMessage()
	// 		if err != nil {
	// 			return
	// 		}
	// 	}
	// }()

	// // Write Goroutine
	// go func() {
	// 	wg.Done()
	// 	for {
	// 		msg, wd := <-ch

	// 		if !wd {
	// 			return
	// 		}

	// 		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
	// 			return
	// 		}
	// 	}
	// }()

	// wg.Wait()

	return nil
}

func (a *app) handshake(ctx context.Context, c *websocket.Conn) (user, error) {
	if err := c.WriteMessage(websocket.TextMessage, []byte("HELLO")); err != nil {
		return user{}, fmt.Errorf("write message: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()

	msg, err := a.readMessage(ctx, c)
	if err != nil {
		return user{}, fmt.Errorf("read message: %w", err)
	}

	var usr user

	if err := json.Unmarshal(msg, &usr); err != nil {
		return user{}, fmt.Errorf("unmarshal message: %w", err)
	}

	fmt.Printf("handshake successful, got message: %s\n", msg)

	v := fmt.Sprintf("WELCOME %s", usr.Name)
	if err := c.WriteMessage(websocket.TextMessage, []byte(v)); err != nil {
		return user{}, fmt.Errorf("write message: %w", err)
	}

	return usr, nil
}

func (a *app) readMessage(ctx context.Context, c *websocket.Conn) ([]byte, error) {
	type response struct {
		msg []byte
		err error
	}
	ch := make(chan response, 1)

	go func() {
		a.log.Info(ctx, "starting handshake read")
		defer a.log.Info(ctx, "completed handshake read")
		_, msg, err := c.ReadMessage()

		if err != nil {
			ch <- response{msg: nil, err: err}
			return
		}

		ch <- response{msg: msg, err: nil}
	}()

	var resp response

	select {
	case <-ctx.Done():
		c.Close()
		return nil, fmt.Errorf("handshake timeout: %w", ctx.Err())
	case resp = <-ch:
		if resp.err != nil {
			return nil, fmt.Errorf("handshake failed: %w", resp.err)
		}
	}

	return resp.msg, nil
}
