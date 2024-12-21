# kjudge

[![Build and Test](https://github.com/natsukagami/kjudge/workflows/Build%20and%20Test/badge.svg)](https://github.com/natsukagami/kjudge/actions?query=workflow%3A%22Build+and+Test%22+branch%3Amaster)
[![Docker Image Version (latest semver)](https://img.shields.io/docker/v/natsukagami/kjudge?logo=docker&sort=semver)](https://hub.docker.com/r/natsukagami/kjudge)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/natsukagami/kjudge?logo=github&sort=semver)](https://github.com/natsukagami/kjudge/releases)
[![godoc](https://godoc.org/github.com/natsukagami/kjudge?status.svg)](https://godoc.org/github.com/natsukagami/kjudge)
[![License is AGPLv3](https://img.shields.io/badge/license-AGPLv3-blue)](https://github.com/natsukagami/kjudge/src/branch/master/LICENSE)
[![Matrix Chatroom](https://img.shields.io/matrix/kjudge:matrix.org)](https://matrix.to/#/#kjudge:matrix.org)

- [kjudge](#kjudge)
  - [Project Goals](#project-goals)
  - [Quick start](#quick-start)
    - [Need more details?](#need-more-details)
  - [Runtime dependencies](#runtime-dependencies)
  - [Command line switches](#command-line-switches)
  - [Build Instructions](#build-instructions)
    - [Production build](#production-build)
    - [Development build](#development-build)
  - [Directory Structure](#directory-structure)
  - [License](#license)

## Project Goals

- As lightweightedly deployable as possible (single binary, minimal dependencies, Docker-compatible)
- User friendly
- Doesn't get in the way (take minimal resources)

## Quick start

The fastest way to get kjudge up and runnning is through the [official Docker image](https://hub.docker.com/r/natsukagami/kjudge).

It contains:

- kjudge itself
- Compilers for C++, Pascal, Java, Go, Rust, Python 2 and 3
- The [`isolate`](https://github.com/ioi/isolate) sandbox

Please checkout the [wiki](https://github.com/natsukagami/kjudge/wiki/Docker-Installation) for more information on how to get it up and running.

### Need more details?

Check out the [wiki](https://github.com/natsukagami/kjudge/wiki) or join our [chatroom](https://matrix.to/#/#kjudge:matrix.org)!

## Runtime dependencies

It should run on all platforms Go compiles to.

Required binaries:

- [`isolate`](https://github.com/ioi/isolate): The recommended sandbox.
  This is actually _optional_, however the only alternate sandbox implementation
  available (as of now) is the **raw** sandbox, which **DOES NOT PROVIDE ANY
  GUARDS AGAINST FOREIGN CODE** (which makes it okay to run when you are the
  only user).
  `isolate` is available on Linux systems only.

## Command line switches

```sh
> ./kjudge -h
Usage of ./kjudge:
  -file string
    	Path to the database file. (default "kjudge.db")
  -https string
    	Path to the directory where the HTTPS private key (kjudge.key) and certificate (kjudge.crt) is located. If omitted or empty, HTTPS is disabled.
  -port int
    	The port for the server to listen on. (default 8088)
  -sandbox string
    	The sandbox implementation to be used (isolate, raw). If anything other than 'raw' is given, isolate is used. (default "isolate")
  -verbose
    	Log every http requests
```

## Build Instructions

Warning: Windows support for kjudge is a WIP (and by that we mean machine-wrecking WIP). Run at your own risk.

External Dependencies:

```yaml
cc:      anything that compiles SQLite 3
go:      >= 1.16
node.js: >= 18
yarn:    >= 1
```

Go dependencies: See `tools.go`.
All Go dependencies can be installed with

```sh
scripts/install_tools.sh
```

### Production build

Build steps:

```sh
scripts/production_build.sh
```

### Development build

**First time contributors**: Please check out the [Code of Conduct](./CODE_OF_CONDUCT.md) and [Contributor Guidelines](./CONTRIBUTING.md)!

First, start the template generator and updater with

```bash
cd frontend
yarn && yarn dev
```

Note that it would block and watch for any updates on the front-end templates.

Now open a new terminal, run

```bash
go generate
```

to generate the models and load a development (live-reloading) version of the template packager.

Finally, run

```bash
go run cmd/kjudge/main.go
```

to run kjudge.

## Directory Structure

```yaml
embeds:   # Static assets that gets compiled into the binary
    templates # Generated templates from frontend
    assets:
        - sql # SQL migration schemas
cmd:          # Main commands
    - kjudge  # Main compile target
    - migrate # Database migration tool, useful for development
db # Database interaction library
docker    # Dockerfile and other docker-related packaging handlers
scripts   # Scripts that helps automating builds
models:          # Database entities
    - generate   # Generator for models
    - verify     # Model verification helpers
frontend:   # Template files and front-end related source codes
    - html  # HTML [template] files
    - css   # CSS files
    - ts    # TypeScript files
worker:        # Automatic judging logic
    - raw      # Raw and isolate are 2 sandbox implementations
    - isolate
server:        # Root of server logic
    - auth     # Authentication logic and session handlers
    - template # Template resolver and renderer implementation
    - user     # /user page handling and contexts
    - admin    # /admin (Admin Panel) page handling and contexts
    - contests # /contests (main contest UI) page handling and contexts
test: # Go code testing handling logic and data
    - integration # Integration tests
tests # Test-handling logic
```

## License

[**GNU Affero General Public License v3.0**](https://choosealicense.com/licenses/agpl-3.0), which **disallows distributing closed-source versions**, so don't do it.
