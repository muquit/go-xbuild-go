## Installing using Homebrew on Mac/Linux

You will need to install [Homebrew](https://brew.sh/) first.

### Install

First install the custom tap, then trust it. Homebrew 6.0+ refuses to load
formulae from third-party taps until they are explicitly trusted.

```
brew tap muquit/go-xbuild-go https://github.com/muquit/go-xbuild-go.git
brew trust muquit/go-xbuild-go
brew install go-xbuild-go
```

Or tap, trust and install in one go:
```
brew tap muquit/go-xbuild-go https://github.com/muquit/go-xbuild-go.git
brew trust muquit/go-xbuild-go
brew install muquit/go-xbuild-go/go-xbuild-go
```

### Upgrade
```
brew upgrade go-xbuild-go
```

### Uninstall
```
brew uninstall go-xbuild-go
```

### Remove the tap
```
brew untap muquit/go-xbuild-go
```
