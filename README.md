# homekit

![CI](https://github.com/MrEhbr/homekit/workflows/CI/badge.svg)
[![License](https://img.shields.io/badge/license-Apache--2.0%20%2F%20MIT-%2397ca00.svg)](https://github.com/MrEhbr/homekit/blob/master/COPYRIGHT)
[![codecov](https://codecov.io/gh/MrEhbr/homekit/branch/master/graph/badge.svg)](https://codecov.io/gh/MrEhbr/homekit)
![Made by Alexey Burmistrov](https://img.shields.io/badge/made%20by-Alexey%20Burmistrov-blue.svg?style=flat)

## Usage

```go
package main

import (
    "context"
    "net/http"

    "github.com/MrEhbr/homekit"
    "github.com/brutella/hc"
    "github.com/brutella/hc/accessory"
    "github.com/brutella/hc/log"
)

func main() {
    // create an accessory
    info := accessory.Info{Name: "Lamp"}
    ac := accessory.NewSwitch(info)

    t, err := homekit.NewTransport(12345, "config", "00102003", ac.Accessory)
    if err != nil {
        log.Info.Panic(err)
    }

    ctx, cancel := context.WithCancel(context.Background())

    hc.OnTermination(func() {
        cancel()
    })

    if err := t.Run(ctx); err != nil && err != http.ErrServerClosed {
        log.Info.Printf("transport Run error: %s", err)
    }
}
```

### Using go

```console
go get -u github.com/MrEhbr/homekit
```
## Attribution

[transport.go](./transport.go), [config.go](./config.go) are forked from the [github.com/brutella/hc](https://github.com/brutella/hc) package.

Those files are Copyright (c) 2017, Matthias Hochgatterer and subject to the conditions in [LICENSE](https://github.com/brutella/hc/blob/master/LICENSE).

Any added files (such as [storage.go](./storage.go)) are subject to the conditions in the [Homekit LICENSE file](./LICENSE)


## License

© 2020 [Alexey Burmistrov]

Licensed under the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0) ([`LICENSE`](LICENSE)). See the [`COPYRIGHT`](COPYRIGHT) file for more details.

`SPDX-License-Identifier: Apache-2.0`
