package privileged

import (
	"encoding/gob"
	"io"
)

type Message interface {
	EncodeWriter(w io.Writer) error
	DecodeReader(r io.Reader) error
}

type action int

const (
	linkFile action = iota + 1
	copyFile
	writeContent
)

func (a action) properMessage() Message {
	switch a {
	case linkFile, copyFile:
		return &SourceDestMessage{}
	case writeContent:
		return &WriteContentMessage{}
	}
	return nil
}

type SourceDestMessage struct {
	Source string
	Dest   string
}

func (m *SourceDestMessage) EncodeWriter(w io.Writer) error {
	enc := gob.NewEncoder(w)
	return enc.Encode(m)
}

func (m *SourceDestMessage) DecodeReader(r io.Reader) error {
	dec := gob.NewDecoder(r)
	return dec.Decode(&m)
}

type WriteContentMessage struct {
	Dest    string
	Content io.Reader
}

func (m *WriteContentMessage) EncodeWriter(w io.Writer) error {
	enc := gob.NewEncoder(w)
	err := enc.Encode(&m.Dest)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, m.Content)
	return err
}

func (m *WriteContentMessage) DecodeReader(r io.Reader) error {
	dec := gob.NewDecoder(r)
	err := dec.Decode(&m.Dest)
	if err != nil {
		return err
	}
	m.Content = r
	return err
}

type Command struct {
	action  action
	message Message
}

func NewLinkFileCommand(msg *SourceDestMessage) Command {
	return Command{
		action:  linkFile,
		message: msg,
	}
}

func NewCopyFileCommand(msg *SourceDestMessage) Command {
	return Command{
		action:  copyFile,
		message: msg,
	}
}

func NewWriteContentMessage(msg *WriteContentMessage) Command {
	return Command{
		action:  writeContent,
		message: msg,
	}
}

func ReceiveCommand(r io.Reader) (Command, error) {
	var cmd Command
	err := gob.NewDecoder(r).Decode(&cmd.action)
	if err != nil {
		return cmd, err
	}

	cmd.message = cmd.action.properMessage()
	return cmd, cmd.message.DecodeReader(r)
}

func (c Command) Send(w io.Writer) error {
	err := gob.NewEncoder(w).Encode(c.action)
	if err != nil {
		return err
	}
	return c.message.EncodeWriter(w)
}
