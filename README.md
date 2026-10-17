# ViRest

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

Go library for managing **libvirt storage pools** from application code. Import `storagePool`, connect to a hypervisor URI, and call typed methods — no HTTP server required.

An optional `cmd/testserver` harness is included for manual/e2e HTTP testing (multi-user JWT auth, TLS, OpenAPI).

## Install

```bash
go get github.com/Hari-Kiri/virest/storagePool
```

Requires **CGO**, **libvirt** development headers, and a reachable libvirt daemon.

```bash
export CGO_ENABLED=1
# Debian/Ubuntu packages commonly needed:
# libvirt-dev gcc
```

## Quick start

```go
package main

import (
	"fmt"
	"log"

	"github.com/Hari-Kiri/virest/storagePool"
)

func main() {
	conn, err := storagePool.Connect("qemu:///system")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	pools, err := conn.List(0, 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(pools), "pools")
}
```

See also [`examples/basic`](examples/basic).

## Library API

Pool models use **`utilities.StoragePool`** (and related types) — callers do not import `libvirtxml`. Full method wiki: [`storagePool/README.md`](storagePool/README.md).

| Method | Purpose |
|--------|---------|
| `Connect` / `Close` | Open and release a hypervisor connection |
| `Define` / `DefineAs` / `CreateTransient` / `CreateAs` | Define or create pools (model or `-as` params) |
| `Build` / `Create` / `Start` | Build and start defined pools |
| `Destroy` / `Undefine` / `Delete` | Stop, remove definition, delete resources |
| `Refresh` / `Autostart` | Refresh volumes; set autostart |
| `List` / `ListActive` / `ListInactive` / `ListAll` / `Info` / `Dump` | Inspect pools |
| `Name` / `UUIDByName` | Resolve name ↔ UUID |
| `Capabilities` / `FindSources` / `FindSourcesAs` | Discovery helpers |
| `WaitEvent` / `StreamEvents` | Lifecycle/refresh events |

`StreamEvents` is transport-agnostic (`context` + callback). For event APIs, register and run the process-wide libvirt event loop once:

```go
storagePool.RegisterEventLoop()
```

## Tests

Unit tests (no live libvirt; mocked interfaces):

```bash
export CGO_ENABLED=1
go test ./storagePool/ ./utilities/ ./cmd/testserver/...
```

Integration tests (live libvirt):

```bash
export CGO_ENABLED=1
export VIREST_LIBVIRT_URI=qemu:///system   # optional
export VIREST_TEST_POOL_DIR=/tmp/virest-pool  # optional
go test -tags=integration ./storagePool/
```

Destructive disk tests run only when `VIREST_TEST_DISK` is set.

## Benchmarks

```bash
go test ./storagePool/ -bench=. -benchmem
```

## Optional testserver

```bash
cd cmd/testserver
cp .env_dummy .env
cp users.example.yaml users.yaml
# edit secrets / users, then:
go run .
```

Swagger UI: `/swagger/`.

## License

MIT
