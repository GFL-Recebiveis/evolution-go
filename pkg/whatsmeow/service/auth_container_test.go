package whatsmeow_service

import (
	"context"
	"os"
	"testing"
)

// Reproduz o vazamento: várias chamadas a StartClient devem reaproveitar o mesmo pool.
func TestGetAuthContainerReuseiaPool(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_AUTH_DB")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_AUTH_DB não definido")
	}
	first, err := getAuthContainer(context.Background(), "postgres", dsn, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 50; i++ {
		c, err := getAuthContainer(context.Background(), "postgres", dsn, nil)
		if err != nil {
			t.Fatal(err)
		}
		if c != first {
			t.Fatalf("chamada %d criou um container novo (vazamento)", i)
		}
	}
	if got := authContainerDB.Stats().OpenConnections; got > authMaxOpenConns {
		t.Fatalf("%d conexões abertas, teto %d", got, authMaxOpenConns)
	}
}
