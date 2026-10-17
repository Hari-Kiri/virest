//go:build integration

package storagePool_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Hari-Kiri/virest/storagePool"
	"github.com/Hari-Kiri/virest/utilities"
)

func TestIntegrationLifecycle(t *testing.T) {
	uri := os.Getenv("VIREST_LIBVIRT_URI")
	if uri == "" {
		uri = "qemu:///system"
	}
	dir := os.Getenv("VIREST_TEST_POOL_DIR")
	if dir == "" {
		dir = t.TempDir()
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	conn, err := storagePool.Connect(uri)
	if err != nil {
		t.Skipf("libvirt unavailable at %s: %v", uri, err)
	}
	defer conn.Close()

	name := "virest-integration-" + filepath.Base(dir)
	uuid, err := conn.Define(utilities.StoragePool{
		Type: utilities.PoolTypeDir,
		Name: name,
		Target: &utilities.StoragePoolTarget{
			Path: dir,
		},
	}, 0)
	if err != nil {
		t.Fatalf("define: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Destroy(uuid)
		_ = conn.Delete(uuid, utilities.Flags.Delete.Normal)
		_ = conn.Undefine(uuid)
	})

	if err := conn.Build(uuid, utilities.Flags.Build.New); err != nil {
		// Some dir pools do not need build; continue if already built.
		t.Logf("build: %v", err)
	}
	if err := conn.Start(uuid, utilities.Flags.Create.Normal); err != nil {
		t.Fatalf("start: %v", err)
	}
	info, err := conn.Info(uuid)
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if info.Name != name {
		t.Fatalf("name=%s", info.Name)
	}
	list, err := conn.List(0, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	found := false
	for _, p := range list {
		if p.Uuid == uuid {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("defined pool not in list")
	}
	if err := conn.Refresh(uuid); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if err := conn.Autostart(uuid, false); err != nil {
		t.Fatalf("autostart: %v", err)
	}
	if err := conn.Destroy(uuid); err != nil {
		t.Fatalf("destroy: %v", err)
	}
}

func TestIntegrationSkipWithoutDisk(t *testing.T) {
	if os.Getenv("VIREST_TEST_DISK") == "" {
		t.Skip("set VIREST_TEST_DISK to run destructive disk tests")
	}
	t.Fatal("disk integration tests are environment-specific; implement against VIREST_TEST_DISK when needed")
}
