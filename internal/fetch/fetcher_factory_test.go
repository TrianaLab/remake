package fetch

import (
	"testing"
)

func TestGetFetcher(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		want    Fetcher
		wantErr bool
		errMsg  string
	}{
		{
			name:    "OCI fetcher",
			ref:     "oci://registry.example.com/repo:tag",
			want:    &OCIFetcher{},
			wantErr: false,
		},
		{
			name:    "HTTP fetcher",
			ref:     "http://example.com/makefile",
			want:    &HTTPFetcher{},
			wantErr: false,
		},
		{
			name:    "HTTPS fetcher",
			ref:     "https://example.com/makefile",
			want:    &HTTPFetcher{},
			wantErr: false,
		},
		{
			name:    "Unsupported protocol",
			ref:     "ftp://example.com/makefile",
			want:    nil,
			wantErr: true,
			errMsg:  "unsupported reference: ftp://example.com/makefile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFetcher(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFetcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err.Error() != tt.errMsg {
					t.Errorf("GetFetcher() error message = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			switch tt.want.(type) {
			case *OCIFetcher:
				if _, ok := got.(*OCIFetcher); !ok {
					t.Errorf("GetFetcher() = %T, want %T", got, tt.want)
				}
			case *HTTPFetcher:
				if _, ok := got.(*HTTPFetcher); !ok {
					t.Errorf("GetFetcher() = %T, want %T", got, tt.want)
				}
			}
		})
	}
}
