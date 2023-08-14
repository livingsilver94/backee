package privileged

import (
	"encoding/gob"
	"io"
)

type WriterEncoder interface {
	EncodeWriter(w io.Writer) error
}

type ReaderDecoder interface {
	DecodeReader(r io.Reader) error
}

type Action int

const (
	LinkFile Action = iota + 1
	CopyFile
	WriteContent
)

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
