//go:build js && wasm

package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/syumai/workers/cloudflare"
	"github.com/syumai/workers/cloudflare/sockets"
)

type HyperdriveProvider struct {
	bindingName          string
	expectedDatabaseName string
}

func NewHyperdriveProvider(bindingName, expectedDatabaseName string) *HyperdriveProvider {
	return &HyperdriveProvider{
		bindingName:          bindingName,
		expectedDatabaseName: expectedDatabaseName,
	}
}

func (p *HyperdriveProvider) Acquire(ctx context.Context) (*Connection, error) {
	binding := cloudflare.GetBinding(p.bindingName)
	if binding.IsUndefined() || binding.IsNull() {
		return nil, fmt.Errorf("Cloudflare binding %s is not configured", p.bindingName)
	}

	connectionString := binding.Get("connectionString").String()
	if strings.TrimSpace(connectionString) == "" {
		return nil, fmt.Errorf("Cloudflare binding %s has no connection string", p.bindingName)
	}
	if err := validateDatabaseName(connectionString, p.expectedDatabaseName); err != nil {
		return nil, err
	}

	connectionConfig, err := pgx.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parse Hyperdrive connection string: %w", err)
	}
	connectionConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	// Cloudflare's TCP socket API resolves the host. Returning it unchanged here
	// prevents Go's native resolver from running in the Wasm runtime.
	connectionConfig.Config.LookupFunc = func(_ context.Context, host string) ([]string, error) {
		return []string{host}, nil
	}
	connectionConfig.Config.DialFunc = func(
		ctx context.Context,
		_ string,
		address string,
	) (net.Conn, error) {
		return sockets.Connect(ctx, address, nil)
	}

	connection, err := pgx.ConnectConfig(ctx, connectionConfig)
	if err != nil {
		return nil, fmt.Errorf("connect through Hyperdrive: %w", err)
	}

	return NewConnection(connection, func(closeContext context.Context) error {
		return connection.Close(closeContext)
	})
}

func validateDatabaseName(connectionString, expected string) error {
	parsed, err := url.Parse(connectionString)
	if err != nil {
		return fmt.Errorf("parse Hyperdrive database name: %w", err)
	}
	databaseName, err := url.PathUnescape(strings.TrimPrefix(parsed.EscapedPath(), "/"))
	if err != nil {
		return fmt.Errorf("parse Hyperdrive database name: %w", err)
	}
	if databaseName != expected {
		return fmt.Errorf(
			"Hyperdrive selects database %q, but this environment expects %q",
			databaseName,
			expected,
		)
	}
	return nil
}
