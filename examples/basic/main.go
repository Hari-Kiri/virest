package main

// Example: list pools like `virsh pool-list`.
//
// Other virsh → Go mappings (see storagePool/README.md):
//   pool-list --all     → conn.ListAll(0)
//   pool-info NAME      → conn.Info(name)
//   pool-define-as …    → conn.DefineAs(utilities.DefineAsParams{…}, 0)
//   pool-start NAME     → conn.Start(name, 0)
//   pool-destroy NAME   → conn.Destroy(name)
//   pool-dumpxml NAME   → conn.Dump(name, 0)

import (
	"fmt"
	"log"
	"os"

	"github.com/Hari-Kiri/virest/storagePool"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	uri := "qemu:///system"
	if v := os.Getenv("VIREST_LIBVIRT_URI"); v != "" {
		uri = v
	}

	conn, err := storagePool.Connect(uri)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	// virsh pool-list (active); use ListAll(0) for --all
	pools, err := conn.ListActive(0)
	if err != nil {
		return fmt.Errorf("list: %w", err)
	}
	fmt.Printf("found %d storage pools on %s\n", len(pools), uri)
	n := len(pools)
	for i := 0; i < n; i++ {
		p := pools[i]
		fmt.Printf("- %s (%s) state=%v\n", p.Name, p.Uuid, p.State)
	}
	return nil
}
