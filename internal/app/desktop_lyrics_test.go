package app

import (
	"testing"
	"time"

	"github.com/egoist/mygo/audio"

	"mygo-music/internal/lyrics"
	"mygo-music/internal/music"
)

// Desktop lyrics show only while a track with lyrics is actively playing and
// the setting is turned on.
func TestDesktopLyricsVisibility(t *testing.T) {
	a := newTestApp()
	dl := &desktopLyrics{a: a, winW: 640, winH: 64}
	a.desktopLyrics = dl

	if dl.shouldShow() {
		t.Fatal("should not show when disabled")
	}

	a.settings.DesktopLyrics = true
	if dl.shouldShow() {
		t.Fatal("should not show when nothing is playing")
	}

	a.pl.Cue([]music.Track{{ID: "1", Name: "Song Without Lyrics", Duration: 200000}})
	out := audio.NewPlayer(&silence{})
	out.SetVolume(0)
	a.pl.Attach(out)
	defer out.Close()

	if dl.shouldShow() {
		t.Fatal("should not show when track has no lyrics")
	}

	a.pl.SetLyrics(lyrics.Build("[00:00.00]First line\n[00:10.00]Second line", ""))
	if !dl.shouldShow() {
		t.Fatal("should show when playing track with lyrics")
	}

	a.pl.TogglePlay()
	if dl.shouldShow() {
		t.Fatal("should hide when paused")
	}

	a.pl.TogglePlay()
	if !dl.shouldShow() {
		t.Fatal("should show again when resumed")
	}

	a.settings.DesktopLyrics = false
	if dl.shouldShow() {
		t.Fatal("should hide when setting is toggled off")
	}
}

// Desktop lyrics pick the current line with translation or the following line,
// or the track name and artists before playback reaches the first line.
func TestDesktopLyricLines(t *testing.T) {
	lines := []lyrics.Line{
		{At: 5 * time.Second, Text: "First line", Translation: "第一行"},
		{At: 10 * time.Second, Text: "Second line"},
		{At: 15 * time.Second, Text: "Third line"},
	}
	tr := &music.Track{
		Name:    "Test Song",
		Artists: []music.Artist{{Name: "Artist A"}, {Name: "Artist B"}},
	}

	// Before the first line starts: show track name and artists.
	sung, sub := desktopLyricLines(lines, 2*time.Second, tr)
	if sung != "Test Song" || sub != "Artist A, Artist B" {
		t.Fatalf("before first line: got %q, %q, want %q, %q", sung, sub, "Test Song", "Artist A, Artist B")
	}

	// Translation is preferred over the next line.
	sung, sub = desktopLyricLines(lines, 6*time.Second, tr)
	if sung != "First line" || sub != "第一行" {
		t.Fatalf("with translation: got %q, %q, want %q, %q", sung, sub, "First line", "第一行")
	}

	// Next line is shown when the current line has no translation.
	sung, sub = desktopLyricLines(lines, 11*time.Second, tr)
	if sung != "Second line" || sub != "Third line" {
		t.Fatalf("without translation: got %q, %q, want %q, %q", sung, sub, "Second line", "Third line")
	}

	// Last line has no following line.
	sung, sub = desktopLyricLines(lines, 16*time.Second, tr)
	if sung != "Third line" || sub != "" {
		t.Fatalf("last line: got %q, %q, want %q, %q", sung, sub, "Third line", "")
	}
}
