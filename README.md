# Golang iRacing SDK

Golang implementation of iRacing SDK

## Install

You need a gcc compiler to build the SDK, Golang does not have (as far as I know) unsafe low level access to memory map files and windows broadcast events, so CGO is used to bridge this functions with C native ones.
The easiest way is to install MiniGw for 64 bits: https://sourceforge.net/projects/mingw-w64/

With a gcc compiler in place, you can follow the standard path get to external libs in Go
1. Execute `go get github.com/hfoxy/iracing-sdk`

## Examples

* [Standard](examples/windows.go): Normal usage of SDK

* [Mock](examples/mock/mock.go): Read data from a mocked replay file
