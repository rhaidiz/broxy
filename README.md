# Broxy

[![License](https://img.shields.io/badge/license-GPLv3-blue.svg)](https://raw.githubusercontent.com/rhaidiz/broxy/master/COPYING)
![status](https://img.shields.io/badge/status-in_progress-yellow.svg)

Broxy is an open source intercept proxy written in Go. It makes use of [goproxy](https://github.com/elazarl/goproxy) as core proxy implementation and the interface is implemented with a [Qt wrapper](https://github.com/therecipe/qt) for Go.

**DISCLAIMER:** this is still an alpha software so bear with me.

![intercept](https://github.com/rhaidiz/broxy/raw/master/media/main.png)

## Features
Broxy is currently a work in progress project. The current version provides the following features:

* Interceptor
* History with filters
* Repeater
* Persistent sessions

# Installation

## Grab a Binary
Grab a compiled version of Broxy [here](https://github.com/rhaidiz/broxy/releases).

## Compiling from Sources
To compile Broxy make sure you have the following installed and properly configured:

* [Go](https://golang.org/doc/install)
* [Qt 5.13 and Qt wrapper](https://github.com/therecipe/qt/wiki/Installation)

Once Go, Qt 5.13 and the Qt wrapper are ready, just do:

    go get github.com/rhaidiz/broxy
    cd $GOPATH/src/github.com/rhaidiz/broxy
    make build 

Once the compilation is finished, you'll have the binary inside the folder `deploy` ... 🤞!

# License
Broxy is released under the GPLv3 license.

