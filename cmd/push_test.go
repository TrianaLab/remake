package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/viper"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
)

// replace oras.Copy with a variable to allow stubbing in tests
var pushCopy = oras.Copy

func TestPushCmd_NoPrefix(t *testing.T) {
	origF, origI := pushFile, pushInsecure
	defer func() { pushFile, pushInsecure = origF, origI }()
	pushFile, pushInsecure = "Makefile", false

	err := pushCmd.RunE(pushCmd, []string{"docker://host/repo:tag"})
	if err == nil || !strings.Contains(err.Error(), "reference must start with oci://") {
		t.Fatalf("expected prefix error, got %v", err)
	}
}

func TestPushCmd_InvalidReference(t *testing.T) {
	origF, origI := pushFile, pushInsecure
	defer func() { pushFile, pushInsecure = origF, origI }()
	pushFile, pushInsecure = "Makefile", false

	err := pushCmd.RunE(pushCmd, []string{"oci://host"})
	if err == nil || !strings.Contains(err.Error(), "invalid reference: oci://host") {
		t.Fatalf("expected invalid reference error, got %v", err)
	}
}

func TestPushCmd_NoMakefileFound(t *testing.T) {
	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)

	dir := t.TempDir()
	os.Chdir(dir)

	origF, origI := pushFile, pushInsecure
	defer func() { pushFile, pushInsecure = origF, origI }()
	pushFile, pushInsecure = "", false

	err := pushCmd.RunE(pushCmd, []string{"oci://example.com/repo:tag"})
	if err == nil || !strings.Contains(err.Error(), "no Makefile found") {
		t.Fatalf("expected no Makefile found error, got %v", err)
	}
}

func TestPushCmd_CopyError(t *testing.T) {
	origCopy := pushCopy
	defer func() { pushCopy = origCopy }()
	pushCopy = func(ctx context.Context, src oras.ReadOnlyTarget, srcRef string, dst oras.Target, dstRef string, opts oras.CopyOptions) (v1.Descriptor, error) {
		return v1.Descriptor{}, fmt.Errorf("copy failed")
	}

	dir := t.TempDir()
	makefile := filepath.Join(dir, "Makefile")
	n := os.WriteFile(makefile, []byte("all:\n\techo hi\n"), 0644)
	if n != nil {
		t.Fatalf("failed to write Makefile: %v", n)
	}

	origF, origI := pushFile, pushInsecure
	defer func() { pushFile, pushInsecure = origF, origI }()
	pushFile, pushInsecure = makefile, false

	err := pushCmd.RunE(pushCmd, []string{"oci://example.com/repo:tag"})
	if err == nil || !strings.Contains(err.Error(), "failed to push artifact") {
		t.Fatalf("expected push artifact error, got %v", err)
	}
}

func TestPushCmd_Success_OCIRegistry(t *testing.T) {
	// setup fake OCI registry endpoints
	handler := http.NewServeMux()
	// ping endpoint
	handler.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	// blob uploads
	handler.HandleFunc("/v2/myrepo/blobs/uploads/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Location", r.URL.Path+"uuid")
			w.WriteHeader(http.StatusAccepted)
			return
		}
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	// manifest upload
	handler.HandleFunc("/v2/myrepo/manifests/latest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")

	// prepare workspace
	dir := t.TempDir()
	os.Chdir(dir)
	err := os.WriteFile("Makefile", []byte("all:\n\techo OK\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write Makefile: %v", err)
	}

	ctx := context.Background()
	src, err := file.New(dir)
	if err != nil {
		t.Fatalf("file.New error: %v", err)
	}
	defer src.Close()
	desc, err := src.Add(ctx, "Makefile", "application/x-makefile", "")
	if err != nil {
		t.Fatalf("src.Add error: %v", err)
	}
	manifest, err := oras.PackManifest(ctx, src, oras.PackManifestVersion1_1, "application/x-makefile", oras.PackManifestOptions{Layers: []v1.Descriptor{desc}})
	if err != nil {
		t.Fatalf("PackManifest error: %v", err)
	}
	err = src.Tag(ctx, manifest, "latest")
	if err != nil {
		t.Fatalf("Tag error: %v", err)
	}

	origF, origI := pushFile, pushInsecure
	defer func() { pushFile, pushInsecure = origF, origI }()
	pushFile, pushInsecure = filepath.Join(dir, "Makefile"), true

	if err := pushCmd.RunE(pushCmd, []string{"oci://" + host + "/myrepo:latest"}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func init() {
	viper.Reset()
}
