package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/andrelom/akira/logger"
	"github.com/andrelom/akira/server"
	"github.com/andrelom/akira/terminal"
)

const (
	Addr        = "localhost:1984"
	LogFileName = "akira.log"
)

func setupLogger() (*logger.Logger, error) {
	lgr, err := logger.NewFile(LogFileName)
	if err != nil {
		return nil, err
	}
	return lgr, nil
}

func setupServer(lgr *logger.Logger) (*server.TCPServer, error) {
	opt := &server.TCPServerOptions{
		Addr: Addr,
	}
	hdl := NewTCPServerHandler(lgr)
	srv := server.NewTCPServer(opt, hdl, lgr)
	if err := srv.Start(); err != nil {
		return nil, err
	}
	return srv, nil
}

func setupExit(srv *server.TCPServer) {
	exit := make(chan os.Signal)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	<-exit
	terminal.ClearCharacters(2)
	if err := srv.Shutdown(); err != nil {
		log.Fatal("Fatal error: ", err)
	}
	os.Exit(1)
}

func main() {
	terminal.ClearEntireScreen()
	terminal.PrintHeader()
	var err error
	lgr, err := setupLogger()
	if err != nil {
		log.Fatal("Fatal error: ", err)
		return
	}
	srv, err := setupServer(lgr)
	if err != nil {
		log.Fatal("Fatal error: ", err)
		return
	}
	setupExit(srv)
}
