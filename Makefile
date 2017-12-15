all: install

install: build
	mkdir -p /usr/share/ddnsd
	cp ./build/ddnsd /usr/share/ddnsd/
	cp -R ./examples /usr/share/ddns/
	cp ./systemd/cloudflare.ddns.service /usr/lib/systemd/system/

build:
	go build -o ./build/ddnsd ./cmd/ddnsd

.PHONY: clean
clean:
	rm -rf ./build

.PHONY: uninstall
uninstall:
	rm -rf /usr/share/ddns
