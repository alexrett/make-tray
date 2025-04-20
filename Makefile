build:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o MakeTray.app/Contents/MacOS/MakeTray main.go
	chmod +x MakeTray.app/Contents/MacOS/MakeTray
	xattr -dr com.apple.quarantine MakeTray.app
	cp -r MakeTray.app /Applications/

createMake:
	touch ~/Library/Mobile Documents/com~apple~CloudDocs/Makefile