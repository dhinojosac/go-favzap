
TARGET = favzap
BUILDIR = build

all:
	go build -o $(BUILDIR)/$(TARGET)

.PHONY: dirs clean build-win

dirs:
	mkdir -p build

build-win: dirs
	go build -o $(BUILDIR)/$(TARGET).exe

build-linux: dirs
	go build -o $(BUILDIR)/$(TARGET)

clean:
	rm -rf $(BUILDIR)