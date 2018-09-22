GOLINT_OPTS := -min_confidence 1.0 -set_exit_status


.PHONY: all
all: jscan


jscan: cmd/jscan/jscan.go
	go build cmd/jscan/jscan.go
	golint $(GOLINT_OPTS) cmd/jscan/jscan.go


.PHONY: clean
clean:
	- rm jscan
	- rm *.exe
