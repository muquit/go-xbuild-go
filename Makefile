#====================================================================
# Requires https://github.com/muquit/go-xbuild-go for cross compiling
# for other platforms.
# Mar-29-2025 muquit@muquit.com 
#====================================================================
README_ORIG=./docs/README.md
README=./README.md
BINARY=./go-xbuild-go
GEN_TOC_PROG=markdown-toc-go

all: build build_all doc

build:
	@echo "*** Compiling ..."
	go build -o $(BINARY)

build_all:
	@echo "*** Cross Compiling ...."
	go-xbuild-go

doc:
	@echo "*** Generating README.md with TOC ..."
	chmod 600 $(README)
	$(GEN_TOC_PROG) -i $(README_ORIG) -o $(README) -f
	chmod 444 $(README)

clean:
	/bin/rm -f $(BINARY)
	/bin/rm -rf ./bin
