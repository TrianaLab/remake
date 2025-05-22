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
// THE SOFTWARE.

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TrianaLab/remake/config"
)

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
