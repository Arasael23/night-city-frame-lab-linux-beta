//go:build linux && linuxbeta

package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"framelab/internal/analysis"
	"framelab/internal/hrec"
)

// LinuxBeta: the cut-down Linux build (-tags linuxbeta). The game runs under Proton or Wine;
// HitchRecorder records inside it, everything Windows-only is left out (see platform/linux_beta.go).
const LinuxBeta = true

func osName() string { return "linux" }

// feltFromHitch writes the stutter-key presses HitchRecorder noted (its MarkKey markers) to felt.txt
// as Unix ms, the same file the Windows editions write themselves.
func feltFromHitch(session string) int {
	f, err := hrec.Read(filepath.Join(session, analysis.FileHitch))
	if err != nil {
		return 0
	}
	u0 := f.Header.UnixMs0()
	var b strings.Builder
	n := 0
	for i := range f.Recs {
		r := &f.Recs[i]
		if r.Kind == hrec.KMark && r.Extra == hrec.MarkKey {
			fmt.Fprintf(&b, "%.0f\n", u0+float64(r.TUs)/1000)
			n++
		}
	}
	if n > 0 {
		os.WriteFile(filepath.Join(session, analysis.FileFelt), []byte(b.String()), 0o644)
	}
	return n
}
