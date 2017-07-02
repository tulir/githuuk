// githuuk - A GitHub webhook receiver written in Go.
// Copyright (C) 2017 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package hookserve

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Version is the version of Hookserve.
var Version = "0.1.0"

// Server is the main server container.
type Server struct {
	Host     string
	Port     int
	Path     string
	PingPath string
	Secret   string
	Events   chan Event
}

// NewServer creates a webhook server with sensible defaults.
// Default settings:
//   Port: 80
//   Path: /webhook
//   Ping path: /webhook/ping
//   Ignore tags: true
func NewServer() *Server {
	return &Server{
		Port:     80,
		Path:     "/webhook",
		PingPath: "/webhook/ping",
		Events:   make(chan Event, 10),
	}
}

// ListenAndServe runs the server and returns if an error occurs.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.Host, s.Port), s)
}

// AsyncListenAndServe runs the server inside a Goroutine and panics if an error occurs.
func (s *Server) AsyncListenAndServe() {
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()
}

// CheckSignature checks that the request is from GitHub.
func (s *Server) CheckSignature(w http.ResponseWriter, r *http.Request, body []byte) bool {
	if s.Secret != "" {
		sig := r.Header.Get("X-Hub-Signature")

		if sig == "" {
			http.Error(w, "403 Forbidden - Missing X-Hub-Signature required for HMAC verification", http.StatusForbidden)
			return false
		}

		mac := hmac.New(sha1.New, []byte(s.Secret))
		mac.Write(body)
		expectedMAC := mac.Sum(nil)
		expectedSig := "sha1=" + hex.EncodeToString(expectedMAC)
		if !hmac.Equal([]byte(expectedSig), []byte(sig)) {
			http.Error(w, "403 Forbidden - HMAC verification failed", http.StatusForbidden)
			return false
		}
	}
	return true
}

// ServeHTTP implements the http.Handler interface.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method == "GET" && r.URL.Path == s.PingPath {
		w.Write([]byte("200 OK"))
		return
	} else if r.URL.Path != s.Path {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	} else if r.Method != "POST" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	eventType := EventType(r.Header.Get("X-GitHub-Event"))
	if eventType == "" {
		http.Error(w, "400 Bad Request - Missing X-GitHub-Event Header", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !s.CheckSignature(w, r, body) {
		return
	}

	var event Event
	if eventType == EventPush {
		event = PushEvent{}
		err = json.Unmarshal(body, &event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if eventType == EventPullRequest {
		event = PullRequestEvent{}
		err = json.Unmarshal(body, &event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, fmt.Sprintf("501 Not Implemented - Unknown event type %s", eventType), http.StatusNotImplemented)
		return
	}
	s.Events <- event

	w.Header().Set("Server", "hookserve/"+Version)
	w.Write([]byte("{}"))
}
