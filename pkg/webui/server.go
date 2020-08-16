package webui

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/xaionaro-go/passform/pkg/ldap"
)

var (
	ErrAlreadyUsed = errors.New("already used")
)

type Server struct {
	BindDN string
	BindPassword string
	Logger *logrus.Entry
	ldapConnector  *ldap.Connector
	httpServer     *http.Server
	httpServerOnce sync.Once
}

func NewServer(
	bindAddress, bindDN, bindPassword string,
	ldapConnector *ldap.Connector,
	logger *logrus.Entry,
) *Server {
	srv := &Server{
		BindDN: bindDN,
		BindPassword: bindPassword,
		Logger: logger,
		ldapConnector: ldapConnector,
	}
	srv.setupHttpServer(bindAddress)
	return srv
}

func (srv *Server) setupHttpServer(bindAddress string) {
	serveMux := http.NewServeMux()
	serveMux.Handle("/", changePassForm{Server: srv})
	serveMux.Handle("/register", registerForm{Server: srv})
	srv.httpServer = &http.Server{
		Addr:              bindAddress,
		Handler:           serveMux,
	}
}

func (srv *Server) Serve(ctx context.Context) error {
	err := ErrAlreadyUsed
	srv.httpServerOnce.Do(func() {
		err = srv.serve(ctx)
	})
	return err
}

func (srv *Server) serve(ctx context.Context) error {
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()
	go func() {
		<-ctx.Done()
		_ = srv.httpServer.Close()
	}()

	ln, err := net.Listen("tcp", srv.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("unable to start listening '%s': %w", srv.httpServer.Addr, err)
	}

	srv.Logger.Debugf("started listening at '%s'", srv.httpServer.Addr)
	err = srv.httpServer.Serve(ln)
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("unable to start serving: %w", err)
	}
	return nil
}
