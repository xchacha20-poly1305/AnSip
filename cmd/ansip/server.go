package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/xchacha20-poly1305/ansip"
	"github.com/xchacha20-poly1305/ansip/internal/freeauth"
)

// server is server config.
type server struct {
	Listen string `json:"listen"`
	Log    string `json:"log"`

	// TLS
	Cert       string `json:"cert"`
	Key        string `json:"key"`
	ServerName string `json:"server_name"`
}

var _ http.Handler = (*sip008Handler)(nil)

type sip008Handler struct {
	auth *freeauth.FreeAuth
	// [string]SIP008
	data   sync.Map
	logger *log.Logger
}

func newSIP008Handler(logger *log.Logger) *sip008Handler {
	return &sip008Handler{
		auth:   freeauth.NewFreeAuth(),
		logger: logger,
	}
}

func (s *sip008Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	s.logger.Debugf("http for: %s", request.RemoteAddr)

	/*contentType := request.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		writer.WriteHeader(http.StatusUnsupportedMediaType)
		logger.Warn().Msgf("not support Content-Type: %s", contentType)
		return
	}*/

	username, password, success := authFromHeader(request.Header)
	if !success {
		writer.WriteHeader(http.StatusUnauthorized)
		s.logger.Warnf("not found auth for: %v", request.RemoteAddr)
		return
	}
	success, isNew := s.auth.Verify(username, password)
	if !success {
		writer.WriteHeader(http.StatusForbidden)
		s.logger.Warnf("failed to verify: %s:%s", username, password)
		return
	}
	if isNew {
		s.logger.Infof("add new user: %s", username)
	}

	switch request.Method {
	case http.MethodPost:
		var data ansip.SIP008
		err := json.NewDecoder(request.Body).Decode(&data)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			s.logger.Warnf("failed to decode: %v", err)
			return
		}
		if data.Version != ansip.SIP008Version {
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte("invalid version"))
			s.logger.Warnf("invalid version: %d", data.Version)
			return
		}
		s.data.Store(username, data)
		return
	case http.MethodGet:
		data, ok := s.data.Load(username)
		if !ok {
			writer.WriteHeader(http.StatusNotFound)
			s.logger.Warnf("not found data for %s", username)
			return
		}
		err := json.NewEncoder(writer).Encode(data)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			s.logger.Warnf("failed to encode: %v", err)
			return
		}
		return
	case http.MethodDelete:
		err := s.auth.Delete(username, password)
		if err != nil {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		s.data.Delete(username)
		return
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func authFromHeader(header http.Header) (username, password string, success bool) {
	auth, found := strings.CutPrefix(header.Get("Authorization"), "Basic ")
	if !found || auth == "" {
		return "", "", false
	}

	raw, err := base64.StdEncoding.DecodeString(auth)
	if err == nil && len(raw) > 0 {
		auth = string(raw)
	}

	auths := strings.SplitN(auth, ":", 2)
	if len(auths) != 2 {
		return "", "", false
	}
	return auths[0], auths[1], true
}
