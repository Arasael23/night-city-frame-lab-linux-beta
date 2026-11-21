//go:build linux && linuxbeta

// Linux beta: the game runs under Proton or Wine, so only what can be read from files and /proc is
// here. HitchRecorder records inside the game as on Windows; PresentMon, the Windows event tracing,
// the PC counters and the Windows checks and changes are Windows only and report so.
package platform

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var errWindowsOnly = errors.New("Windows only (not in the Linux beta)")

var start = time.Now()

func QPC() int64     { return int64(time.Since(start).Nanoseconds() / 100) }
func QPCFreq() int64 { return 10_000_000 }

// IsAdmin: nothing in the Linux beta needs root.
func IsAdmin() bool { return true }

// SingleInstance: the Linux main takes a lock file itself.
func SingleInstance() bool          { return true }
func MessageBox(title, text string) { os.Stderr.WriteString(title + ": " + text + "\n") }
func HideWindow(cmd *exec.Cmd)      {}

func xdgOpen(target string) {
	cmd := exec.Command("xdg-open", target)
	if cmd.Start() == nil {
		go cmd.Wait()
	}
}

func OpenFolder(path string) { xdgOpen(path) }
func OpenURL(url string)     { xdgOpen(url) }

// winePath turns a Wine path into a Linux one: Z:\home\me\... -> /home/me/... Other drive letters
// can't be mapped from here and give "".
func winePath(p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	if len(p) > 2 && (p[0] == 'Z' || p[0] == 'z') && p[1] == ':' {
		return strings.ReplaceAll(p[2:], `\`, "/")
	}
	return ""
}

func baseAny(p string) string {
	if i := strings.LastIndexAny(p, `\/`); i >= 0 {
		return p[i+1:]
	}
	return p
}

// procsWith lists the processes that have an argument naming exe (Wine keeps the Windows path in
// the command line; Proton's own wrapper has the Linux path).
func procsWith(exe string) []GameProc {
	ents, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}
	self := os.Getpid()
	var out []GameProc
	for _, e := range ents {
		pid, err := strconv.Atoi(e.Name())
		if err != nil || pid == self {
			continue
		}
		b, err := os.ReadFile(filepath.Join("/proc", e.Name(), "cmdline"))
		if err != nil || len(b) == 0 {
			continue
		}
		for _, arg := range strings.Split(strings.TrimRight(string(b), "\x00"), "\x00") {
			if strings.EqualFold(baseAny(arg), exe) {
				out = append(out, GameProc{Pid: uint32(pid), Path: winePath(arg)})
				break
			}
		}
	}
	return out
}

func ProcessPaths(exe string) []string {
	var out []string
	for _, p := range procsWith(exe) {
		if p.Path != "" {
			out = append(out, p.Path)
		}
	}
	return out
}

// FindGame: the running Cyberpunk2077.exe (under Proton or Wine). When the Windows path can't be
// mapped, the path stays empty and the game counts as the armed build's.
func FindGame() []GameProc { return procsWith("Cyberpunk2077.exe") }

// ReadHitchState: HitchRecorder's live block lives in Wine's memory, out of reach from Linux.
func ReadHitchState() *HitchState { return &HitchState{} }

// ---- graphics card: name and vendor from /sys and the drivers' own files, no live readings

func readTrim(p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func GPU() GpuInfo {
	type card struct{ vendor, name string }
	var cards []card
	devs, _ := filepath.Glob("/sys/class/drm/card[0-9]*/device/vendor")
	seen := map[string]bool{}
	for _, v := range devs {
		dev := filepath.Dir(v)
		if real, err := filepath.EvalSymlinks(dev); err == nil {
			if seen[real] {
				continue
			}
			seen[real] = true
		}
		switch readTrim(v) {
		case "0x10de":
			cards = append(cards, card{"nvidia", nvidiaName()})
		case "0x1002":
			cards = append(cards, card{"amd", pciName(dev, "AMD Radeon graphics")})
		case "0x8086":
			cards = append(cards, card{"intel", pciName(dev, "Intel graphics")})
		}
	}
	if len(cards) == 0 {
		return GpuInfo{Name: "", Source: "Linux", Note: "no graphics card found in /sys/class/drm"}
	}
	// the card the game most likely runs on: NVIDIA or AMD before Intel
	sort.SliceStable(cards, func(i, j int) bool { return cards[i].vendor != "intel" && cards[j].vendor == "intel" })
	return GpuInfo{Name: cards[0].name, Vendor: cards[0].vendor, Source: "Linux (no live readings in the beta)"}
}

func nvidiaName() string {
	fs, _ := filepath.Glob("/proc/driver/nvidia/gpus/*/information")
	for _, f := range fs {
		for _, l := range strings.Split(readTrim(f), "\n") {
			if strings.HasPrefix(l, "Model:") {
				return strings.TrimSpace(strings.TrimPrefix(l, "Model:"))
			}
		}
	}
	return "NVIDIA graphics card"
}

var reLspci = regexp.MustCompile(`"[^"]*" "[^"]*" "([^"]*)"`)

// pciName asks lspci (if there) for the card's marketing name.
func pciName(dev, fallback string) string {
	slot := filepath.Base(readLink(dev))
	if slot == "" || slot == "." {
		return fallback
	}
	out, err := exec.Command("lspci", "-mm", "-s", slot).Output()
	if err != nil {
		return fallback
	}
	if m := reLspci.FindStringSubmatch(string(out)); m != nil && m[1] != "" {
		return m[1]
	}
	return fallback
}

func readLink(p string) string {
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		return ""
	}
	return r
}

