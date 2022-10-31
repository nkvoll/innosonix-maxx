## WIP imctl - A tool for interacting with Innosonix Maxx Amplifiers

This is a **work in progress**, no guarantees about suitability for purpose for anything, nor that it will not damage your equipment. Use at your own risk.

The main reason for doing this was a desire to have a simple way of automatically enabling and disabling amplifier channels on an [Innosonix MAXX / MA32/D²](https://innosonix.de/maxxSeries.html?typ=ma32d2) amplifier to realize some power savings.


### Basic usage

```
$ go run ./cmd/main/main.go help
$ go run ./cmd/main/main.go examples auto-ampenable --addr <addr> --token <token>
```

#### Docker usage

An not guaranteed to be up to date version is every once in a while pushed to `nkvoll/innosonix-maxx:latest` (no version tagging) to experience with without building/installing.

Example:

```
$ docker run --rm -it nkvoll/innosonix-maxx:latest imctl help
$ docker run --rm -it nkvoll/innosonix-maxx:latest imctl examples mute --addr <your-amplifier-ip> --token <your-token>
$ docker run --rm -it nkvoll/innosonix-maxx:latest imctl examples auto-ampenable --addr <your-amplifier-ip> --token <your-token>
```

### Looking for alternatives?

- [rest.sh OpenAPI](https://rest.sh/#/openapi) could be a more generic approach to the whole CLI part for now?
