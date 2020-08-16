package ldap

import (
	"fmt"
	"sync"

	"github.com/go-ldap/ldap/v3"
)

type Conn struct {
	*Connector

	ldapLocker sync.Mutex
	ldapConn *ldap.Conn
}

func (l *Conn) Request(fn func(conn *ldap.Conn) error) error {
	l.Logger.Tracef("Request")
	defer l.Logger.Tracef("/Request")
	l.ldapLocker.Lock()
	defer l.ldapLocker.Unlock()

	l.Logger.Tracef("fn(l.ldapConn)")
	err := fn(l.ldapConn)
	if err == nil {
		l.Logger.Tracef("Request: success")
		return nil
	}
	l.Logger.Warnf("error: %v", err)

	// Retry:

	l.ldapConn.Close()
	l.ldapConn = nil
	err = l.initConn()
	if err != nil {
		return fmt.Errorf("unable to re-connect to ldap: %w", err)
	}

	l.Logger.Tracef("fn(l.ldapConn) again")
	err = fn(l.ldapConn)
	if err != nil {
		return fmt.Errorf("unable to perform a LDAP operation: %w", err)
	}
	l.Logger.Warnf("error: %v", err)

	return nil
}

func (l *Conn) Release() {
	go func() {
		l.ldapConn.Close()
	}()
}

func (l *Conn) initConn() error {
	conn, err := ldap.DialURL(l.ServerAddress)
	if err != nil {
		return fmt.Errorf("unable to connect to %s: %w", l.ServerAddress, err)
	}

	l.ldapConn = conn
	return nil
}
