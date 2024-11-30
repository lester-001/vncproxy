package proxy

import (
	"net"
	"path"
	"strconv"
	"time"

	"github.com/amitbet/vncproxy/client"
	"github.com/amitbet/vncproxy/common"
	"github.com/amitbet/vncproxy/encodings"
	"github.com/amitbet/vncproxy/logger"
	listeners "github.com/amitbet/vncproxy/recorder"
	"github.com/amitbet/vncproxy/wsserver"
)

type VncProxy struct {
	TCPListeningURL  string      // empty = not listening on tcp
	WsListeningURL   string      // empty = not listening on ws
	RecordingDir     string      // empty = no recording
	ProxyVncPassword string      //empty = no auth
	SingleSession    *VncSession // to be used when not using sessions
	UsingSessions    bool        //false = single session - defined in the var above
	sessionManager   *SessionManager
}

func (vp *VncProxy) createClientConnection(target string, vncPass string) (*client.ClientConn, error) {
	var (
		nc  net.Conn
		err error
	)

	if target[0] == '/' {
		nc, err = net.Dial("unix", target)
	} else {
		nc, err = net.Dial("tcp", target)
	}

	if err != nil {
		logger.Errorf("error connecting to vnc server: %s", err)
		return nil, err
	}

	var noauth client.ClientAuthNone
	authArr := []client.ClientAuth{&client.PasswordAuth{Password: vncPass}, &noauth}

	clientConn, err := client.NewClientConn(nc,
		&client.ClientConfig{
			Auth:      authArr,
			Exclusive: true,
		})

	if err != nil {
		logger.Errorf("error creating client: %s", err)
		return nil, err
	}

	return clientConn, nil
}

// if sessions not enabled, will always return the configured target server (only one)
func (vp *VncProxy) getProxySession(sessionId string) (*VncSession, error) {

	if !vp.UsingSessions {
		if vp.SingleSession == nil {
			logger.Errorf("SingleSession is empty, use sessions or populate the SingleSession member of the VncProxy struct.")
		}
		return vp.SingleSession, nil
	}
	return vp.sessionManager.GetSession(sessionId)
}

func (vp *VncProxy) newwsServerConnHandler(cfg *wsserver.ServerConfig, conn common.IServerConn) error {
	var err error
	session, err := vp.getProxySession(conn.SessionId())
	if err != nil {
		logger.Errorf("Proxy.newServerConnHandler can't get session: %d", conn.SessionId())
		return err
	}

	var rec *listeners.Recorder

	if session.Type == SessionTypeRecordingProxy {
		recFile := "recording" + strconv.FormatInt(time.Now().Unix(), 10) + ".rbs"
		recPath := path.Join(vp.RecordingDir, recFile)
		rec, err = listeners.NewRecorder(recPath)
		if err != nil {
			logger.Errorf("Proxy.newServerConnHandler can't open recorder save path: %s", recPath)
			return err
		}

		conn.Listeners().AddListener(rec)
	}

	session.Status = SessionStatusInit
	if session.Type == SessionTypeProxyPass || session.Type == SessionTypeRecordingProxy {
		target := session.Target
		if session.TargetHostname != "" && session.TargetPort != "" {
			target = session.TargetHostname + ":" + session.TargetPort
		}

		cconn, err := vp.createClientConnection(target, session.TargetPassword)
		if err != nil {
			session.Status = SessionStatusError
			logger.Errorf("Proxy.newServerConnHandler error creating connection: %s", err)
			return err
		}
		if session.Type == SessionTypeRecordingProxy {
			cconn.Listeners().AddListener(rec)
		}

		//creating cross-listeners between server and client parts to pass messages through the proxy:

		// gets the bytes from the actual vnc server on the env (client part of the proxy)
		// and writes them through the server socket to the vnc-client
		serverUpdater := &wsServerUpdater{conn}
		cconn.Listeners().AddListener(serverUpdater)

		// gets the messages from the server part (from vnc-client),
		// and write through the client to the actual vnc-server
		clientUpdater := &ClientUpdater{cconn}
		conn.Listeners().AddListener(clientUpdater)

		err = cconn.Connect()
		if err != nil {
			session.Status = SessionStatusError
			logger.Errorf("Proxy.newServerConnHandler error connecting to client: %s", err)
			return err
		}

		encs := []common.IEncoding{
			&encodings.RawEncoding{},
			&encodings.TightEncoding{},
			&encodings.EncCursorPseudo{},
			&encodings.EncLedStatePseudo{},
			&encodings.TightPngEncoding{},
			&encodings.RREEncoding{},
			&encodings.ZLibEncoding{},
			&encodings.ZRLEEncoding{},
			&encodings.CopyRectEncoding{},
			&encodings.CoRREEncoding{},
			&encodings.HextileEncoding{},
		}
		cconn.Encs = encs

		if err != nil {
			session.Status = SessionStatusError
			logger.Errorf("Proxy.newServerConnHandler error connecting to client: %s", err)
			return err
		}
	}
	/*
		if session.Type == SessionTypeReplayServer {
			fbs, err := player.ConnectFbsFile(session.ReplayFilePath, sconn)

			if err != nil {
				logger.Error("TestServer.NewConnHandler: Error in loading FBS: ", err)
				return err
			}
			sconn.Listeners.AddListener(player.NewFBSPlayListener(sconn, fbs))
			return nil

		}
	*/
	session.Status = SessionStatusActive
	return nil
}

func (vp *VncProxy) StartListening() {

	wssecHandlers := []wsserver.SecurityHandler{&wsserver.ServerAuthNone{}}

	if vp.ProxyVncPassword != "" {
		wssecHandlers = []wsserver.SecurityHandler{&wsserver.ServerAuthVNC{vp.ProxyVncPassword}}
	}

	wscfg := &wsserver.ServerConfig{
		SecurityHandlers: wssecHandlers,
		Encodings:        []common.IEncoding{&encodings.RawEncoding{}, &encodings.TightEncoding{}, &encodings.CopyRectEncoding{}},
		PixelFormat:      common.NewPixelFormat(32),
		ClientMessages:   wsserver.DefaultClientMessages,
		DesktopName:      []byte("workDesk"),
		Height:           uint16(768),
		Width:            uint16(1024),
		NewConnHandler:   vp.newwsServerConnHandler,
		UseDummySession:  !vp.UsingSessions,
	}

	{
		longcfg := wsserver.LongConnServerConfig{UseDummySession: !vp.UsingSessions}
		go wsserver.WsLongServer("http://0.0.0.0:5908/ws", &longcfg)

	}

	if vp.TCPListeningURL != "" && vp.WsListeningURL != "" {
		logger.Infof("running two listeners: tcp port: %s, ws url: %s", vp.TCPListeningURL, vp.WsListeningURL)

		go wsserver.WsServe(vp.WsListeningURL, wscfg)
		//go server.WsServe(vp.WsListeningURL, cfg)
		wsserver.TcpServe(vp.TCPListeningURL, wscfg)
	}

	if vp.WsListeningURL != "" {
		logger.Infof("running ws listener url: %s", vp.WsListeningURL)
		wsserver.WsServe(vp.WsListeningURL, wscfg)
		//server.WsServe(vp.WsListeningURL, cfg)
	}
	if vp.TCPListeningURL != "" {
		logger.Infof("running tcp listener on port: %s", vp.TCPListeningURL)
		wsserver.TcpServe(vp.TCPListeningURL, wscfg)
	}
}
