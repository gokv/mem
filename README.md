# gokv/mem
[![GoDoc](https://godoc.org/github.com/gokv/mem?status.svg)](https://godoc.org/github.com/gokv/mem)
[![Build Status](https://travis-ci.org/gokv/mem.svg?branch=master)](https://travis-ci.org/gokv/mem)

A naïve (read: incomplete) in-memory key-value store implementing `github.com/gokv/store`'s Store interface.

The focus is on readability and simplicity, rather than on efficiency.

## This package is not ready for production use.

### Known issues:

* The expired values are never garbage-collected
