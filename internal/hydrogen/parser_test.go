package hydrogen

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"reflect"
	"testing"
)

func TestParsePatternFileSupportsCurrentAndLegacyFormats(t *testing.T) {
	t.Parallel()

	t.Run("current", func(t *testing.T) {
		t.Parallel()

		path := "/home/phil/git/hydrogen-index/res/hydrogen-artifacts/v2.0.0.h2pattern"
		artifact, err := ParsePatternFile(path, buildParseOptions(t, path, "https://example.com/assets/v2.0.0.h2pattern"))
		if err != nil {
			t.Fatalf("ParsePatternFile() returned error: %v", err)
		}

		if artifact.Name != "pat" || artifact.Author != "Hydrogen dev team" {
			t.Fatalf("unexpected common metadata: %+v", artifact)
		}

		if artifact.FormatVersion != 2 || artifact.Version != 0 || artifact.Notes != 20 {
			t.Fatalf("unexpected numeric metadata: %+v", artifact)
		}

		if got, want := artifact.Tags, []string{"Example", "Pattern"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected tags: got %#v want %#v", got, want)
		}

		if got, want := artifact.InstrumentTypes, []string{"Hand Clap", "Kick", "Snare", "Stick"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected instrument types: got %#v want %#v", got, want)
		}
	})

	t.Run("legacy", func(t *testing.T) {
		t.Parallel()

		path := "/home/phil/git/hydrogen-index/res/hydrogen-artifacts/legacy-patterns/legacy_pattern.h2pattern"
		artifact, err := ParsePatternFile(path, buildParseOptions(t, path, "https://example.com/assets/legacy_pattern.h2pattern"))
		if err != nil {
			t.Fatalf("ParsePatternFile() returned error: %v", err)
		}

		if artifact.Name != "Demo 1" {
			t.Fatalf("unexpected name: %+v", artifact)
		}

		if artifact.FormatVersion != 1 || artifact.Version != 0 || artifact.Notes != 21 {
			t.Fatalf("unexpected numeric metadata: %+v", artifact)
		}

		if got, want := artifact.Tags, []string{"demo-songs"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected tags: got %#v want %#v", got, want)
		}

		if len(artifact.InstrumentTypes) != 0 {
			t.Fatalf("legacy pattern should not invent instrument types: %+v", artifact)
		}
	})
}

func TestParseSongFileSupportsCurrentAndLegacyFormats(t *testing.T) {
	t.Parallel()

	t.Run("current", func(t *testing.T) {
		t.Parallel()

		path := "/home/phil/git/hydrogen-index/res/hydrogen-artifacts/v2.0.0.h2song"
		artifact, err := ParseSongFile(path, buildParseOptions(t, path, "https://example.com/assets/v2.0.0.h2song"))
		if err != nil {
			t.Fatalf("ParseSongFile() returned error: %v", err)
		}

		if artifact.Name != "Untitled Song" || artifact.Author != "hydrogen" {
			t.Fatalf("unexpected common metadata: %+v", artifact)
		}

		if artifact.FormatVersion != 2 || artifact.Version != 0 || artifact.Patterns != 10 {
			t.Fatalf("unexpected numeric metadata: %+v", artifact)
		}

		if got, want := artifact.Tags, []string{"Example", "Song"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected tags: got %#v want %#v", got, want)
		}
	})

	t.Run("legacy", func(t *testing.T) {
		t.Parallel()

		path := "/home/phil/git/hydrogen-index/res/hydrogen-artifacts/legacy-songs/test_song_1.2.2.h2song"
		artifact, err := ParseSongFile(path, buildParseOptions(t, path, "https://example.com/assets/test_song_1.2.2.h2song"))
		if err != nil {
			t.Fatalf("ParseSongFile() returned error: %v", err)
		}

		if artifact.Name != "Untitled Song" || artifact.Author != "hydrogen" {
			t.Fatalf("unexpected common metadata: %+v", artifact)
		}

		if artifact.FormatVersion != 1 || artifact.Version != 0 || artifact.Patterns != 10 {
			t.Fatalf("unexpected numeric metadata: %+v", artifact)
		}

		if len(artifact.Tags) != 0 {
			t.Fatalf("legacy song should not invent tags: %+v", artifact)
		}
	})
}

func TestParseDrumkitSupportsArchiveAndLegacyXMLFormats(t *testing.T) {
	t.Parallel()

	t.Run("archive", func(t *testing.T) {
		t.Parallel()

		path := "/home/phil/git/hydrogen-index/res/hydrogen-artifacts/v2.0.0.h2drumkit"
		artifact, err := ParseDrumkitArchive(path, buildParseOptions(t, path, "https://example.com/assets/v2.0.0.h2drumkit"))
		if err != nil {
			t.Fatalf("ParseDrumkitArchive() returned error: %v", err)
		}

		if artifact.Name != "testKit" || artifact.Author != "theGreatWhiteShark" {
			t.Fatalf("unexpected common metadata: %+v", artifact)
		}

		if artifact.FormatVersion != 2 || artifact.Version != 0 || artifact.Instruments != 3 || artifact.Components != 3 || artifact.Samples != 3 {
			t.Fatalf("unexpected numeric metadata: %+v", artifact)
		}

		if got, want := artifact.Tags, []string{"Drumkit", "Example"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected tags: got %#v want %#v", got, want)
		}

		if got, want := artifact.InstrumentTypes, []string{"Hi-Hat", "Kick", "Snare"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected instrument types: got %#v want %#v", got, want)
		}
	})

	t.Run("legacy xml", func(t *testing.T) {
		t.Parallel()

		path := "/home/phil/git/hydrogen-index/res/hydrogen-artifacts/legacy-drumkits/kit-1.2.3/drumkit.xml"
		artifact, err := ParseDrumkitXMLFile(path, buildParseOptions(t, path, "https://example.com/assets/kit-1.2.3/drumkit.xml"))
		if err != nil {
			t.Fatalf("ParseDrumkitXMLFile() returned error: %v", err)
		}

		if artifact.Name != "Legacy-1.2.3" || artifact.Author != "Hydrogen Dev" {
			t.Fatalf("unexpected common metadata: %+v", artifact)
		}

		if artifact.FormatVersion != 1 || artifact.Version != 0 || artifact.Instruments != 1 || artifact.Components != 2 || artifact.Samples != 2 {
			t.Fatalf("unexpected numeric metadata: %+v", artifact)
		}

		if got, want := artifact.InstrumentTypes, []string{"Kick"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("unexpected instrument types: got %#v want %#v", got, want)
		}
	})
}

func buildParseOptions(t *testing.T, path string, url string) ParseOptions {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) failed: %v", path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("os.Stat(%q) failed: %v", path, err)
	}

	sum := sha256.Sum256(data)
	return ParseOptions{
		URL:  url,
		Size: info.Size(),
		Hash: hex.EncodeToString(sum[:]),
	}
}
