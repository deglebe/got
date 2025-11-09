TARGET	= got
DEST	= /usr/local/bin
SRC	= .

build:
	go build -o $(TARGET) $(SRC)

install: build
	sudo cp $(TARGET) $(DEST)/$(TARGET)
	sudo chmod +x $(DEST)/$(TARGET)
	@echo "installing config file..."
	@mkdir -p ~/.config/got
	@if [ ! -f ~/.config/got/config.yaml ]; then \
		cp config.yaml ~/.config/got/config.yaml && \
		echo "config installed to ~/.config/got/config.yaml"; \
		echo "edit the file to add your github token"; \
	else \
		echo "config already exists at ~/.config/got/config.yaml"; \
	fi

clean:
	rm -f $(TARGET)

uninstall:
	sudo rm -f $(DEST)/$(TARGET)
	rm -rf ~/.config/got

fmt:
	go fmt ./...

.PHONY: build install clean uninstall fmt
