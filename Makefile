BIN_DIR ?= $(shell test -w /opt/homebrew/bin && echo /opt/homebrew/bin || echo $(HOME)/.local/bin)

.PHONY: build test install build-bar install-bar release clean

build:
	env -u GOROOT go build -o build/runningback .

test:
	env -u GOROOT go test ./...

install: build
	mkdir -p $(BIN_DIR)
	cp build/runningback $(BIN_DIR)/runningback
	@echo "installed $(BIN_DIR)/runningback"

build-bar: build
	cd menubar && swift build -c release
	rm -rf build/RunningBackBar.app
	mkdir -p build/RunningBackBar.app/Contents/MacOS build/RunningBackBar.app/Contents/Resources
	cp menubar/.build/release/RunningBackBar build/RunningBackBar.app/Contents/MacOS/RunningBackBar
	cp build/runningback build/RunningBackBar.app/Contents/Resources/runningback
	printf '%s\n' \
		'<?xml version="1.0" encoding="UTF-8"?>' \
		'<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' \
		'<plist version="1.0">' \
		'<dict>' \
		'  <key>CFBundleName</key><string>RunningBackBar</string>' \
		'  <key>CFBundleDisplayName</key><string>RunningBack</string>' \
		'  <key>CFBundleIdentifier</key><string>dev.jeksn.runningback.bar</string>' \
		'  <key>CFBundleExecutable</key><string>RunningBackBar</string>' \
		'  <key>CFBundlePackageType</key><string>APPL</string>' \
		'  <key>LSMinimumSystemVersion</key><string>13.0</string>' \
		'  <key>LSUIElement</key><true/>' \
		'  <key>NSHighResolutionCapable</key><true/>' \
		'</dict>' \
		'</plist>' > build/RunningBackBar.app/Contents/Info.plist

install-bar: install build-bar
	cp -R build/RunningBackBar.app /Applications/ 2>/dev/null || { echo "copy to /Applications failed, app is at build/RunningBackBar.app"; exit 1; }
	@echo "installed /Applications/RunningBackBar.app"

# Build distributable artifacts for Apple Silicon macOS:
#   build/runningback-darwin-arm64.tar.gz  (standalone CLI)
#   build/RunningBackBar.zip                (standalone menu bar app)
release: test build-bar
	cp build/runningback build/runningback-darwin-arm64
	tar -czf build/runningback-darwin-arm64.tar.gz -C build runningback-darwin-arm64
	rm -f build/RunningBackBar.zip
	ditto -c -k --keepParent build/RunningBackBar.app build/RunningBackBar.zip
	@echo "artifacts in build/: runningback-darwin-arm64.tar.gz, RunningBackBar.zip"

clean:
	rm -rf build menubar/.build
