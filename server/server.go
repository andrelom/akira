package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/andrelom/akira/logger"
	"github.com/andrelom/akira/protocol"
)

const (
	TotalWorkers      = 16
	QueueSize         = 1024
	ShutdownTimeout   = 5 * time.Second
	ConnectionTimeout = 5 * time.Second
	ConnectionIOSize  = 4096
)

type TCPServerHandler interface {
	Handle(msg *protocol.Message, ctx context.Context) (*protocol.Message, error)
}

type TCPServerOptions struct {
	Addr string
}

type TCPServer struct {
	opt *TCPServerOptions
	hdl TCPServerHandler
	lgr *logger.Logger

	started  bool
	ctx      context.Context
	cancel   context.CancelFunc
	queue    chan net.Conn
	listener net.Listener
	wg       sync.WaitGroup
}

func NewTCPServer(opt *TCPServerOptions, hdl TCPServerHandler, lgr *logger.Logger) *TCPServer {
	return &TCPServer{
		opt: opt,
		hdl: hdl,
		lgr: lgr,
	}
}

func (srv *TCPServer) Start() error {
	if srv.started {
		return fmt.Errorf("server is already started")
	}
	srv.started = true
	srv.ctx, srv.cancel = context.WithCancel(context.Background())
	srv.queue = make(chan net.Conn, QueueSize)
	if err := srv.startListener(); err != nil {
		return err
	}
	srv.startWorkerPool()
	go srv.startAcceptingConnections()
	srv.lgr.Info("Server is listening at %s", srv.opt.Addr)
	return nil
}

func (srv *TCPServer) Shutdown() error {
	if !srv.started {
		return nil
	}
	defer func() {
		close(srv.queue)
		srv.started = false
		srv.ctx = nil
		srv.cancel = nil
		srv.queue = nil
		srv.listener = nil
	}()
	srv.lgr.Info("Shutting down server...")
	srv.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	if err := srv.waitWorkersToFinish(ctx); err != nil {
		return fmt.Errorf("waiting for workers to finish: %v", err)
	}
	if err := srv.listener.Close(); err != nil {
		return fmt.Errorf("closing listener: %v", err)
	}
	srv.lgr.Info("Shutdown process completed gracefully")
	return nil
}

func (srv *TCPServer) startListener() error {
	listener, err := net.Listen("tcp", srv.opt.Addr)
	if err != nil {
		return err
	}
	srv.listener = listener
	return nil
}

func (srv *TCPServer) startWorkerPool() {
	for idx := 0; idx < TotalWorkers; idx++ {
		srv.wg.Add(1)
		go srv.startWorker()
	}
}

func (srv *TCPServer) startWorker() {
	defer srv.wg.Done()
	for {
		select {
		case <-srv.ctx.Done():
			return
		case conn := <-srv.queue:
			go srv.processIncomingConnection(conn)
		}
	}
}

func (srv *TCPServer) waitWorkersToFinish(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		srv.wg.Wait()
		close(done)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (srv *TCPServer) startAcceptingConnections() {
	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			select {
			case <-srv.ctx.Done():
				return
			default:
				srv.lgr.Error("Accepting connection: %v", err)
				continue
			}
		}
		srv.queue <- conn
	}
}

func (srv *TCPServer) processIncomingConnection(conn net.Conn) {
	srv.wg.Add(1)
	done := make(chan error, 1)
	ctx, cancel := context.WithTimeout(srv.ctx, ConnectionTimeout)
	defer cancel()
	defer conn.Close()
	defer srv.wg.Done()
	reader := bufio.NewReaderSize(conn, ConnectionIOSize)
	writer := bufio.NewWriterSize(conn, ConnectionIOSize)
	go func() {
		done <- srv.handle(reader, writer, ctx)
	}()
	select {
	case <-ctx.Done():
		srv.lgr.Warn("Connection timed out for %s", conn.RemoteAddr())
	case err := <-done:
		if err != nil {
			srv.lgr.Error("Error handling connection for %s: %v", conn.RemoteAddr(), err)
		}
	}
}
