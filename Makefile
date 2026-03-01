build:
	go build -o writebetterer .

run:
	go run .

deploy: build install-man install-completion
	cp writebetterer ~/Dropbox/Typinator/Includes/Text/
	cp writebetterer ~/.local/bin/

install-man:
	install -d /usr/local/share/man/man1
	install -m 644 writebetterer.1 /usr/local/share/man/man1/writebetterer.1

install-completion:
	install -d /usr/local/share/zsh/site-functions
	install -m 644 _writebetterer /usr/local/share/zsh/site-functions/_writebetterer
