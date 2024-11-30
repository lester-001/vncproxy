package wsserver

import (
	"net/http"
	"net/url"

	"github.com/amitbet/vncproxy/logger"
	"github.com/gorilla/websocket"
)

type LongConnServerConfig struct {
	UseDummySession bool
}

type LongConnServer struct {
	cfg *LongConnServerConfig
}

type LongConn struct {
	c    *websocket.Conn
	cfg  *LongConnServerConfig
	quit chan struct{}
}

type LongConnServerHandler func(*websocket.Conn, *LongConnServerConfig, string)

func wsLongHandlerFunc(ws *websocket.Conn, cfg *LongConnServerConfig, sessionId string) {

	conn := LongConn{
		cfg:  cfg,
		quit: make(chan struct{}),
	}
	ws.WriteMessage(websocket.TextMessage, []byte("Here is a string...."))
	for {
		select {
		case <-conn.quit:
			return
		default:
			mt, data, err := conn.c.ReadMessage()
			if err != nil {
				continue
			}

			logger.Debugf("%d %s", mt, string(data))
		}
	}
}

func WsLongServer(url string, cfg *LongConnServerConfig) error {
	server := LongConnServer{cfg}
	logger.Debugf("LongConnServer")
	server.Listen(url, LongConnServerHandler(wsLongHandlerFunc))
	return nil
}

var upgraderLong = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (wsServer *LongConnServer) Listen(urlStr string, handlerFunc LongConnServerHandler) {

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

			conn, err := upgraderLong.Upgrade(w, r, nil)
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
