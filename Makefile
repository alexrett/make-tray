build:
	go mod tidy
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o MakeTray.app/Contents/MacOS/MakeTray main.go
	chmod +x MakeTray.app/Contents/MacOS/MakeTray
	xattr -dr com.apple.quarantine MakeTray.app
	cp -r MakeTray.app /Applications/

run:
	go mod tidy
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o MakeTray.app/Contents/MacOS/MakeTray main.go
	./MakeTray.app/Contents/MacOS/MakeTray

createMake:
	touch ~/Library/Mobile Documents/com~apple~CloudDocs/Makefile