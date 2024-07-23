# goutils

## install

```sh
go get -u github.com/117503445/goutils
```

## usage

```go
package main

import (
    "fmt"

    "github.com/117503445/goutils"
    "github.com/rs/zerolog/log"
)

func main() {
    // init zerolog
    goutils.InitZeroLog()
    log.Info().Msg("hello world")

    // run `ls -l` in /tmp
    if err := goutils.CMD("/tmp", "ls", "-l"); err != nil {
        log.Error().Err(err).Msg("run cmd failed")
    }
}
```
