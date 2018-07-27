# gokv/mem
[![GoDoc](https://godoc.org/github.com/gokv/mem?status.svg)](https://godoc.org/github.com/gokv/mem)
[![Build Status](https://travis-ci.org/gokv/mem.svg?branch=master)](https://travis-ci.org/gokv/mem)

An in-memory key-value store implementing the `github.com/gokv/store` Store interface.

The focus is on readability and simplicity.

## Usage

```Go
func main() {
	s := mem.New()
	defer s.Close() // close the mem.Store to avoid leaking goroutines

	err := s.SetWithTimeout(context.Background(), "key", Value{p:1}, timeout)
	if err != nil {
		panic(err)
	}

	var v Value // Value is a type that implements json.Marshaler/Unmarshaler
	ok, err := s.Get(context.Background(), "key", &v)

	if err != nil {
		panic(err)
	}

	if !ok {
		panic(errors.New("Value not found!"))
	}

	fmt.Println("The retrieved value is %q", v)
}
```
