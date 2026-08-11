package service

import (
	"context"
	"testing"

	"github.com/I-Frostbyte/rvpay-go/clients/db/repo"
	"github.com/I-Frostbyte/rvpay-go/clients/db/sqlc"
	clientsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/clientsgrpc"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockClientRepo is a test double for ClientRepo
type mockClientRepo struct {
	clients map[string]sqlc.Client
}

func newMockClientRepo() *mockClientRepo {
	return &mockClientRepo{
		clients: make(map[string]sqlc.Client),
	}
}

func (m *mockClientRepo) Create(ctx context.Context, name string, status sqlc.ClientStatus) (sqlc.Client, error) {
	client := sqlc.Client{
		ID:         uuid.New(),
		ClientName: name,
		Status:     status,
	}
	m.clients[client.ID.String()] = client
	return client, nil
}

func (m *mockClientRepo) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Client, error) {
	client, ok := m.clients[id.String()]
	if !ok {
		return sqlc.Client{}, repo.ErrNotFound
	}
	return client, nil
}

func (m *mockClientRepo) GetByName(ctx context.Context, name string) (sqlc.Client, error) {
	for _, client := range m.clients {
		if client.ClientName == name {
			return client, nil
		}
	}
	return sqlc.Client{}, repo.ErrNotFound
}

func (m *mockClientRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.ClientStatus) (sqlc.Client, error) {
	client, ok := m.clients[id.String()]
	if !ok {
		return sqlc.Client{}, repo.ErrNotFound
	}
	client.Status = status
	m.clients[id.String()] = client
	return client, nil
}

func (m *mockClientRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := m.clients[id.String()]; !ok {
		return repo.ErrNotFound
	}
	delete(m.clients, id.String())
	return nil
}

func (m *mockClientRepo) List(ctx context.Context, limit, offset int32) ([]sqlc.Client, error) {
	clients := make([]sqlc.Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	return clients, nil
}

func (m *mockClientRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.clients)), nil
}

func (m *mockClientRepo) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	_, ok := m.clients[id.String()]
	return ok, nil
}

func (m *mockClientRepo) ListActive(ctx context.Context, limit, offset int32) ([]sqlc.Client, error) {
	clients := make([]sqlc.Client, 0)
	for _, client := range m.clients {
		if client.Status == sqlc.ClientStatusACTIVE {
			clients = append(clients, client)
		}
	}
	return clients, nil
}

func TestCreateClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  *clientsgrpc.CreateClientRequest
		code codes.Code
	}{
		{name: "missing request", code: codes.InvalidArgument},
		{name: "empty name", req: &clientsgrpc.CreateClientRequest{Name: ""}, code: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			clientRepo := newMockClientRepo()
			svc := NewClientsServiceImpl(clientRepo, zerolog.Nop())

			_, err := svc.CreateClient(context.Background(), tt.req)
			if got := status.Code(err); got != tt.code {
				t.Fatalf("status code = %s, want %s", got, tt.code)
			}
		})
	}
}

func TestGetClient(t *testing.T) {
	t.Parallel()

	clientRepo := newMockClientRepo()
	service := NewClientsServiceImpl(clientRepo, zerolog.Nop())

	client, err := clientRepo.Create(context.Background(), "Test Client", sqlc.ClientStatusACTIVE)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	resp, err := service.GetClient(context.Background(), &clientsgrpc.GetClientRequest{
		Id: client.ID.String(),
	})
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}

	if resp.Client.Name != "Test Client" {
		t.Fatalf("client name = %s, want Test Client", resp.Client.Name)
	}
}

func TestDeleteClient(t *testing.T) {
	t.Parallel()

	clientRepo := newMockClientRepo()
	service := NewClientsServiceImpl(clientRepo, zerolog.Nop())

	client, err := clientRepo.Create(context.Background(), "Test Client", sqlc.ClientStatusCLOSED)
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	_, err = service.DeleteClient(context.Background(), &clientsgrpc.DeleteClientRequest{
		Id: client.ID.String(),
	})
	if err != nil {
		t.Fatalf("DeleteClient failed: %v", err)
	}

	_, err = clientRepo.GetByID(context.Background(), client.ID)
	if err != repo.ErrNotFound {
		t.Fatal("deleted client should not be found")
	}
}
