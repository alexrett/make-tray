package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"

	_ "embed"
)

//go:embed icon.png
var iconData []byte

// default config lives in ~/Library/Application Support/MakeTray/config.json
var (
	configPath string
	config     Config
	groupMenus []*systray.MenuItem
	manageMenu *systray.MenuItem // pointer to “Manage Makefiles ▸”
	exitMenu   *systray.MenuItem // pointer to Exit”
)

// Config format:
//
//	{
//	  "makefiles": [
//	    { "path": "~/proj/foo/Makefile", "label": "Foo" },
//	    { "path": "$HOME/bar/Makefile" }                // label optional
//	  ]
//	}
type Config struct {
	Makefiles []Entry `json:"makefiles"`
}

// Entry describes a single Makefile and optional menu label.
type Entry struct {
	Path  string `json:"path"`
	Label string `json:"label,omitempty"`
}

func main() {
	home, _ := os.UserHomeDir()
	defaultCfg := filepath.Join(home, "Library", "Application Support", "MakeTray", "config.json")

	flag.StringVar(
		&configPath,
		"config",
		defaultCfg,
		"path to config file (JSON); if missing, it will be created",
	)
	flag.Parse()

	if err := loadConfig(configPath); err != nil {
		log.Fatalf("config error: %v", err)
	}
	if len(config.Makefiles) == 0 {
		log.Fatalf("no Makefiles listed in %s", configPath)
	}

	systray.Run(onReady, nil)
}

// expandPath trims surrounding quotes, expands environment variables
// (e.g. $HOME) and ~, and returns an absolute, cleaned path.
func expandPath(p string) string {
	p = strings.Trim(p, "\"'")
	p = os.ExpandEnv(p)

	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, p[2:])
		}
	}
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	return p
}

// loadConfig reads JSON config; if the file does not exist it creates
// a stub file with a single example entry.
func loadConfig(path string) error {
	path = expandPath(path)

	// ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// create default stub
		cfg := Config{
			Makefiles: []Entry{
				{Path: "./Makefile", Label: "Local Makefile"},
			},
		}
		data, _ := json.MarshalIndent(cfg, "", "  ")
		if writeErr := os.WriteFile(path, data, 0o644); writeErr != nil {
			return writeErr
		}
		config = cfg
		log.Printf("created stub config at %s — edit it to add your Makefiles", path)
		return nil
	}

	if err := json.Unmarshal(b, &config); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// expand paths
	for i := range config.Makefiles {
		config.Makefiles[i].Path = expandPath(config.Makefiles[i].Path)
	}
	return nil
}

func onReady() {
	systray.SetTitle("")
	systray.SetTooltip("MakeTray")
	systray.SetIcon(iconData)

	for _, entry := range config.Makefiles {
		addMakefileGroup(entry)
	}

	buildManagementMenu()
	// start watcher for config + Makefiles
	var watchPaths []string
	watchPaths = append(watchPaths, expandPath(configPath))
	for _, e := range config.Makefiles {
		watchPaths = append(watchPaths, e.Path)
	}
	startWatch(watchPaths)
}

func buildManagementMenu() {
	// Management submenu
	systray.AddSeparator()
	manageMenu = systray.AddMenuItem("Settings", "")
	addNew := manageMenu.AddSubMenuItem("Add…", "Add a new Makefile via dialog")
	openCfg := manageMenu.AddSubMenuItem("Open config…", "Open JSON config in default editor")
	exitMenu = manageMenu.AddSubMenuItem("Exit", "Close the app")
	// click handlers
	go func() {
		for {
			select {
			case <-addNew.ClickedCh:
				go func() {
					path, _ := chooseFileGUI()
					if path == "" {
						return // user cancelled
					}
					label, _ := promptLabelGUI()
					config.Makefiles = append(config.Makefiles, Entry{Path: path, Label: label})
					if data, err := json.MarshalIndent(config, "", "  "); err == nil {
						_ = os.WriteFile(expandPath(configPath), data, 0o644)
					}
					_ = exec.Command("osascript", "-e",
						`display notification "Makefile added – menu refreshed." with title "MakeTray"`).Run()
				}()
			case <-openCfg.ClickedCh:
				go exec.Command("open", "-t", expandPath(configPath)).Run()
			}
		}
	}()

	go func() { <-exitMenu.ClickedCh; systray.Quit() }()
}

