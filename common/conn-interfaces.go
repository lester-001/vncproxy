package common

import (
	"io"
)

type IServerConn interface {
	//IServerConn() io.ReadWriter

	SetSessionId(string)
	SessionId() string
	Protocol() string
	CurrentPixelFormat() *PixelFormat
	SetPixelFormat(*PixelFormat) error
	//ColorMap() *ColorMap
	SetColorMap(*ColorMap)
	Encodings() []IEncoding
	SetEncodings([]EncodingType) error
	Width() uint16
	Height() uint16
	SetWidth(uint16)
	SetHeight(uint16)
	DesktopName() string
	SetDesktopName(string)
	//Flush() error
	SetProtoVersion(string)
	Write([]byte) (int, error)
	WriteMessage(messageType int, buf []byte) (int, error)

	Listeners() *MultiListener

	Reader() (io.Reader, error)
	NextReader() (io.Reader, error)
	Read(buf []byte) (int, error)
	Close() error
	Run() error
}

type IServerConnIO interface {
	io.ReadWriter
	//IServerConn() io.ReadWriter
	Protocol() string
	CurrentPixelFormat() *PixelFormat
	SetPixelFormat(*PixelFormat) error
	//ColorMap() *ColorMap
	SetColorMap(*ColorMap)
	Encodings() []IEncoding
	SetEncodings([]EncodingType) error
	Width() uint16
	Height() uint16
	SetWidth(uint16)
	SetHeight(uint16)
	DesktopName() string
	SetDesktopName(string)
	//Flush() error
	SetProtoVersion(string)
	// Write([]byte) (int, error)
}

type IClientConn interface {
	CurrentPixelFormat() *PixelFormat
	//CurrentColorMap() *ColorMap
	Encodings() []IEncoding
}
