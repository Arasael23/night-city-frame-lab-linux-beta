//go:build !(linux && linuxbeta)

package app

// LinuxBeta is false in the Windows editions and the Linux dev server (see os_linuxbeta.go).
const LinuxBeta = false

func osName() string { return "" }

func feltFromHitch(session string) int { return 0 }
