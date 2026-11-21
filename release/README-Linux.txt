Night City Frame Lab 2.5.2 - Linux beta
=======================================

A cut-down Frame Lab for Cyberpunk 2077 on Linux: Steam with Proton, Heroic, Lutris.
It's a beta and hasn't been run on a real Linux game install yet, so what you report decides
what gets fixed.

What it does
- Records a game session from inside the game with HitchRecorder (a RED4ext plugin): the time of
  every frame, loading and world streaming, shader builds, the game's VRAM against its budget,
  and REDscript and CET time per mod in the parts you mark with F11.
- F10 in the game marks "I felt a stutter".
- The report with its numbers and charts (loading screens left out), Compare, and the Mods tab,
  which checks the build's mod files without starting the game.

What it doesn't do (these are Windows only)
- PresentMon: GPU time per frame, frame pacing, input lag.
- PC counters (CPU, RAM, graphics card clocks and temperature), Windows events, PC checks and the
  VRAM page.
- Live numbers while the game runs: everything shows in the report after you quit the game.
- File names in reports: game files show as numbers.

What it needs
- RED4ext loading under Proton or Wine. On Steam, the game's launch options need:
      WINEDLLOVERRIDES="winmm,version=n,b" %command%
  If you already play with RED4ext or CET mods, you have this.
- Cyber Engine Tweaks for the game markers. Without it, loading screens count as play.
- A 64-bit (x86-64) Linux and a web browser.

How to start
1. Unpack this folder somewhere you can write to, for example ~/Night City Frame Lab
2. In a terminal:
       cd ~/"Night City Frame Lab"
       ./NightCityFrameLab
   (If it says "Permission denied": chmod +x NightCityFrameLab)
3. Frame Lab opens in your browser. If it doesn't, open the address it prints in the terminal.
   Leave the terminal open while you use it. Ctrl+C in the terminal quits Frame Lab.
4. Settings > Find game installs. Or Add a build and type the game folder (the one with bin/x64
   in it).
5. Record > Start session. Start the game from Steam, Heroic or Lutris, play for about 10
   minutes, then quit the game. The report opens in Sessions.

For each session, Frame Lab puts HitchRecorder (red4ext/plugins/HitchRecorder) and its markers mod
(bin/x64/plugins/cyber_engine_tweaks/mods/NightCityFrameLab) into the game, and takes them out
again when you quit the game. HitchRecorder is the same file as in the Windows download.
If the game crashes during a session, the next start is clean anyway: HitchRecorder switches itself
off after one game run. Please post the crash with your setup (see below).

If something doesn't work
Post on the Nexus page (Posts or Bugs) with:
- your distro, Steam / Heroic / Lutris, the Proton or Wine version, and your graphics card
- what the Record tab's step list said
- the session's hitch.log if there is one (Sessions > Open folder) and the newest file in the
  game's red4ext/logs folder
