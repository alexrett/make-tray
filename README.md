# **MakeTray**

**MakeTray** is a tiny macOS menu‑bar app that lets you launch any Makefile target with a single click—no terminal juggling, no remembering command names.

## **Key Features**



| **What you get**      | **Details**                                                  |
| --------------------- | ------------------------------------------------------------ |
| **One‑click targets** | Every goal in your Makefile appears as its own menu item in the system bar. |
| **Live re‑parsing**   | Edit the Makefile in iCloud (or any path you pass at launch) and the list refreshes instantly. |
| **Terminal output**   | Each target opens a new Terminal window so you can watch logs in real time—then auto‑closes if you prefer. |
| **Zero clutter**      | Runs as an LSUIElement, so there’s no Dock icon or task‑switcher entry—just the menu‑bar glyph. |



## **Installation**

1. **Clone the repo**

```
git clone https://github.com/alexrett/make-tray.git
cd make-tray
```

2. **Install Go modules**

```
go mod tidy
```

3. **Build the app bundle**

```
make build      # => MakeTray.app inside /Application
```


## **Quick Start**

1. **Create a Makefile in your iCloud root**

```
make createMake   # helper target – drops a skeleton Makefile in iCloud
```

1. **Launch MakeTray**

   *Double‑click* MakeTray.app (or run open /Applications/MakeTray.app).

   The 🛠️ icon appears near the clock; click it to see your targets.

2. **Edit your Makefile**

   Add or tweak targets—MakeTray picks them up automatically.



## **Advanced Usage**



| **Scenario**                   | **Command**                                        |
| ------------------------------ | -------------------------------------------------- |
| Use a custom Makefile location | MakeTray -makefile "$HOME/Projects/foo/Makefile"   |
| Add an app‑icon of your own    | Replace icon.png and rebuild.                      |
| Keep the Terminal window open  | Comment out the osascript “close” line in main.go. |



## **Contributing**

Pull requests are welcome!

1. Fork → feature branch → PR.
2. Follow go vet and golangci-lint checks.
3. Add yourself to CONTRIBUTORS.md.



## **License**

Released under the [MIT License](LICENSE).