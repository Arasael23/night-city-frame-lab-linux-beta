# Night City Frame Lab - Linux beta

A cut-down Linux version of [Night City Frame Lab](https://github.com/Arasael23/night-city-frame-lab), the stutter and performance recorder for modded Cyberpunk 2077. For Steam with Proton, Heroic and Lutris.

**This is a beta.** It hasn't been run on a real Linux game install yet, so feedback from players decides what gets fixed. If you try it, please say how it went on the [Nexus page](https://www.nexusmods.com/cyberpunk2077/mods/36756) (Posts or Bugs) or in this repository's Issues, with your distro, Steam / Heroic / Lutris, the Proton or Wine version, your graphics card, what the Record tab said, and the logs the README in the zip lists.

Download: the [releases](https://github.com/Arasael23/night-city-frame-lab-linux-beta/releases) here, or "Night City Frame Lab 2.5 Linux beta" under Files on [Nexus](https://www.nexusmods.com/cyberpunk2077/mods/36756) (the same zip).

## What it does

- Records a session from inside the game with HitchRecorder (a RED4ext plugin): the time of every frame, loading and world streaming, shader builds, the game's VRAM against its budget, and REDscript and CET time per mod.
- F10 in the game marks "I felt a stutter".
- The report with its numbers and charts (loading screens left out), Compare, and the Mods tab, which checks the build's mod files without starting the game.

## What it leaves out (Windows only)

PresentMon (GPU time per frame, frame pacing, input lag), PC counters and checks, Windows events, the VRAM page, live numbers while the game runs, game file names in reports, and the changes the Windows 2.5E edition can make.

## What it needs

- RED4ext loading under Proton or Wine. On Steam, the game's launch options need `WINEDLLOVERRIDES="winmm,version=n,b" %command%` (if you already play with RED4ext or CET mods, you have this).
- Cyber Engine Tweaks for the game markers. Without it, loading screens count as play.
- A 64-bit (x86-64) Linux and a web browser.

Unpack the zip, run `./NightCityFrameLab` in a terminal and Frame Lab opens in your browser. The README in the zip has the steps.

## Source

This is Frame Lab 2.5.2's source with the Linux side added. The Linux app is built with `-tags linuxbeta`; the Linux parts are `framelab/internal/platform/linux_beta.go`, `framelab/internal/app/os_linuxbeta.go` and `framelab/cmd/framelab/main_linux.go`. The Windows editions build from the same tree.

| Folder | What it is |
|---|---|
| `framelab/` | The app (Go, no cgo). On Linux the UI (HTML/JS in `ui/`) opens in your browser. `vendor/` holds the exact third-party Go code it is built with. |
| `hitchrecorder/` | `HitchRecorder.dll`, the RED4ext plugin that records inside the game (C, with MinHook in `third_party/`). The same file as in the Windows download. |
| `markers/` | The small mods Frame Lab puts into the game for a session: `cet/` (loading screens, menus, where V is) and `FrameLab.Markers*.reds`. |
| `release/` | The READMEs (`README-Linux.txt` is the one in the Linux zip), licence files and `optimise/` (what Frame Lab knows from downloaded mods: Nexus ids, names, file paths; no file contents). |
| `build.sh` | Builds everything; `./build.sh package` also makes `out/Night City Frame Lab 2.5.2 Linux beta.zip`. |

See [BUILD.md](BUILD.md) for the toolchain.

## Licence

Same as Frame Lab: all rights reserved, the source is here to read, not to reuse or re-upload (see [LICENSE](LICENSE)). Third-party parts keep their own licences, see [THIRD_PARTY.md](THIRD_PARTY.md) and `release/THIRD_PARTY-Linux.txt`.