func GPUName() string { return GPU().Name }

func StartSampler(gamePid uint32, out string) (Sampler, error) {
	return nil, errors.New("PC counters (CPU, RAM, graphics card) aren't in the Linux beta yet")
}

func LiveGPU() map[string]float64 { return map[string]float64{} }

func ExtractNames(wolvenKitCommon, oodleDll, outGz string) error { return errWindowsOnly }

func Beep(on bool) {}

func Oodle(oodleDll string) (func(src []byte, size int) ([]byte, error), error) {
	return nil, errWindowsOnly
}

// KeyHit: Linux keeps the keyboard to the window that has it. The stutter key is caught by
// HitchRecorder inside the game instead (its MarkKey).
func KeyHit(vk int) bool { return false }

func SystemEvents(since time.Time) []SysEvent { return nil }

// SystemVersions: the Linux distribution and kernel, and the NVIDIA driver when there is one.
func SystemVersions(gpuName string) (win, driver string) {
	name := ""
	if f, err := os.Open("/etc/os-release"); err == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if v, ok := strings.CutPrefix(sc.Text(), "PRETTY_NAME="); ok {
				name = strings.Trim(v, `"`)
			}
		}
		f.Close()
	}
	if k := readTrim("/proc/sys/kernel/osrelease"); k != "" {
		name = strings.TrimSpace(name + " (Linux " + k + ")")
	}
	if v := readTrim("/proc/driver/nvidia/version"); v != "" {
		if m := regexp.MustCompile(`Kernel Module\s+(?:for \S+\s+)?([0-9.]+)`).FindStringSubmatch(v); m != nil {
			driver = m[1]
		}
	}
	return name, driver
}

type CPUTemp struct{ Source string }

func OpenCPUTemp(modDir string) *CPUTemp { return &CPUTemp{} }
func (c *CPUTemp) Read() float64         { return 0 }
func (c *CPUTemp) Close()                {}

// ---- checks: the PC checks are Windows ones; the Linux beta hides that page

func ReadCheckFacts() CheckFacts { return CheckFacts{BootTimer: map[string]string{}} }
func PCIeNow() PCIeFact          { return PCIeFact{} }
func FileVersion(path string) (string, string) {
	return "", ""
}
func ExeSignature(path string) string { return "ok" }
func DriveInfo(path string) DriveFact { return DriveFact{} }

// UserSettingsPath: the game's UserSettings.json inside the Proton / Wine prefix (the newest one found).
func UserSettingsPath() string {
	const rel = "AppData/Local/CD Projekt Red/Cyberpunk 2077/UserSettings.json"
	var pfx []string
	for _, lib := range steamLibraries() {
		pfx = append(pfx, filepath.Join(lib, "steamapps", "compatdata", "1091500", "pfx"))
	}
	home, _ := os.UserHomeDir()
	for _, g := range []string{"Games/Heroic/Prefixes/*", "Games/Heroic/Prefixes/*/*", "Games/*"} {
		m, _ := filepath.Glob(filepath.Join(home, g))
		pfx = append(pfx, m...)
	}
	best, bestT := "", time.Time{}
	for _, p := range pfx {
		m, _ := filepath.Glob(filepath.Join(p, "drive_c", "users", "*", rel))
		for _, f := range m {
			if fi, err := os.Stat(f); err == nil && fi.ModTime().After(bestT) {
				best, bestT = f, fi.ModTime()
			}
		}
	}
	return best
}

