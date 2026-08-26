// Command prefetch-embedded populates the embedded-postgres binary cache
// (~/.embedded-postgres-go) so tests that need it never depend on the network
// at test time.
//
// It starts and immediately stops a real PostgreSQL 16 — the same version the
// test harness (backend/v3/.../postgres/embedded) pins — because the library
// downloads and unpacks the binary as a side effect of the first Start. A
// download-only mode does not exist upstream, and reimplementing their fetch
// logic here would drift from what the tests actually resolve.
//
// Run via: bash dev/preflight.sh --fetch-embedded
package main

import (
	"fmt"
	"net"
	"os"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
)

func main() {
	path, err := os.MkdirTemp("", "nomen-prefetch-embedded-*")
	if err != nil {
		fail("cannot create a temp dir: %v", err)
	}
	defer os.RemoveAll(path)

	// A random free port, so this never trips over the dev cluster on 5433 or
	// anything else on the box.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fail("cannot pick a port: %v", err)
	}
	port := uint32(l.Addr().(*net.TCPAddr).Port)
	_ = l.Close()

	start := time.Now()
	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Version(embeddedpostgres.V16).
		Port(port).
		RuntimePath(path).
		StartTimeout(90 * time.Second))

	if err := pg.Start(); err != nil {
		fail("embedded postgres did not start: %v", err)
	}
	if err := pg.Stop(); err != nil {
		fail("embedded postgres started but did not stop cleanly: %v", err)
	}
	fmt.Printf("embedded postgres 16 cached and verified in %s\n", time.Since(start).Round(time.Millisecond))
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "prefetch-embedded: "+format+"\n", args...)
	os.Exit(1)
}
