package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/charmbracelet/log"
)

var version = "unknown"

func printVersion() {
	fmt.Printf("Version: %s\n", version)
}

var (
	showVersion bool

	configPath     string
	logLevelString string
)

var logger = log.Default()

func init() {
	flag.BoolVar(&showVersion, "v", false, "Print version")
	flag.StringVar(&configPath, "c", "config.json", "Path to your config")
	flag.StringVar(&logLevelString, "l", "info", "Log level")

	flag.Parse()
	if showVersion {
		printVersion()
		os.Exit(0)
	}

	logLevel, err := log.ParseLevel(logLevelString)
	if err != nil {
		logger.Fatalf("parse log level: %v", err)
	}
	logger.SetLevel(logLevel)
}

func main() {
	file, err := os.Open(configPath)
	if err != nil {
		logger.Fatalf("open config: %v", err)
	}
	defer file.Close()
	jsonDecoder := json.NewDecoder(file)
	var config server
	err = jsonDecoder.Decode(&config)
	if err != nil {
		logger.Fatalf("decode json: %v", err)
	}

	switch config.Log {
	case "", "stdout":
		logger.SetOutput(os.Stdout)
	case "stderr":
		logger.SetOutput(os.Stderr)
	case "null", "none", "ignore":
		logger.SetOutput(io.Discard)
	default:
		logFile, err := os.Open(config.Log)
		if err != nil {
			logger.Fatalf("open log: %v", err)
		}
		defer logFile.Close()
		logger.SetOutput(logFile)
	}
	logger.Infof("config: %v", config)

	tcpListener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		logger.Fatalf("listen: %v", err)
	}
	defer tcpListener.Close()

	var listener net.Listener
	if config.Cert != "" {
		_, err = tls.LoadX509KeyPair(config.Cert, config.Key)
		if err != nil {
			logger.Fatalf("load x509 key pair: %v", err)
		}
		tlsConfig := tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				cert, err := tls.LoadX509KeyPair(config.Cert, config.Key)
				if err != nil {
					return nil, err
				}
				return &cert, nil
			},
			MinVersion: tls.VersionTLS13,
			MaxVersion: tls.VersionTLS13,
			ServerName: config.ServerName,
		}
		listener = tls.NewListener(tcpListener, &tlsConfig)
		defer listener.Close()
	} else {
		listener = tcpListener
	}

	err = http.Serve(listener, newSIP008Handler(logger))
	if err != nil {
		logger.Fatalf("serve: %v", err)
	}
}
