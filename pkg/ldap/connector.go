package ldap

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

type Connector struct {
	ServerAddress string
	Logger *logrus.Entry
}

func NewConnector(serverAddress string, entry *logrus.Entry) (*Connector, error) {
	return &Connector{
		ServerAddress: serverAddress,
		Logger: entry,
	}, nil
}

func (ctrl *Connector) AcquireConn() (*Conn, error) {
	ctrl.Logger.Tracef("AcquireConn")
	defer ctrl.Logger.Tracef("/AcquireConn")

	l, err := ctrl.newConn()
	if err != nil {
		return nil, fmt.Errorf("unable to create a connection to LDAP '%s': %w",
			ctrl.ServerAddress, err)
	}

	return l, nil
}

func (ctrl *Connector) newConn() (*Conn, error) {
	l := &Conn{
		Connector: ctrl,
	}
	err := l.initConn()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to LDAP: %w", err)
	}
	return l, nil
}
