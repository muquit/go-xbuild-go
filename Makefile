#====================================================================
# Requires https://github.com/muquit/go-xbuild-go for cross compiling
# for other platforms.
# Mar-29-2025 muquit@muquit.com 
#====================================================================
README_ORIG=./docs/README.md
README=./README.md
BINARY=./go-xbuild-go
VERSION := $(shell cat VERSION)
LDFLAGS := -ldflags "-w -s -X 'github.com/muquit/go-xbuild-go/pkg/version.Version=$(VERSION)'"
BUILD_OPTIONS = -trimpath
MARKDOWN_TOC_PROG=markdown-toc-go
GLOSSARY_FILE=./docs/glossary.txt
SF=./docs/synopsis.txt
VF=./docs/version.md
BADGEF=./docs/badges.md
MAIN_MD=docs/main.md

all: build build_all doc

build:
	@echo "*** Compiling ..."
	go build $(BUILD_OPTIONS) $(LDFLAGS) -o $(BINARY)

build_all: build doc
	@/bin/rm -rf ./bin
	@echo "*** Cross Compiling ...."
	# -build-args was added on v1.0.6 Sep-14-2025 
	$(BINARY) -build-args '$(BUILD_OPTIONS) $(LDFLAGS)' \
		-additional-files 'build-config.json'

doc: gen_files
	@echo "*** Generating README.md with TOC ..."
	@touch $(README)
	$(MARKDOWN_TOC_PROG) -i $(MAIN_MD) -o $(README) --glossary ${GLOSSARY_FILE} -pre-toc-file $(BADGEF) -f


gen_files: gen_synopsis ver

gen_synopsis: build
	echo '## Synopsis' > $(SF)
	echo '```' >> $(SF)
	$(BINARY) -h >> $(SF) 2>&1
	echo '```' >> $(SF)

ver:
	echo "## Latest Version ($(VERSION))" > $(VF)
	echo "The current version is $(VERSION)" >> $(VF)
	echo "Please look at @CHANGELOG@ for what has changed in the current version.">> $(VF)

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
