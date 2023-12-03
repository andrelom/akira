package main

import (
	"context"

	"github.com/andrelom/akira/logger"
	"github.com/andrelom/akira/protocol"
)

type TCPServerHandler struct {
	lgr *logger.Logger
}

func NewTCPServerHandler(lgr *logger.Logger) *TCPServerHandler {
	return &TCPServerHandler{
		lgr: lgr,
	}
}

func (hdl *TCPServerHandler) Handle(req *protocol.Message, ctx context.Context) (*protocol.Message, error) {
	res := &protocol.Message{
		Type: "PONG",
		Body: "",
	}
	return res, nil
}