func FileReadOnly(path string) (bool, bool) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, false
	}
	return fi.Mode().Perm()&0o200 == 0, true
}

func FTHEntries() []string                         { return nil }
func ClearFTH() (map[string][]byte, error)         { return nil, errWindowsOnly }
func RestoreFTH(saved map[string][]byte) error     { return nil }
func PawnIOVersion() string                        { return "" }
func ClearBootTimer() (map[string]string, error)   { return nil, errWindowsOnly }
func RestoreBootTimer(was map[string]string) error { return nil }
func ClearReservedCpuSets() ([]byte, error)        { return nil, errWindowsOnly }
func RestoreReservedCpuSets(b []byte) error        { return nil }

// ---- processor

func CPU() CPUInfo {
	name := ""
	cores := map[string]bool{}
	logical := 0
	if f, err := os.Open("/proc/cpuinfo"); err == nil {
		sc := bufio.NewScanner(f)
		phys, core := "", ""
		for sc.Scan() {
			k, v, ok := strings.Cut(sc.Text(), ":")
			if !ok {
				continue
			}
			k, v = strings.TrimSpace(k), strings.TrimSpace(v)
			switch k {
			case "processor":
				logical++
			case "model name":
				if name == "" {
					name = v
				}
			case "physical id":
				phys = v
			case "core id":
				core = v
				cores[phys+"/"+core] = true
			}
		}
		f.Close()
	}
	n := len(cores)
	if n == 0 {
		n = logical
	}
	info := CPUInfo{Name: cpuName(name), Cores: n, PCores: n}
	// Intel hybrid processors: the kernel lists P-core and E-core threads separately
	if p, e := cpuList("/sys/devices/cpu_core/cpus"), cpuList("/sys/devices/cpu_atom/cpus"); p > 0 && e > 0 {
		info.Hybrid, info.ECores = true, e
		info.PCores = n - e
		if info.PCores <= 0 {
			info.PCores = p / 2
		}
	}
	return info
}

// cpuList counts the CPUs in a kernel list like "0-15,20".
func cpuList(path string) int {
	s := readTrim(path)
	if s == "" {
		return 0
	}
	n := 0
	for _, part := range strings.Split(s, ",") {
		a, b, ok := strings.Cut(part, "-")
		x, err1 := strconv.Atoi(a)
		if !ok {
			if err1 == nil {
				n++
			}
			continue
		}
		y, err2 := strconv.Atoi(b)
		if err1 == nil && err2 == nil && y >= x {
			n += y - x + 1
		}
	}
	return n
}

func PinGame() int { return 0 }
func UnpinGame()   {}

// ---- where the game is installed: Steam (Proton), Heroic and Lutris

