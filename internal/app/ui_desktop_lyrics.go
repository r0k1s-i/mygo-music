package app

import (
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"mygo-music/internal/lyrics"
	"mygo-music/internal/music"
)

// desktopLyrics shows the current lyric line in a floating, transparent
// overlay near the bottom center of the screen, above other windows.
type desktopLyrics struct {
	a   *app
	win *mygo.Window

	winW, winH float32
	screenX    float32
	screenY    float32
}

func newDesktopLyrics(a *app) *desktopLyrics {
	dl := &desktopLyrics{a: a, winW: 640, winH: 64}
	dl.place()
	dl.win = mygo.NewWindow(mygo.WindowOptions{
		Title:         Tf("%s Desktop Lyrics", mygo.App.Name()),
		Panel:         true,
		Hidden:        true,
		Frameless:     true,
		Transparent:   true,
		DisableShadow: true,
		DisableResize: true,
		X:             int(dl.screenX),
		Y:             int(dl.screenY),
		Width:         int(dl.winW),
		Height:        int(dl.winH),
		Content:       ui.View(dl.view),
	})
	dl.win.OnClose(func(e *mygo.CloseEvent) {
		if !a.quitting {
			e.PreventDefault()
		}
	})
	dl.win.SetIgnoreMouseEvents(true)
	mygo.Screen.OnDisplaysChanged(func() {
		dl.place()
		if dl.win != nil {
			dl.win.SetPosition(int(dl.screenX), int(dl.screenY))
		}
	})
	dl.sync()
	return dl
}

// place positions the desktop lyrics overlay near the bottom center of the
// primary display.
func (dl *desktopLyrics) place() {
	d := mygo.Screen.PrimaryDisplay()
	area := d.WorkArea
	if area.Width == 0 || area.Height == 0 {
		area = d.Bounds
	}
	dl.screenX = float32(area.X) + (float32(area.Width)-dl.winW)/2
	dl.screenY = float32(area.Y) + float32(area.Height) - dl.winH - 48
}

// shouldShow reports whether the desktop lyrics window should currently be
// visible.
func (dl *desktopLyrics) shouldShow() bool {
	if !dl.a.settings.DesktopLyrics {
		return false
	}
	pl := dl.a.pl
	if pl == nil || !pl.Playing() || pl.Current() == nil {
		return false
	}
	lines, loaded := pl.Lyrics()
	return loaded && len(lines) > 0
}

// sync shows, hides, or repaints the desktop lyrics overlay as playback and
// settings change.
func (dl *desktopLyrics) sync() {
	if dl.win == nil {
		return
	}
	if dl.shouldShow() {
		if !dl.win.IsVisible() {
			dl.win.ShowInactive()
		}
		dl.win.Invalidate()
	} else if dl.win.IsVisible() {
		dl.win.Hide()
	}
}

// desktopLyricLines picks the text for the primary and secondary lines of the
// desktop lyrics overlay at the current position.
func desktopLyricLines(lines []lyrics.Line, pos time.Duration, tr *music.Track) (sung, sub string) {
	if len(lines) == 0 {
		return "", ""
	}
	cur := lyrics.Index(lines, pos)
	if cur >= 0 {
		sung = lines[cur].Text
		if lines[cur].Translation != "" {
			sub = lines[cur].Translation
		} else if cur+1 < len(lines) {
			sub = lines[cur+1].Text
		}
		return sung, sub
	}
	if tr != nil && tr.Name != "" {
		sung = tr.Name
		var names []string
		for _, ar := range tr.Artists {
			if ar.Name != "" {
				names = append(names, ar.Name)
			}
		}
		sub = strings.Join(names, ", ")
		return sung, sub
	}
	return "", lines[0].Text
}

// view draws the current lyric line and its translation or the following
// line inside a translucent capsule.
func (dl *desktopLyrics) view(c *ui.Context) {
	c.Root().Background(ui.Transparent)
	pl := dl.a.pl
	if pl == nil {
		return
	}
	lines, _ := pl.Lyrics()
	if len(lines) == 0 {
		return
	}
	pos := pl.Position()
	cur := lyrics.Index(lines, pos)

	// Redraw as the next line begins or the current one ends.
	if pl.Playing() {
		if cur < 0 && len(lines) > 0 && lines[0].At > pos {
			c.After(max(20*time.Millisecond, lines[0].At-pos+20*time.Millisecond))
		} else if cur+1 < len(lines) {
			c.After(max(20*time.Millisecond, lines[cur+1].At-pos+20*time.Millisecond))
		} else if cur >= 0 && lines[cur].End > pos {
			c.After(max(20*time.Millisecond, lines[cur].End-pos+20*time.Millisecond))
		}
	}

	sung, sub := desktopLyricLines(lines, pos, pl.Current())
	if sung == "" && sub == "" {
		return
	}

	ui.Column(c).Grow(1).Center().Children(func() {
		ui.Column(c).AlignItems(ui.Center).
			Background(ui.RGBA(0, 0, 0, 0.65)).
			Radius(16).
			Padding(6, 20).
			Children(func() {
				if sung != "" {
					ui.Text(c, sung).
						FontSize(20).
						FontWeight(700).
						TextColor(white).
						TextShadow(0, 1, 3, ui.RGBA(0, 0, 0, 0.8)).
						SingleLine()
				}
				if sub != "" {
					ui.Text(c, sub).
						FontSize(13).
						FontWeight(600).
						TextColor(white.Alpha(0.65)).
						TextShadow(0, 1, 2, ui.RGBA(0, 0, 0, 0.8)).
						Margin(2, 0, 0, 0).
						SingleLine()
				}
			})
	})
}
