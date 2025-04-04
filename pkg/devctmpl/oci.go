package devctmpl

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// IsNotOCIRepository determines if the given source is NOT an OCI repository
func isOCIRepository(source string) bool {
	if _, err := os.Stat(source); err == nil {
		return false
	}

	// Check if the source contains URL-specific characters
	if strings.Contains(source, "//") || strings.Contains(source, "?") {
		return false
	}

	_, err := name.ParseReference(source)
	return err == nil
}

func pullOCITemplate(reference string, destDir string) error {
	// Parse the reference
	ref, err := name.ParseReference(reference)
	if err != nil {
		return fmt.Errorf("invalid reference %q: %w", reference, err)
	}

	// Pull the image
	img, err := remote.Image(ref, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Get all layers
	layers, err := img.Layers()
	if err != nil {
		return fmt.Errorf("failed to get layers: %w", err)
	}

	// Extract each layer
	for _, layer := range layers {
		// Get layer content
		rc, err := layer.Uncompressed()
		if err != nil {
			return fmt.Errorf("failed to get layer content: %w", err)
		}
		defer rc.Close()

		// Extract the layer
		if err := extractTar(rc, destDir); err != nil {
			return fmt.Errorf("failed to extract layer: %w", err)
		}
	}

	return nil
}

func extractTar(r io.Reader, dest string) error {
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}