func steamLibraries() []string {
	home, _ := os.UserHomeDir()
	roots := []string{".local/share/Steam", ".steam/steam", ".steam/root", ".var/app/com.valvesoftware.Steam/.local/share/Steam", "snap/steam/common/.local/share/Steam"}
	reP := regexp.MustCompile(`"path"\s+"([^"]+)"`)
	seen := map[string]bool{}
	var out []string
	add := func(p string) {
		p = filepath.Clean(p)
		k := p
		if r, err := filepath.EvalSymlinks(p); err == nil {
			k = r
		}
		if !seen[k] {
			seen[k] = true
			out = append(out, p)
		}
	}
	for _, r := range roots {
		root := filepath.Join(home, r)
		if _, err := os.Stat(filepath.Join(root, "steamapps")); err != nil {
			continue
		}
		add(root)
		if b, err := os.ReadFile(filepath.Join(root, "steamapps", "libraryfolders.vdf")); err == nil {
			for _, m := range reP.FindAllStringSubmatch(string(b), -1) {
				add(strings.ReplaceAll(m[1], `\\`, `\`))
			}
		}
	}
	return out
}

func GameInstallHints() []InstallHint {
	var out []InstallHint
	for _, lib := range steamLibraries() {
		out = append(out, InstallHint{"Steam", filepath.Join(lib, "steamapps", "common", "Cyberpunk 2077")})
	}
	home, _ := os.UserHomeDir()
	// Heroic (GOG and Epic): its lists of installed games, native and Flatpak
	reI := regexp.MustCompile(`"install_path"\s*:\s*"([^"]+)"`)
	for _, f := range []string{
		".config/heroic/gog_store/installed.json", ".config/heroic/legendaryConfig/legendary/installed.json",
		".var/app/com.heroicgameslauncher.hgl/config/heroic/gog_store/installed.json",
		".var/app/com.heroicgameslauncher.hgl/config/heroic/legendaryConfig/legendary/installed.json"} {
		b, err := os.ReadFile(filepath.Join(home, f))
		if err != nil {
			continue
		}
		store := "GOG"
		if strings.Contains(f, "legendary") {
			store = "Epic"
		}
		for _, m := range reI.FindAllStringSubmatch(string(b), -1) {
			if p := strings.ReplaceAll(m[1], `\/`, `/`); strings.Contains(strings.ToLower(p), "cyberpunk") {
				out = append(out, InstallHint{store, p})
			}
		}
	}
	// the usual folders: Heroic's default, Lutris prefixes, a plain copy in ~/Games
	for _, g := range []string{
		"Games/Heroic/Cyberpunk*", "Games/Cyberpunk*", "Games/*/Cyberpunk*",
		"Games/*/drive_c/GOG Games/Cyberpunk*", "Games/*/drive_c/Program Files*/GOG Galaxy/Games/Cyberpunk*",
		"Games/*/drive_c/Program Files*/Epic Games/Cyberpunk*", "Games/*/drive_c/Program Files*/Steam/steamapps/common/Cyberpunk*"} {
		m, _ := filepath.Glob(filepath.Join(home, g))
		for _, p := range m {
			out = append(out, InstallHint{"", p})
		}
	}
	return out
}

func Drives() []string { return nil }

// ---- experiments and Windows settings: none on Linux

func ExpRead(id, game string) (on bool, detail string, ok bool) {
	return false, "Windows only", false
}
func ExpOn(id, game string) (map[string]string, error)         { return nil, errWindowsOnly }
func ExpRestore(id, game string, prev map[string]string) error { return nil }
func PurgeStandby() error                                      { return errWindowsOnly }
func ExpRun(id, game string) (string, error)                   { return "", errWindowsOnly }
func TimerResolution() float64                                 { return 0 }
func HoldTimer(on bool)                                        {}
func PinGameX3D() int                                          { return 0 }
func QuietStart() []string                                     { return nil }
func QuietEnd(names []string)                                  {}
func OpenCoreIsolation() error                                 { return errWindowsOnly }

func GpuSnapshot() ([]GpuAdapter, []GpuProc, error) { return nil, nil, errWindowsOnly }
func GpuPrefs() (map[string]string, error)          { return map[string]string{}, nil }
func SetGpuPref(path, data string) error            { return errWindowsOnly }
func DisplayDrivers() []string                      { return nil }
func ProgramExists(path string) bool                { _, err := os.Stat(path); return err == nil }
func NvProfile() ([]NvSetting, error)               { return nil, errWindowsOnly }

var ErrPickCancelled = errors.New("cancelled")

// PickFolder: zenity (GNOME, Cinnamon, ...) or kdialog (KDE). Without either, the path is typed in.
func PickFolder(title, start string) (string, error) {
	if start == "" {
		start, _ = os.UserHomeDir()
	}
	var cmd *exec.Cmd
	if p, err := exec.LookPath("zenity"); err == nil {
		cmd = exec.Command(p, "--file-selection", "--directory", "--title="+title, "--filename="+strings.TrimRight(start, "/")+"/")
	} else if p, err := exec.LookPath("kdialog"); err == nil {
		cmd = exec.Command(p, "--getexistingdirectory", start, "--title", title)
	} else {
		return "", errors.New("no folder picker found (zenity or kdialog): type or paste the folder path into the box instead")
	}
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", ErrPickCancelled
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
