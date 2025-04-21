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

build_all: build
	@/bin/rm -rf ./bin
	@echo "*** Cross Compiling ...."
	$(BINARY)

doc:
	@echo "*** Generating README.md with TOC ..."
	chmod 600 $(README)
	$(GEN_TOC_PROG) -i $(README_ORIG) -o $(README) -pre-toc-file docs/badges.md -f
	chmod 444 $(README)

# make sure:
#  - to run: make clean
#  - to run: make doc
#  - to check VERSION file
#  - run 'make build_all' before release
#  - release_notes.md exists in cwd
release:
	@echo "*** Releasing on github ..."
	$(BINARY) -release

clean:
	/bin/rm -f $(BINARY)
	/bin/rm -rf ./bin
