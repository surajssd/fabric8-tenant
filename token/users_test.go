package token

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/fabric8-services/fabric8-tenant/configuration"
)

func TestUserProfileClient_GetUserCluster(t *testing.T) {
	want := "https://fake-cluster.com"
	wantOutput := `
	{
	  "data": {
		"attributes": {
		  "cluster": "` + want + `"
		}
	  }
	}`

	tests := []struct {
		name    string
		userID  string
		want    string
		wantErr bool
		URL     string
		status  int
		output  string
	}{
		{
			name:    "normal input to see if cluster is parsed",
			want:    want,
			wantErr: false,
			userID:  "fake-userid",
		},
		{
			name:    "misformed URL",
			URL:     "google.com",
			userID:  "fake-userid",
			wantErr: true,
		},
		{
			name:    "bad status code",
			wantErr: true,
			status:  http.StatusNotFound,
			userID:  "fake-userid",
		},
		{
			name:    "make code fail on parsing output",
			wantErr: true,
			output:  "foobar",
			userID:  "fake-userid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Fatalf("Expected 'GET' request, got %q", r.Method)
				}
				path := filepath.Join("/api/users", tt.userID)
				if r.URL.EscapedPath() != path {
					t.Errorf("Expected request to %q, got %q", path, r.URL.EscapedPath())
				}

				// if no status code given in test case, set the default
				if tt.status == 0 {
					tt.status = http.StatusOK
				}
				w.WriteHeader(tt.status)

				// if the output of the server is not set in testcase, set the default
				if tt.output == "" {
					tt.output = wantOutput
				}
				w.Write([]byte(tt.output))
			}))

			// if the URL is not given in test case then set what is given by user
			if tt.URL == "" {
				tt.URL = ts.URL
			}

			config, err := configuration.GetData()
			if err != nil {
				t.Fatalf("could not retrieve configuration: %v", err)
			}

			// set the URL given by the temporary server
			os.Setenv("F8_AUTH_URL", tt.URL)

			uc := &UserProfileClient{
				Config: config,
			}
			got, err := uc.GetUserCluster(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("UserProfileClient.GetUserCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err != nil && tt.wantErr {
				t.Logf("UserProfileClient.GetUserCluster() failed with error = %v", err)
			}
			if got != tt.want {
				t.Errorf("UserProfileClient.GetUserCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}
