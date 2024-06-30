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

	"github.com/rs/zerolog"
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

var logger = zerolog.New(os.Stdout)

func init() {
	flag.BoolVar(&showVersion, "v", false, "Print version")
	flag.StringVar(&configPath, "c", "config.json", "Path to your config")
	flag.StringVar(&logLevelString, "l", "info", "Log level")

	flag.Parse()
	if showVersion {
		printVersion()
		os.Exit(0)
	}

	logLevel, err := zerolog.ParseLevel(logLevelString)
	if err != nil {
		zerolog.DefaultContextLogger.Fatal().Msgf("parse log level: %v", err)
	}
	logger = logger.Level(logLevel)
}

func main() {
	file, err := os.Open(configPath)
	if err != nil {
		logger.Fatal().Msgf("open config: %v", err)
	}
	defer file.Close()
	jsonDecoder := json.NewDecoder(file)
	var config server
	err = jsonDecoder.Decode(&config)
	if err != nil {
		logger.Fatal().Msgf("decode json: %v", err)
	}
	logger.Info().Msgf("config: %v", config)

	switch config.Log {
	case "", "stdout":
		logger = logger.Output(os.Stdout)
	case "stderr":
		logger = logger.Output(os.Stderr)
	case "null", "none", "ignore":
		logger = logger.Output(io.Discard)
	default:
		logFile, err := os.Open(config.Log)
		if err != nil {
			logger.Fatal().Msgf("open log: %v", err)
		}
		defer logFile.Close()
		logger = logger.Output(logFile)
	}

	tcpListener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		logger.Fatal().Msgf("listen: %v", err)
	}
	defer tcpListener.Close()

	var listener net.Listener
	if config.Cert != "" {
		_, err = tls.LoadX509KeyPair(config.Cert, config.Key)
		if err != nil {
			logger.Fatal().Msgf("load x509 key pair: %v", err)
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

	err = http.Serve(listener, newSIP008Handler(&logger))
	if err != nil {
		logger.Fatal().Msgf("serve: %v", err)
	}
}
