// The MIT License (MIT)
//
// Copyright © 2025 TrianaLab - Eduardo Diaz <edudiazasencio@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/TrianaLab/remake/config"
	"github.com/spf13/viper"
	"oras.land/oras-go/v2/registry/remote/retry"
)

type badBody struct{}
type badTransport struct{}
type bodyCloseError struct{}
type transportCloseError struct{}

func (b *badBody) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func (b *badBody) Close() error {
	return nil
}

func (t *badTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       &badBody{},
	}, nil
}

func (b *bodyCloseError) Read(p []byte) (int, error) {
	return 0, io.EOF
}
func (b *bodyCloseError) Close() error {
	return errors.New("close error")
}

func (t *transportCloseError) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       &bodyCloseError{},
	}, nil
}

func TestNewClientTypes(t *testing.T) {
	cfg := &config.Config{}
	if _, ok := NewClient(cfg, "http://x").(*HTTPClient); !ok {
		t.Error("expected HTTPClient")
	}
	if _, ok := NewClient(cfg, "repo:tag").(*OCIClient); !ok {
		t.Error("expected OCIClient")
	}
}

func TestHTTPClientPullSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := NewHTTPClient()
	data, err := client.Pull(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "ok" {
		t.Errorf("unexpected data: %s", data)
	}
}

func TestHTTPClientPullError(t *testing.T) {
	client := NewHTTPClient()
	_, err := client.Pull(context.Background(), "http://invalid.invalid")
	if err == nil {
		t.Error("expected error")
	}
}

func TestOCIClientPullBadRef(t *testing.T) {
	cfg := &config.Config{}
	client := NewOCIClient(cfg)
	_, err := client.Pull(context.Background(), "not-a-ref")
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestOCIClientLoginBadRegistry(t *testing.T) {
	cfg := &config.Config{}
	client := NewOCIClient(cfg)
	if err := client.Login(context.Background(), "://bad", "u", "p"); err == nil {
		t.Error("expected error")
	}
}

func TestNewClientLocalReferenceDefault(t *testing.T) {
	tmp, err := os.CreateTemp("", "f*.mk")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())

	cfg := &config.Config{}
	c := NewClient(cfg, tmp.Name())
	if _, ok := c.(*OCIClient); !ok {
		t.Errorf("expected default branch to return OCIClient for local ref, got %T", c)
	}
}

func TestHTTPClientLoginNoop(t *testing.T) {
	h := NewHTTPClient()
	err := h.Login(context.Background(), "any", "user", "pass")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestHTTPClientPushNoop(t *testing.T) {
	h := NewHTTPClient()
	err := h.Push(context.Background(), "http://example.com", "path")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestHTTPClientPullBadURL(t *testing.T) {
	h := NewHTTPClient()
	_, err := h.Pull(context.Background(), "%ht!tp://bad-url")
	if err == nil {
		t.Error("expected error for bad URL, got nil")
	}
}

func TestHTTPClientPullNonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	h := NewHTTPClient()
	_, err := h.Pull(context.Background(), server.URL)
	if err == nil || !strings.Contains(err.Error(), strconv.Itoa(http.StatusNotFound)) {
		t.Errorf("expected non-200 status code error, got %v", err)
	}
}

func TestHTTPClientPullReadBodyError(t *testing.T) {
	h := NewHTTPClient()
	h.httpClient = &http.Client{Transport: &badTransport{}}

	_, err := h.Pull(context.Background(), "http://any")
	if err == nil || !strings.Contains(err.Error(), "failed to read HTTP response body") {
		t.Errorf("expected read body error, got %v", err)
	}
}

func TestHTTPClientPullCloseError(t *testing.T) {
	h := NewHTTPClient()
	h.httpClient = &http.Client{Transport: &transportCloseError{}}

	data, err := h.Pull(context.Background(), "http://any")
	if data == nil {
		t.Fatalf("expected data slice, got nil")
	}
	if err == nil || err.Error() != "close error" {
		t.Errorf("expected close error, got %v", err)
	}
}

func TestOCIClientLoginSuccess(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())
	viper.SetConfigFile(tmpFile.Name())

	retry.DefaultClient = server.Client()

	cfg := &config.Config{}
	client := NewOCIClient(cfg)
	registry := server.Listener.Addr().String()

	user := "testuser"
	pass := "testpass"
	err = client.Login(context.Background(), registry, user, pass)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	key := config.NormalizeKey(registry)
	if got := viper.GetString("registries." + key + ".username"); got != user {
		t.Errorf("expected username %q, got %q", user, got)
	}
	if got := viper.GetString("registries." + key + ".password"); got != pass {
		t.Errorf("expected password %q, got %q", pass, got)
	}
}

func TestOCIClientLoginPingError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())
	viper.SetConfigFile(tmpFile.Name())

	retry.DefaultClient = server.Client()

	cfg := &config.Config{}
	client := NewOCIClient(cfg)

	reference := server.Listener.Addr().String()
	err = client.Login(context.Background(), reference, "user", "pass")
	if err == nil {
		t.Error("expected Ping error, got nil")
	}
}

// Tests for Push
func TestOCIClientPushInvalidScheme(t *testing.T) {
	cfg := &config.Config{DefaultRegistry: "example.com"}
	client := NewOCIClient(cfg)
	err := client.Push(context.Background(), "http://example.com/repo:tag", "path")
	if err == nil || !strings.Contains(err.Error(), "invalid OCI reference") {
		t.Errorf("expected invalid OCI reference error, got %v", err)
	}
}

func TestOCIClientPushParseError(t *testing.T) {
	cfg := &config.Config{DefaultRegistry: "example.com"}
	client := NewOCIClient(cfg)
	err := client.Push(context.Background(), "oci://not$$invalid/ref", "path")
	if err == nil {
		t.Error("expected parse error, got nil")
	}
}

func TestOCIClientPushMissingFile(t *testing.T) {
	cfg := &config.Config{DefaultRegistry: "example.com"}
	client := NewOCIClient(cfg)
	err := client.Push(context.Background(), "oci://example.com/myrepo:latest", "nofile")
	if err == nil || !strings.Contains(err.Error(), "adding file to store") {
		t.Errorf("expected file add error, got %v", err)
	}
}
