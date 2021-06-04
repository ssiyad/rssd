GO = /usr/bin/go

rssd:
	$(GO) build -ldflags "-s -w"

install:
	sudo install -Dm755 ./rssd /usr/bin/rssd
	sudo install -Dm755 ./rssd.service /usr/lib/systemd/user/
	sudo install -Dm755 ./rssd.timer /usr/lib/systemd/user/

uninstall:
	sudo rm /usr/bin/rssd
	sudo rm /usr/lib/systemd/user/rssd.service
	sudo rm /usr/lib/systemd/user/rssd.timer

clean:
	rm -rf out/