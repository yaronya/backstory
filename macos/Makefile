SDK = /Library/Developer/CommandLineTools/SDKs/MacOSX26.2.sdk
TARGET = arm64-apple-macosx14.0
VFS_OVERLAY = vfs-overlay.yaml
OUTPUT = .build/Backstory

SOURCES = $(shell find Sources -name '*.swift' | sort)

.PHONY: build clean run

build: $(OUTPUT)

$(OUTPUT): $(SOURCES) $(VFS_OVERLAY)
	@mkdir -p .build
	swiftc -parse-as-library \
		-target $(TARGET) \
		-sdk $(SDK) \
		-vfsoverlay $(VFS_OVERLAY) \
		$(SOURCES) \
		-o $(OUTPUT)

clean:
	rm -rf .build

run: build
	./$(OUTPUT)
