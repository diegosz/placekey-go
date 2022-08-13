# placekey-go

Unofficial port of the Python library [placekey-py](https://github.com/Placekey/placekey-py), not affiliated with the Placekey project.

## Install

```go
go get github.com/diegosz/placekey-go
```

This package requires **Go 1.18** or later.

## Prerequisites

This library uses the amazing [akhenakh/goh3](https://github.com/akhenakh/goh3) native Go h3 port build using ccgo, so CGO is not required to be enabled.

> This is still an experiment, use at your own risk

The library [uber/h3-go](https://github.com/uber/h3-go) requires [CGO](https://golang.org/cmd/cgo/) (```CGO_ENABLED=1```) in order to be built, we don't need it here.

## References

- <https://www.placekey.io>
- <https://docs.placekey.io>
- <https://docs.placekey.io/Placekey_Encoding_Specification_White_Paper.pdf>
- <https://github.com/Placekey/placekey-py>
- <https://github.com/engelsjk/placekey-go>
- <https://github.com/ringsaturn/pk>
- <https://github.com/akhenakh/goh3>
- <https://blog.nobugware.com/post/2022/surprising-result-while-transpiling-go/>

## FIXME

The TestH3ToGeoBoundary test is not working as expected. It works fine in debug mode, but fails to run in testing.

ToGeoBoundary returns empty. Something smells...

The example test ExampleH3_ToGeoBoundary is a working fine. Also running in an executable binary works fine, at least for the moment...

May be something related to the [thread-local storage](https://groups.google.com/g/golang-nuts/c/tGamryo50BY).

¯\\_(ツ)_/¯
