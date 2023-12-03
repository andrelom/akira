package server

import (
	"bufio"
	"context"
	"encoding/json"

	"github.com/andrelom/akira/protocol"
)

func (srv *TCPServer) handle(reader *bufio.Reader, writer *bufio.Writer, ctx context.Context) error {
	req, err := srv.read(reader)
	if err != nil {
		return err
	}
	res, err := srv.hdl.Handle(req, ctx)
	if err != nil {
		return err
	}
	return srv.respond(res, writer)
}

func (srv *TCPServer) read(reader *bufio.Reader) (*protocol.Message, error) {
	val, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	var msg protocol.Message
	err = json.Unmarshal(val, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (srv *TCPServer) respond(msg *protocol.Message, writer *bufio.Writer) error {
	byt, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = writer.Write(append(byt, '\n'))
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}
