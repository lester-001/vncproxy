package wsserver

import (
	"fmt"
	"log"
	"net"

	"github.com/amitbet/vncproxy/common"
	"github.com/amitbet/vncproxy/logger"
	"github.com/gorilla/websocket"
)

var DefaultClientMessages = []common.ClientMessage{
	&MsgSetPixelFormat{},
	&MsgSetEncodings{},
	&MsgFramebufferUpdateRequest{},
	&MsgKeyEvent{},
	&MsgPointerEvent{},
	&MsgClientCutText{},
	&MsgClientQemuExtendedKey{},
}

// FramebufferUpdate holds a FramebufferUpdate wire format message.
type FramebufferUpdate struct {
	_       [1]byte             // padding
	NumRect uint16              // number-of-rectangles
	Rects   []*common.Rectangle // rectangles
}

type ServerHandler func(*ServerConfig, common.IServerConn) error

type ServerConfig struct {
	SecurityHandlers []SecurityHandler
	Encodings        []common.IEncoding
	PixelFormat      *common.PixelFormat
	ColorMap         *common.ColorMap
	ClientMessages   []common.ClientMessage
	DesktopName      []byte
	Height           uint16
	Width            uint16
	UseDummySession  bool

	//handler to allow for registering for messages, this can't be a channel
	//because of the websockets handler function which will kill the connection on exit if conn.handle() is run on another thread
	NewConnHandler ServerHandler
}

func wsHandlerFunc(ws *websocket.Conn, cfg *ServerConfig, sessionId string) {
	conn, err := NewServerConn(ws, cfg)
	if err != nil {
		return
	}

	err = attachNewServerConn(conn, cfg, sessionId)
	if err != nil {
		log.Fatalf("Error attaching new connection. %v", err)
	}
}

func WsServe(url string, cfg *ServerConfig) error {
	server := WebsocketServer{cfg}
	logger.Errorf("WsServe")
	server.Listen(url, WebsocketHandler(wsHandlerFunc))
	return nil
}

func TcpServe(url string, cfg *ServerConfig) error {
	ln, err := net.Listen("tcp", url)
	if err != nil {
		log.Fatalf("Error listen. %v", err)
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			return err
		}
		conn, err := NewServerConnIO(c, cfg)
		if err != nil {
			return nil
		}

		go attachNewServerConn(conn, cfg, "dummySession")
	}
	return nil
}

func attachNewServerConn(conn common.IServerConn, cfg *ServerConfig, sessionId string) error {
	if err := ServerVersionHandler(cfg, conn); err != nil {
		fmt.Errorf("err: %v\n", err)
		conn.Close()
		return err
	}

	if err := ServerSecurityHandler(cfg, conn); err != nil {
		conn.Close()
		return err
	}

	//run the handler for this new incoming connection from a vnc-client
	//this is done before the init sequence to allow listening to server-init messages (and maybe even interception in the future)
	err := cfg.NewConnHandler(cfg, conn)
	if err != nil {
		conn.Close()
		return err
	}

	if err := ServerClientInitHandler(cfg, conn); err != nil {
		conn.Close()
		return err
	}

	if err := ServerServerInitHandler(cfg, conn); err != nil {
		conn.Close()
		return err
	}

	conn.SetSessionId(sessionId)
	if cfg.UseDummySession {
		conn.SetSessionId("dummySession")
	}

	//go here will kill ws connections
	conn.Run()

	return nil
}
