package wsserver

import (
	"bytes"
	"encoding/binary"

	"github.com/amitbet/vncproxy/common"
	"github.com/amitbet/vncproxy/logger"
)

// Key represents a VNC key press.
type Key uint32

//go:generate stringer -type=Key

// Keys is a slice of Key values.
type Keys []Key

// MsgSetPixelFormat holds the wire format message.
type MsgSetPixelFormat struct {
	_  [3]byte            // padding
	PF common.PixelFormat // pixel-format
	_  [3]byte            // padding after pixel format
}

func (*MsgSetPixelFormat) Type() common.ClientMessageType {
	return common.SetPixelFormatMsgType
}

func (msg *MsgSetPixelFormat) Write(c common.IServerConn) error {
	data := bytes.Buffer{}
	if err := binary.Write(&data, binary.BigEndian, msg.Type()); err != nil {
		return err
	}

	if err := binary.Write(&data, binary.BigEndian, msg); err != nil {
		return err
	}

	c.Write(data.Bytes())
	//pf := c.CurrentPixelFormat()
	// Invalidate the color map.
	// if pf.TrueColor {
	// 	c.SetColorMap(&common.ColorMap{})
	// }

	return nil
}

func (*MsgSetPixelFormat) Read(c common.IServerConn) (common.ClientMessage, error) {
	msg := MsgSetPixelFormat{}
	r, err := c.Reader()
	if err != nil {
		return nil, nil
	}

	if err := binary.Read(r, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// MsgSetEncodings holds the wire format message, sans encoding-type field.
type MsgSetEncodings struct {
	_         [1]byte // padding
	EncNum    uint16  // number-of-encodings
	Encodings []common.EncodingType
}

func (*MsgSetEncodings) Type() common.ClientMessageType {
	return common.SetEncodingsMsgType
}

func (*MsgSetEncodings) Read(c common.IServerConn) (common.ClientMessage, error) {
	msg := MsgSetEncodings{}
	var pad [1]byte

	r, err := c.Reader()
	if err != nil {
		logger.Errorf("MsgSetEncodings.read: unkown reader")
		return nil, nil
	}

	if err := binary.Read(r, binary.BigEndian, &pad); err != nil {
		logger.Errorf("MsgSetEncodings.read: unkown pad")
		return nil, err
	}

	if err := binary.Read(r, binary.BigEndian, &msg.EncNum); err != nil {
		logger.Errorf("MsgSetEncodings.read: unkown EncNum")
		return nil, err
	}
	var enc common.EncodingType
	for i := uint16(0); i < msg.EncNum; i++ {
		if err := binary.Read(r, binary.BigEndian, &enc); err != nil {
			logger.Errorf("MsgSetEncodings.read: unkown enc")
			return nil, err
		}
		msg.Encodings = append(msg.Encodings, enc)
	}
	c.SetEncodings(msg.Encodings)
	return &msg, nil
}

func (msg *MsgSetEncodings) Write(c common.IServerConn) error {
	data := bytes.Buffer{}
	if err := binary.Write(&data, binary.BigEndian, msg.Type()); err != nil {
		return err
	}

	var pad [1]byte
	if err := binary.Write(&data, binary.BigEndian, pad); err != nil {
		return err
	}

	if uint16(len(msg.Encodings)) > msg.EncNum {
		msg.EncNum = uint16(len(msg.Encodings))
	}
	if err := binary.Write(&data, binary.BigEndian, msg.EncNum); err != nil {
		return err
	}
	for _, enc := range msg.Encodings {
		if err := binary.Write(&data, binary.BigEndian, enc); err != nil {
			return err
		}
	}
	c.Write(data.Bytes())
	return nil
}

// MsgFramebufferUpdateRequest holds the wire format message.
type MsgFramebufferUpdateRequest struct {
	Inc           uint8  // incremental
	X, Y          uint16 // x-, y-position
	Width, Height uint16 // width, height
}

func (*MsgFramebufferUpdateRequest) Type() common.ClientMessageType {
	return common.FramebufferUpdateRequestMsgType
}

func (*MsgFramebufferUpdateRequest) Read(c common.IServerConn) (common.ClientMessage, error) {
	msg := MsgFramebufferUpdateRequest{}

	r, err := c.Reader()
	if err != nil {
		return nil, nil
	}

	if err := binary.Read(r, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *MsgFramebufferUpdateRequest) Write(c common.IServerConn) error {
	data := bytes.Buffer{}
	if err := binary.Write(&data, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(&data, binary.BigEndian, msg); err != nil {
		return err
	}
	c.Write(data.Bytes())
	return nil
}

// MsgKeyEvent holds the wire format message.
type MsgKeyEvent struct {
	Down uint8   // down-flag
	_    [2]byte // padding
	Key  Key     // key
}

func (*MsgKeyEvent) Type() common.ClientMessageType {
	return common.KeyEventMsgType
}

func (*MsgKeyEvent) Read(c common.IServerConn) (common.ClientMessage, error) {
	msg := MsgKeyEvent{}
	r, err := c.Reader()
	if err != nil {
		return nil, nil
	}

	if err := binary.Read(r, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *MsgKeyEvent) Write(c common.IServerConn) error {
	data := bytes.Buffer{}
	if err := binary.Write(&data, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(&data, binary.BigEndian, msg); err != nil {
		return err
	}
	c.Write(data.Bytes())
	return nil
}

// MsgKeyEvent holds the wire format message.
type MsgQEMUExtKeyEvent struct {
	SubmessageType uint8  // submessage type
	DownFlag       uint16 // down-flag
	KeySym         Key    // key symbol
	KeyCode        uint32 // scan code
}

func (*MsgQEMUExtKeyEvent) Type() common.ClientMessageType {
	return common.QEMUExtendedKeyEventMsgType
}

func (*MsgQEMUExtKeyEvent) Read(c common.IServerConn) (common.ClientMessage, error) {
	msg := MsgKeyEvent{}
	r, err := c.Reader()
	if err != nil {
		return nil, nil
	}

	if err := binary.Read(r, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *MsgQEMUExtKeyEvent) Write(c common.IServerConn) error {
	data := bytes.Buffer{}
	if err := binary.Write(&data, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(&data, binary.BigEndian, msg); err != nil {
		return err
	}
	c.Write(data.Bytes())
	return nil
}

// PointerEventMessage holds the wire format message.
type MsgPointerEvent struct {
	Mask uint8  // button-mask
	X, Y uint16 // x-, y-position
}

func (*MsgPointerEvent) Type() common.ClientMessageType {
	return common.PointerEventMsgType
}

func (*MsgPointerEvent) Read(c common.IServerConn) (common.ClientMessage, error) {
	msg := MsgPointerEvent{}
	r, err := c.Reader()
	if err != nil {
		return nil, nil
	}

	if err := binary.Read(r, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *MsgPointerEvent) Write(c common.IServerConn) error {
	data := bytes.Buffer{}
	if err := binary.Write(&data, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(&data, binary.BigEndian, msg); err != nil {
		return err
	}
	c.Write(data.Bytes())
	return nil
}

type MsgClientFence struct {
}

func (*MsgClientFence) Type() common.ClientMessageType {
	return common.ClientFenceMsgType
}

func (cf *MsgClientFence) Read(c common.IServerConn) (common.ClientMessage, error) {
	panic("not implemented!")
}

func (msg *MsgClientFence) Write(c common.IServerConn) error {
	panic("not implemented!")
}

// MsgClientCutText holds the wire format message, sans the text field.
type MsgClientCutText struct {
	_      [3]byte // padding
	Length uint32  // length
	Text   []byte
}

func (*MsgClientCutText) Type() common.ClientMessageType {
	return common.ClientCutTextMsgType
}

func (*MsgClientCutText) Read(c common.IServerConn) (common.ClientMessage, error) {
	msg := MsgClientCutText{}
	var pad [3]byte

	r, err := c.Reader()
	if err != nil {
		return nil, nil
	}

	if err := binary.Read(r, binary.BigEndian, &pad); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.BigEndian, &msg.Length); err != nil {
		return nil, err
	}

	msg.Text = make([]byte, msg.Length)
	if err := binary.Read(r, binary.BigEndian, &msg.Text); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *MsgClientCutText) Write(c common.IServerConn) error {
	data := bytes.Buffer{}
	if err := binary.Write(&data, binary.BigEndian, msg.Type()); err != nil {
		return err
	}

	var pad [3]byte
	if err := binary.Write(&data, binary.BigEndian, &pad); err != nil {
		return err
	}

	if uint32(len(msg.Text)) > msg.Length {
		msg.Length = uint32(len(msg.Text))
	}

	if err := binary.Write(&data, binary.BigEndian, msg.Length); err != nil {
		return err
	}

	if err := binary.Write(&data, binary.BigEndian, msg.Text); err != nil {
		return err
	}

	c.Write(data.Bytes())
	return nil
}

// MsgClientQemuExtendedKey holds the wire format message, for qemu keys
type MsgClientQemuExtendedKey struct {
	SubType uint8  // sub type
	IsDown  uint16 // button down indicator
	KeySym  uint32 // key symbol
	KeyCode uint32 // key code
}

func (*MsgClientQemuExtendedKey) Type() common.ClientMessageType {
	return common.QEMUExtendedKeyEventMsgType
}

func (*MsgClientQemuExtendedKey) Read(c common.IServerConn) (common.ClientMessage, error) {
	msg := MsgClientQemuExtendedKey{}
	r, err := c.Reader()
	if err != nil {
		return nil, nil
	}

	if err := binary.Read(r, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *MsgClientQemuExtendedKey) Write(c common.IServerConn) error {
	data := bytes.Buffer{}
	if err := binary.Write(&data, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(&data, binary.BigEndian, msg); err != nil {
		return err
	}
	c.Write(data.Bytes())
	return nil
}