// addMakefileGroup creates a submenu for a Makefile and populates
// it with its targets.
func addMakefileGroup(entry Entry) {
	mf := entry.Path
	label := entry.Label
	if label == "" {
		// fallback: directory name containing the Makefile
		label = filepath.Base(filepath.Dir(mf))
	}

	group := systray.AddMenuItem(label, mf)
	groupMenus = append(groupMenus, group)

	targets, err := parseMakefileTargets(mf)
	if err != nil {
		log.Printf("%s: parse error: %v", mf, err)
		sub := group.AddSubMenuItem("(error)", "parse error")
		sub.Disable()
		return
	}
	if len(targets) == 0 {
		sub := group.AddSubMenuItem("(no targets)", "empty Makefile")
		sub.Disable()
		return
	}

	for _, t := range targets {
		item := group.AddSubMenuItem(t, fmt.Sprintf("make -f %s %s", label, t))
		go func(target string, mi *systray.MenuItem) {
			for range mi.ClickedCh {
				go runTarget(mf, target)
			}
		}(t, item)
	}
}

// refreshMenus hides old Makefile groups and rebuilds them from current config.
func refreshMenus() {
	// hide previous dynamic items
	if manageMenu != nil {
		manageMenu.Hide()
	}
	if exitMenu != nil {
		exitMenu.Hide()
	}
	for _, g := range groupMenus {
		g.Hide()
	}
	groupMenus = nil

	// reload config
	if err := loadConfig(configPath); err != nil {
		log.Printf("reload config error: %v", err)
		return
	}

	// rebuild menu: groups first, then management+exit
	for _, e := range config.Makefiles {
		addMakefileGroup(e)
	}
	buildManagementMenu()
	// restart watcher with updated list
	var watchPaths []string
	watchPaths = append(watchPaths, expandPath(configPath))
	for _, e := range config.Makefiles {
		watchPaths = append(watchPaths, e.Path)
	}
	startWatch(watchPaths)
}

// runTarget opens Terminal and runs the selected target using the
// specified Makefile.
func runTarget(makefile, target string) {
	abs, _ := filepath.Abs(makefile)
	dir := filepath.Dir(abs)

	script := fmt.Sprintf(`tell application "Terminal"
        activate
        do script "cd %s && make -f '%s' %s"
    end tell`, dir, abs, target)

	if err := exec.Command("osascript", "-e", script).Run(); err != nil {
		log.Printf("unable to run target %s (%s): %v", target, filepath.Base(makefile), err)
	}
}

// chooseFileGUI shows the standard macOS “choose file” dialog via AppleScript
// and returns the selected path as a string (empty if user cancelled).
func chooseFileGUI() (string, error) {
	out, err := exec.Command("osascript", "-e",
		`set f to (choose file with prompt "Select a Makefile")`,
		"-e", `POSIX path of f`).CombinedOutput()
	if err != nil {
		// user may have cancelled; treat that as no selection
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

// promptLabelGUI asks the user for a label using a simple dialog.
// Returns the label (may be empty) and error.
func promptLabelGUI() (string, error) {
	out, err := exec.Command("osascript", "-e",
		`text returned of (display dialog "Optional label for menu item:" default answer "")`).CombinedOutput()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

// triggerReload closes the current systray loop so main() can build a new one.
func triggerReload() {
	refreshMenus()
}

// startWatch sets up fsnotify watchers on the given file paths. When any of
// them is written/created, the app respawns itself to refresh menus.
func startWatch(paths []string) {
	go func() {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			log.Printf("watch error: %v", err)
			return
		}
		defer w.Close()

		for _, p := range paths {
			_ = w.Add(p)
		}

		for {
			select {
			case ev := <-w.Events:
				needStop := false
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
					log.Printf("change detected in %s — reloading", ev.Name)
					needStop = true
					go triggerReload()
					return
				}
				if needStop {
					w.Close()
					break // stop watching
				}
			case err := <-w.Errors:
				log.Printf("watch error: %v", err)
			}
		}
	}()
}

// parseMakefileTargets returns a sorted list of simple targets found
// in a Makefile.
func parseMakefileTargets(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	re := regexp.MustCompile(`^([A-Za-z0-9_\-\.]+):`)
	targets := make(map[string]struct{})

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 ||
			strings.HasPrefix(line, "\t") ||
			strings.HasPrefix(line, "#") {
			continue
		}
		if m := re.FindStringSubmatch(line); m != nil && m[1] != ".PHONY" {
			targets[m[1]] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	list := make([]string, 0, len(targets))
	for t := range targets {
		list = append(list, t)
	}
	sort.Strings(list)
	return list, nil
}
