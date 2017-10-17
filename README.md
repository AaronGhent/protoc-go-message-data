# protoc-go-message-data

## Why?

Golang [protobuf](https://github.com/golang/protobuf) doesn't support
[custom tags to generated structs](https://github.com/golang/protobuf/issues/52). This
script injects custom tags to generated protobuf files, useful for
things like validation struct tags.

## Install

```
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt-get update
sudo apt-get install golang-go
```

### Install protobuf3
```
INSTALL_SRC_PACKAGE=${SRC_PACKAGES:-$HOME/source/.packages}

mkdir -p ${INSTALL_SRC_PACKAGE}
cd ${INSTALL_SRC_PACKAGE}

sudo apt-get install autoconf automake libtool curl make g++ unzip

git clone https://github.com/google/protobuf

cd protobuf/

./autogen.sh
./configure
make
make check
sudo make install
sudo ldconfig

cd ..
```

## Install go deps
```
export GOPATH="${HOME}/go"

go get -u github.com/golang/protobuf/protoc-gen-go
```

## Usage

Add a comment with syntax `// @name: alladin`
before fields to add custom tag to in .proto files.

Example:

```
// file: test.proto
syntax = "proto3";

package pb;

message IP {
  // @name: alladin
  string Address = 1;
}
```

Generate with protoc command as normal.

```
protoc --go_out=. test.proto
```

Run `protoc-go-message-data` with generated file `test.pb.go`.

```
protoc-go-message-data -input=./test.pb.go
```

