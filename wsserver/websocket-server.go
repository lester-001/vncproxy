package wsserver

import (
	"net/http"
	"net/url"

	"github.com/amitbet/vncproxy/logger"
	"github.com/gorilla/websocket"
)

type WebsocketServer struct {
	cfg *ServerConfig
}

type WebsocketHandler func(*websocket.Conn, *ServerConfig, string)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (wsServer *WebsocketServer) Listen(urlStr string, handlerFunc WebsocketHandler) {

	if urlStr == "" {
		urlStr = "/"
	}
	url, err := url.Parse(urlStr)
	if err != nil {
		logger.Errorf("error while parsing url: ", err)
	}

	http.HandleFunc(url.Path,
		func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			var sessionId string
			if path != "" {
				sessionId = path[1:]
			}

			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				// panic(err)
				logger.Errorf("%s, error while Upgrading websocket connection\n", err.Error())
				return
			}

			handlerFunc(conn, wsServer.cfg, sessionId)
		})

	err = http.ListenAndServe(url.Host, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
