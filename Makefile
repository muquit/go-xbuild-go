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

.PHONY: all build build_all doc docs gen_files gen_synopsis ver release check_github_token clean

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
	$(MARKDOWN_TOC_PROG) -i docs/ChangeLog.md -o ./ChangeLog.md --glossary docs/glossary.txt -f -no-credit

docs: doc


gen_files: gen_synopsis ver

gen_synopsis: build
	echo '# Synopsis' > $(SF)
	echo '```' >> $(SF)
	$(BINARY) -h >> $(SF) 2>&1
	echo '```' >> $(SF)

ver:
	echo "# Latest Version ($(VERSION))" > $(VF)
	echo "The current version is $(VERSION)" >> $(VF)
	echo "Please look at @CHANGELOG@ for what has changed in the current version.">> $(VF)

# make sure:
#  - to run: make clean
#  - to run: make doc
#  - to check VERSION file
#  - run 'make build_all' before release
#  - release_notes.md exists in cwd
release: check_github_token gen_release_notes
	@echo "*** Releasing on github ..."
	$(BINARY) -release

gen_release_notes:
	./scripts/mk_release_notes.sh

# check if GITHUB_TOKEN is set and valid, fail the build otherwise
check_github_token:
	@if [ -z "$(GITHUB_TOKEN)" ]; then \
		echo "*** ERROR: GITHUB_TOKEN is not set"; \
		exit 1; \
	fi
	@status=$$(curl -s -o /tmp/check_github_token.$$$$.json -w '%{http_code}' \
        -H "Authorization: token $(GITHUB_TOKEN)" https://api.github.com/user); \
	if [ "$$status" != "200" ]; then \
		echo "*** ERROR: GITHUB_TOKEN is not valid (HTTP $$status)"; \
		cat /tmp/check_github_token.$$$$.json; \
		rm -f /tmp/check_github_token.$$$$.json; \
		exit 1; \
	fi; \
	jq '{login, name, type}' < /tmp/check_github_token.$$$$.json; \
	rm -f /tmp/check_github_token.$$$$.json
	@curl -sI -H "Authorization: token $(GITHUB_TOKEN)" \
        https://api.github.com/user | grep -i x-oauth-scopes

clean:
	/bin/rm -f $(BINARY)
	/bin/rm -rf ./bin
