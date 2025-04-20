package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/getlantern/systray"

	_ "embed"
)

//go:embed icon.png
var iconData []byte
var makefile string

func main() {
	flag.StringVar(&makefile, "makefile", "~/Library/Mobile Documents/com~apple~CloudDocs/Makefile", "путь к Makefile, из которого читаются цели")
	flag.Parse()
	makefile = expandPath(makefile)

	if _, err := os.Stat(makefile); err != nil {
		log.Fatalf("Makefile не найден: %v", err)
	}

	systray.Run(onReady, nil)
}

// expandPath разворачивает '~', удаляет кавычки и убирает экранирование
func expandPath(p string) string {
	// удаляем внешние кавычки, если пользователь передал путь в "…" или '…'
	p = strings.Trim(p, "\"'")

	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, p[2:])
		}
	}
	// filepath.Abs приведёт к абсолютному пути и нормализует пробелы/.. и т.‑д.
	if abs, err := filepath.Abs(p); err == nil {
		p = abs
	}
	return p
}

func replaceSpaces(s string) string {
	// заменяем пробелы на %20, чтобы передать в osascript
	s = strings.ReplaceAll(s, " ", "\\\\ ")
	return s
}

func onReady() {
	systray.SetTitle("")
	systray.SetTooltip("Меню целей Make")
	systray.SetIcon(iconData)

	targets, err := parseMakefile(makefile)
	if err != nil {
		log.Printf("Ошибка парсинга Makefile: %v", err)
	}

	// создаём пункт меню на каждую цель и вешаем обработчик клика
	for _, t := range targets {
		item := systray.AddMenuItem(t, fmt.Sprintf("make %s", t))

		// обработчик в отдельной горутине, чтобы не блокировать главный поток macOS
		go func(target string, mi *systray.MenuItem) {
			for range mi.ClickedCh {
				go runTarget(target)
			}
		}(t, item)
	}

	systray.AddSeparator()
	quit := systray.AddMenuItem("Выход", "Закрыть приложение")
	go func() {
		<-quit.ClickedCh
		systray.Quit()
	}()
}

// runTarget открывает Terminal и запускает make <target> в каталоге Makefile
func runTarget(target string) {
	abs, _ := filepath.Abs(makefile)
	dir := replaceSpaces(filepath.Dir(abs))
	script := fmt.Sprintf(`tell application "Terminal"
        activate
        do script "cd %s && make -f %s %s"
    end tell`, dir, replaceSpaces(abs), target)

	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		log.Printf("Не удалось запустить цель %s: %v", target, err)
	}
}

// parseMakefile вытаскивает простые цели вида "target:" без зависимостей
func parseMakefile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	re := regexp.MustCompile(`^([A-Za-z0-9_\-\.]+):`) // имя до ':'
	targets := make(map[string]struct{})

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if len(line) == 0 || line[0] == '\t' || line[0] == '#' {
			continue
		}
		if m := re.FindStringSubmatch(line); m != nil && m[1] != ".PHONY" {
			targets[m[1]] = struct{}{}
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	list := make([]string, 0, len(targets))
	for t := range targets {
		list = append(list, t)
	}
	sort.Strings(list)
	return list, nil
}
