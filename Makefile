RELEASE=$(shell git describe --always --long --dirty)

clean:
	@rm -rf ./dist

build: clean
	@goxc -pv=v$(version)-$(RELEASE)

version:
	@echo $(RELEASE)
