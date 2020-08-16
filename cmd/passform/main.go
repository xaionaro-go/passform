package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/xaionaro-go/passform/pkg/ldap"
	"github.com/xaionaro-go/passform/pkg/webui"
)

func usageError(description string) {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "error: %s\n", description)
	flag.Usage()
	os.Exit(2)
}

func assertNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	netPprof := flag.String("net-pprof", "", "")
	webUIBindAddress := flag.String("web-ui-bind-address", "tcp:127.0.0.1:8080", "")
	ldapAddr := flag.String("ldap-address", "ldap://127.0.0.1:389", "")
	ldapBindDN := flag.String("ldap-bind-dn", "", "")
	ldapBindPass := flag.String("ldap-bind-pass", "", "")
	logLevel := flag.String("log-level", "warning", "")
	flag.Parse()

	if *webUIBindAddress == "" {
		usageError("empty Web UI bind address")
	}
	if !strings.HasPrefix(*webUIBindAddress, "tcp:") {
		usageError("currently only TCP listeners are supported to Web UI")
	}
	*webUIBindAddress = (*webUIBindAddress)[len("tcp:"):]

	if *ldapAddr == "" {
		usageError("empty LDAP address")
	}

	if *ldapBindDN == "" {
		usageError("empty LDAP bind DN")
	}

	if *ldapBindPass == "" {
		usageError("empty LDAP bind password")
	}

	lvl, err := logrus.ParseLevel(*logLevel)
	assertNoError(err)

	loggerRaw := logrus.New()
	loggerRaw.Level = lvl
	logger := logrus.NewEntry(loggerRaw).WithField("module", "main")

	if *netPprof != "" {
		go func() {
			logger := logger.WithField("module", "pprof")
			logger.Tracef("starting net pprof at '%s'", *netPprof)
			logger.Error(http.ListenAndServe(*netPprof, nil))
		}()
	}

	logger.Debugf("starting LDAP connector")
	ldapConnector, err := ldap.NewConnector(*ldapAddr, logger.WithField("module", "ldap"))
	assertNoError(err)

	logger.Debugf("starting Web UI")
	uiServer := webui.NewServer(
		*webUIBindAddress, *ldapBindDN, *ldapBindPass, ldapConnector,
		logger.WithField("module", "webui"),
	)

	logger.Debugf("starting uiServer.Serve")
	err = uiServer.Serve(context.Background())
	assertNoError(err)
}
