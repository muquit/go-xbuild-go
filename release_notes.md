# go-xbuild-go v1.0.5

**Major feature**: Multi-binary Go project support with full backward compatibility.

## New Features
- **Multi-target builds**: New `-config` flag for JSON configuration files
- **Multiple main packages**: Build `cmd/cli`, `cmd/server`, etc. in one command  
Please look at
[go-multi-main-example](https://github.com/muquit/go-multi-main-example) for
an example.
- **Per-target customization**: Individual ldflags, build flags, and additional files
- **New `-list-targets` flag**: Show available build targets

## Usage
Legacy projects work unchanged. For multi-binary projects:
```bash
go-xbuild-go -config build-config.json
```

Please look at [ChangeLog](ChangeLog.md) for details on what has changed in the current version. The binaries are cross-compiled with https://github.com/muquit/go-xbuild-go. Do not forget to check checksums of the archives before using.

Thanks!

~
