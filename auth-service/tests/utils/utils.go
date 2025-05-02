package utils

import (
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
	tc "github.com/testcontainers/testcontainers-go"

	"context"
	"testing"
	"time"
)

type TestDatabase struct {
	Host		string
	Port		string
	DBName		string
	User		string
	Password	string
}

type TestUserService struct {
	Host	string
	Port	string
}

func SetupTestNetwork(
	ctx context.Context, t *testing.T, netName string,
) (network tc.Network, cleanup func()) {
	t.Helper()

	network, err := tc.GenericNetwork(ctx, tc.GenericNetworkRequest{
		NetworkRequest: tc.NetworkRequest{
			Name:		netName,
			CheckDuplicate:	true,
		},
	})
	require.NoError(t, err, "Failed to create a network")

	cleanup = func() {
		if err := network.Remove(ctx); err != nil {
			t.Errorf("Failed to remove test network: %v", err)
		}
	}

	return network, cleanup
}

func SetupTestDatabase(
	ctx context.Context,
	t *testing.T,
	network tc.Network,
	netName string,
) (db *TestDatabase, cleanup func()) {
	t.Helper()

	host := "postgres"
	dbName := "testdb"
	user := "testuser"
	password := "testpass"

	req := tc.ContainerRequest{
		Image: "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB": dbName,
			"POSTGRES_USER": user,
			"POSTGRES_PASSWORD": password,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(10*time.Second),
		Networks: []string{netName},
		NetworkAliases: map[string][]string{netName: {host}},
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started: true,
	})
	require.NoError(t, err, "Failed to start PostgreSQL container")

	cleanup = func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate PostgreSQL container: %v", err)
		}
	}

	testDatabase := TestDatabase{
		Host: host,
		Port: "5432",
		DBName: dbName,
		User: user,
		Password: password,
	}

	return &testDatabase, cleanup
}

func SetupTestUserService(
	ctx context.Context,
	t *testing.T,
	db *TestDatabase,
	network tc.Network,
	netName string,
) (client *TestUserService, cleanup func()) {
	t.Helper()

	env := map[string]string{
		"POSTGRES_HOST":	db.Host,
		"POSTGRES_PORT":	db.Port,
		"POSTGRES_DB":		db.DBName,
		"POSTGRES_USER":	db.User,
		"POSTGRES_PASSWORD":	db.Password,
	}

	req := tc.ContainerRequest{
		FromDockerfile: tc.FromDockerfile{
			Context:	"../../",
			Dockerfile:	"Dockerfile",
		},
		ExposedPorts:	[]string{"8080/tcp"},
		Env:		env,
		Networks:	[]string{netName},
		WaitingFor:	wait.ForHTTP("/hello").WithStartupTimeout(30*time.Second),
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest:	req,
		Started:		true,
	})
	require.NoError(t, err, "Failed to start user service container")

	cleanup = func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate user service container: %v", err)
		}
	}

	host, err := container.Host(ctx)
	require.NoError(t, err, "Failed to get the user service host")

	port, err := container.MappedPort(ctx, "8080")
	require.NoError(t, err, "Failed to get the user service port")

	testUserService := TestUserService{
		Host: host,
		Port: port.Port(),
	}

	return &testUserService, cleanup
}
