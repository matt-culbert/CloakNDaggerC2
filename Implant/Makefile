# Generate a random UUID-like number (e.g., a random integer between 1000 and 9999)
RAND_NUM := $(shell powershell -Command "[System.Random]::new().Next(1000,9999)")
COMPRESSION := $(shell python preprocessor.py)

# Default build rule
http:
	@echo "Building default options with random ID: $(RAND_NUM)"
	@echo "Building with compression? $(COMPRESSION)"
	@if "$(COMPRESSION)"=="y" ( \
		go build -trimpath -ldflags "-X main.CompUUID=$(RAND_NUM) -s -w" -tags "withComp withHttp" ./base_config/daemon  \
	) else ( \
		go build -trimpath -ldflags "-X main.CompUUID=$(RAND_NUM) -s -w" -tags "withHttp" ./base_config/daemon \
	)

# DNS build rule
dns:
	@echo "Building for DNS with random ID: $(RAND_NUM)"
	@echo "Building with compression? $(COMPRESSION)"
	@if "$(COMPRESSION)"=="y" ( \
		go build -trimpath -ldflags "-X main.CompUUID=$(RAND_NUM) -s -w" -tags "withComp withDns" ./base_config/daemon  \
	) else ( \
		go build -trimpath -ldflags "-X main.CompUUID=$(RAND_NUM) -s -w" -tags "withDns" ./base_config/daemon \
	)

# Clean rule to remove the binary
clean:
	rm -f daemon.exe