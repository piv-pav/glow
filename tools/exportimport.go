package tools

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/piv-pav/glow/internal/article"
	"github.com/piv-pav/glow/internal/config"
	"github.com/piv-pav/glow/internal/storage"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [wiki] [output.tar.gz]",
	Short: "Export a wiki to a tar.gz archive of markdown files",
	Args:  cobra.ExactArgs(2),
	RunE:  runExport,
}

var importCmd = &cobra.Command{
	Use:   "import [wiki] [input.tar.gz]",
	Short: "Import articles from a tar.gz archive into a wiki",
	Args:  cobra.ExactArgs(2),
	RunE:  runImport,
}

func runExport(cmd *cobra.Command, args []string) error {
	wikiName, outPath := args[0], args[1]

	exists, err := config.WikiExists(wikiName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("wiki not found: %s", wikiName)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	n := 0
	err = withStore(wikiName, func(store storage.Store) error {
		names, err := store.List()
		if err != nil {
			return err
		}
		for _, name := range names {
			art, err := store.Read(name)
			if err != nil {
				return fmt.Errorf("failed to read %q: %w", name, err)
			}
			data, err := art.Serialize()
			if err != nil {
				return fmt.Errorf("failed to serialize %q: %w", name, err)
			}
			hdr := &tar.Header{
				Name:    name + ".md",
				Mode:    0644,
				Size:    int64(len(data)),
				ModTime: time.Now(),
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}
			if _, err := tw.Write(data); err != nil {
				return err
			}
			n++
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("Exported %d articles from wiki %q to %s\n", n, wikiName, outPath)
	return nil
}

func runImport(cmd *cobra.Command, args []string) error {
	wikiName, inPath := args[0], args[1]

	exists, err := config.WikiExists(wikiName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("wiki not found: %s — create it first with: glow init %s", wikiName, wikiName)
	}

	f, err := os.Open(inPath)
	if err != nil {
		return fmt.Errorf("failed to open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to read gzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	n := 0
	skipped := 0
	err = withStore(wikiName, func(store storage.Store) error {
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read archive: %w", err)
			}

			if !strings.HasSuffix(hdr.Name, ".md") {
				continue
			}

			data, err := io.ReadAll(tr)
			if err != nil {
				return fmt.Errorf("failed to read %q: %w", hdr.Name, err)
			}

			art, err := article.Parse(data)
			if err != nil {
				return fmt.Errorf("failed to parse %q: %w", hdr.Name, err)
			}

			name := strings.TrimSuffix(filepath.ToSlash(hdr.Name), ".md")

			if err := store.Create(name, art); err != nil {
				// Skip already-existing articles, warn and continue
				fmt.Fprintf(os.Stderr, "skipped %q: %v\n", name, err)
				skipped++
				continue
			}
			n++
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("Imported %d articles into wiki %q", n, wikiName)
	if skipped > 0 {
		fmt.Printf(" (%d skipped — already exist)", skipped)
	}
	fmt.Println()
	return nil
}
