package devctmpl_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mazurov/devcontainer-template/pkg/devctmpl"
)

func TestCheckTemplate(t *testing.T) {
	tests := []struct {
		name    string
		dirPath string
		wantErr bool
	}{
		{
			name:    "valid template",
			dirPath: "testdata/valid_template",
			wantErr: false,
		},
		{
			name:    "invalid template - missing files",
			dirPath: "testdata/invalid_template",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := filepath.Abs(tt.dirPath)
			if err != nil {
				t.Fatalf("failed to get absolute path: %v", err)
			}
			tempDir, err := os.MkdirTemp("", "devcontainer-test-*")
			if err != nil {
				t.Fatalf("failed to create temporary directory: %v", err)
			}
			defer os.RemoveAll(tempDir)
			err = devctmpl.GenerateTemplate(dir, tempDir, map[string]string{"imageVariant": "asd"})
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
