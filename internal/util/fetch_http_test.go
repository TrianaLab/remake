package util

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFetchHTTP(t *testing.T) {
	content := "data"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer srv.Close()
	path, err := FetchHTTP(srv.URL + "/x.mk")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	d, _ := os.ReadFile(path)
	if string(d) != content {
		t.Errorf("got %s", d)
	}
}
