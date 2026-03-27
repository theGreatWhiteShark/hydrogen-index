package parser_test

import (
	"archive/tar"
	"io"
	"os"
	"testing"

	"github.com/hydrogen-music/hydrogen-index/internal/parser"
)

// openTarEntry opens a tar archive and returns an io.ReadCloser for the first
// entry whose Name matches entryPath, along with a cleanup function.
// The caller must call cleanup() when done.
func openTarEntry(t *testing.T, archivePath, entryPath string) (io.Reader, func()) {
	t.Helper()

	f, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("open archive %s: %v", archivePath, err)
	}

	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			f.Close()
			t.Fatalf("entry %q not found in %s", entryPath, archivePath)
		}
		if err != nil {
			f.Close()
			t.Fatalf("reading tar %s: %v", archivePath, err)
		}
		if hdr.Name == entryPath {
			// tr reads from f, so cleanup must close f.
			return tr, func() { f.Close() }
		}
	}
}

func TestParseDrumkit(t *testing.T) {
	t.Run("v2.0.0 drumkit from tar", func(t *testing.T) {
		r, cleanup := openTarEntry(t,
			"../../res/hydrogen-artifacts/v2.0.0.h2drumkit",
			"testKit/drumkit.xml",
		)
		defer cleanup()

		got, err := parser.ParseDrumkit(r)
		if err != nil {
			t.Fatalf("ParseDrumkit: %v", err)
		}

		checkString(t, "Name", got.Name, "testKit")
		checkString(t, "Author", got.Author, "theGreatWhiteShark")
		checkString(t, "License", got.License, "MIT")
		checkInt(t, "FormatVersion", got.FormatVersion, 2)
		checkInt(t, "UserVersion", got.UserVersion, 0)
		checkStrings(t, "Tags", got.Tags, []string{"Example", "Drumkit"})
		checkInt(t, "Instruments", got.Instruments, 3)
		// One "Main" component referenced by all 3 instruments — 1 unique name.
		checkInt(t, "Components", got.Components, 1)
		checkInt(t, "Samples", got.Samples, 3)
		// All three instruments have empty <type>; InstrumentTypes must be empty, not nil.
		checkStrings(t, "InstrumentTypes", got.InstrumentTypes, []string{})
	})

	t.Run("v1.2.3 drumkit with componentList", func(t *testing.T) {
		f, err := os.Open("../../res/hydrogen-artifacts/legacy-drumkits/kit-1.2.3/drumkit.xml")
		if err != nil {
			t.Fatalf("open fixture: %v", err)
		}
		defer f.Close()

		got, err := parser.ParseDrumkit(f)
		if err != nil {
			t.Fatalf("ParseDrumkit: %v", err)
		}

		checkString(t, "Name", got.Name, "Legacy-1.2.3")
		checkString(t, "Author", got.Author, "Hydrogen Dev")
		checkString(t, "License", got.License, "CC0")
		checkInt(t, "FormatVersion", got.FormatVersion, 0)
		checkStrings(t, "Tags", got.Tags, []string{})
		checkInt(t, "Instruments", got.Instruments, 1)
		checkInt(t, "Components", got.Components, 2)
		checkInt(t, "Samples", got.Samples, 2)
		checkStrings(t, "InstrumentTypes", got.InstrumentTypes, []string{})
	})

	t.Run("v0.9.3 legacy drumkit", func(t *testing.T) {
		f, err := os.Open("../../res/hydrogen-artifacts/legacy-drumkits/kit-0.9.3/drumkit.xml")
		if err != nil {
			t.Fatalf("open fixture: %v", err)
		}
		defer f.Close()

		got, err := parser.ParseDrumkit(f)
		if err != nil {
			t.Fatalf("ParseDrumkit: %v", err)
		}

		checkString(t, "Name", got.Name, "kit-0.9.3")
		checkString(t, "Author", got.Author, "Hydrogen dev team")
		checkString(t, "Info", got.Info, "Mock Drumkit derived from one created (probably) using 0.9.3 or previous. Licensed as CC-0.")
		checkInt(t, "FormatVersion", got.FormatVersion, 0)
		checkStrings(t, "Tags", got.Tags, []string{})
		checkInt(t, "Instruments", got.Instruments, 3)
		checkInt(t, "Components", got.Components, 0)
		// Kick: 4 layers, Crash: 2 layers, "32": 0 layers = 6 total (verified from fixture).
		checkInt(t, "Samples", got.Samples, 6)
		checkStrings(t, "InstrumentTypes", got.InstrumentTypes, []string{})
	})
}

func checkString(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %q, want %q", field, got, want)
	}
}

func checkInt(t *testing.T, field string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %d, want %d", field, got, want)
	}
}

func checkStrings(t *testing.T, field string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: got %v (len=%d), want %v (len=%d)", field, got, len(got), want, len(want))
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s[%d]: got %q, want %q", field, i, got[i], want[i])
		}
	}
}
