//go:build linux && linuxbeta

// Night City Frame Lab, Linux beta: the same app in the browser. Run it from a terminal and leave the
// terminal open while you record; Ctrl+C (or closing the terminal) quits.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"framelab/internal/app"
	"framelab/internal/platform"
	"framelab/ui"
)

func main() {
	exe, _ := os.Executable()
	root := filepath.Dir(exe)
	// one Frame Lab per folder: two would both take HitchRecorder in and out of the game
	if lock, err := os.OpenFile(filepath.Join(root, ".framelab.lock"), os.O_CREATE|os.O_RDWR, 0o644); err == nil {
		if syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) != nil {
			fmt.Fprintln(os.Stderr, "Frame Lab is already open (in another terminal).")
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Frame Lab can't write to its own folder (%s): %v\nPut it in a folder of your own (for example ~/Night City Frame Lab) and start it again.\n", root, err)
		os.Exit(1)
	}
	srv := app.NewServer(app.NewPaths(root), ui.FS)
	url, err := srv.Start(os.Getenv("FRAMELAB_ADDR"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not start:", err)
		os.Exit(1)
	}
	fmt.Printf("Night City Frame Lab %s (v%s) is running.\n\n", app.Edition(), app.Version)
	fmt.Printf("If your browser didn't open it, open this address:\n  %s\n\n", url)
	fmt.Println("Leave this terminal open while you record. Ctrl+C here quits Frame Lab.")
	if os.Getenv("FRAMELAB_NO_BROWSER") == "" {
		platform.OpenURL(url)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	<-sig
	fmt.Println("\nClosing Frame Lab ...")
	srv.Ctl.Shutdown()
}
