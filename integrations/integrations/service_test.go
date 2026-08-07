package integrations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/I-Frostbyte/rvpay-go/integrations/db/repo/mocks"
	"github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc"
	sqlcmocks "github.com/I-Frostbyte/rvpay-go/integrations/db/sqlc/mocks"
	integrationsgrpc "github.com/I-Frostbyte/rvpay-go/grpc/go/integrationsgrpc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newTestService(t *testing.T, repo *mocks.MockIntegrationsRepo) *Impl {
	t.Helper()
	return NewIntegrationService(repo, zerolog.Nop())
}

func newMockRepoWithQuerier(t *testing.T, ctrl *gomock.Controller, querier *sqlcmocks.MockQuerier) *mocks.MockIntegrationsRepo {
	t.Helper()

	repo := mocks.NewMockIntegrationsRepo(ctrl)
	repo.EXPECT().Do().Return(querier).AnyTimes()
	return repo
}

func TestCreateIntegration(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	integrationID := uuid.New()

	tests := []struct {
		name    string
		req     *integrationsgrpc.CreateIntegrationRequest
		setup   func(*sqlcmocks.MockQuerier)
		wantErr bool
		want    *integrationsgrpc.Integration
	}{
		{
			name:    "missing request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "missing token expires at",
			req: &integrationsgrpc.CreateIntegrationRequest{
				Provider:     "highlevel",
				LocationId:   "loc_123",
				AccessToken:  "access",
				RefreshToken: "refresh",
			},
			wantErr: true,
		},
		{
			name: "repository error",
			req: &integrationsgrpc.CreateIntegrationRequest{
				Provider:       "highlevel",
				LocationId:     "loc_123",
				AccessToken:    "access",
				RefreshToken:   "refresh",
				TokenExpiresAt: timestamppb.New(now),
			},
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().CreateIntegration(gomock.Any(), gomock.Any()).
					Return(sqlc.Integration{}, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "success",
			req: &integrationsgrpc.CreateIntegrationRequest{
				Provider:       "highlevel",
				LocationId:     "loc_123",
				AccessToken:    "access",
				RefreshToken:   "refresh",
				TokenExpiresAt: timestamppb.New(now),
			},
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().CreateIntegration(gomock.Any(), gomock.Any()).
					Return(sqlc.Integration{
						ID:             integrationID,
						Provider:       "highlevel",
						LocationID:     "loc_123",
						AccessToken:    "access",
						RefreshToken:   "refresh",
						TokenExpiresAt: now,
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil)
			},
			want: &integrationsgrpc.Integration{
				Id:             integrationID.String(),
				Provider:       "highlevel",
				LocationId:     "loc_123",
				AccessToken:    "access",
				RefreshToken:   "refresh",
				TokenExpiresAt: timestamppb.New(now),
				CreatedAt:      timestamppb.New(now),
				UpdatedAt:      timestamppb.New(now),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			querier := sqlcmocks.NewMockQuerier(ctrl)
			if tt.setup != nil {
				tt.setup(querier)
			}

			repo := newMockRepoWithQuerier(t, ctrl, querier)
			service := newTestService(t, repo)
			got, err := service.CreateIntegration(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.GetIntegration().GetId() != tt.want.GetId() {
				t.Fatalf("id = %s, want %s", got.GetIntegration().GetId(), tt.want.GetId())
			}
			if got.GetIntegration().GetProvider() != tt.want.GetProvider() {
				t.Fatalf("provider = %s, want %s", got.GetIntegration().GetProvider(), tt.want.GetProvider())
			}
			if got.GetIntegration().GetLocationId() != tt.want.GetLocationId() {
				t.Fatalf("location_id = %s, want %s", got.GetIntegration().GetLocationId(), tt.want.GetLocationId())
			}
		})
	}
}

func TestCreateIntegrationStatusCodes(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	querier := sqlcmocks.NewMockQuerier(ctrl)
	querier.EXPECT().CreateIntegration(gomock.Any(), gomock.Any()).
		Return(sqlc.Integration{}, errors.New("db error"))

	repo := newMockRepoWithQuerier(t, ctrl, querier)
	service := newTestService(t, repo)
	_, err := service.CreateIntegration(context.Background(), &integrationsgrpc.CreateIntegrationRequest{
		Provider:       "highlevel",
		LocationId:     "loc_123",
		AccessToken:    "access",
		RefreshToken:   "refresh",
		TokenExpiresAt: timestamppb.New(time.Now()),
	})

	if got := status.Code(err); got != codes.Internal {
		t.Fatalf("status code = %s, want %s", got, codes.Internal)
	}
}

func TestGetIntegration(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	integrationID := uuid.New()

	tests := []struct {
		name    string
		req     *integrationsgrpc.GetIntegrationRequest
		setup   func(*sqlcmocks.MockQuerier)
		wantErr bool
		want    *integrationsgrpc.Integration
	}{
		{
			name:    "missing request",
			req:     nil,
			wantErr: true,
		},
		{
			name:    "invalid id",
			req:     &integrationsgrpc.GetIntegrationRequest{Id: "invalid"},
			wantErr: true,
		},
		{
			name: "not found",
			req:  &integrationsgrpc.GetIntegrationRequest{Id: integrationID.String()},
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().GetIntegrationByID(gomock.Any(), integrationID).
					Return(sqlc.Integration{}, pgx.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "repository error",
			req:  &integrationsgrpc.GetIntegrationRequest{Id: integrationID.String()},
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().GetIntegrationByID(gomock.Any(), integrationID).
					Return(sqlc.Integration{}, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "success",
			req:  &integrationsgrpc.GetIntegrationRequest{Id: integrationID.String()},
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().GetIntegrationByID(gomock.Any(), integrationID).
					Return(sqlc.Integration{
						ID:             integrationID,
						Provider:       "highlevel",
						LocationID:     "loc_123",
						AccessToken:    "access",
						RefreshToken:   "refresh",
						TokenExpiresAt: now,
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil)
			},
			want: &integrationsgrpc.Integration{
				Id:             integrationID.String(),
				Provider:       "highlevel",
				LocationId:     "loc_123",
				AccessToken:    "access",
				RefreshToken:   "refresh",
				TokenExpiresAt: timestamppb.New(now),
				CreatedAt:      timestamppb.New(now),
				UpdatedAt:      timestamppb.New(now),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			querier := sqlcmocks.NewMockQuerier(ctrl)
			if tt.setup != nil {
				tt.setup(querier)
			}

			repo := newMockRepoWithQuerier(t, ctrl, querier)
			service := newTestService(t, repo)
			got, err := service.GetIntegration(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.GetIntegration().GetId() != tt.want.GetId() {
				t.Fatalf("id = %s, want %s", got.GetIntegration().GetId(), tt.want.GetId())
			}
			if got.GetIntegration().GetProvider() != tt.want.GetProvider() {
				t.Fatalf("provider = %s, want %s", got.GetIntegration().GetProvider(), tt.want.GetProvider())
			}
		})
	}
}

func TestGetIntegrationStatusCodes(t *testing.T) {
	t.Parallel()

	integrationID := uuid.New()

	tests := []struct {
		name  string
		setup func(*sqlcmocks.MockQuerier)
		code  codes.Code
	}{
		{
			name: "not found",
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().GetIntegrationByID(gomock.Any(), integrationID).
					Return(sqlc.Integration{}, pgx.ErrNoRows)
			},
			code: codes.NotFound,
		},
		{
			name: "internal error",
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().GetIntegrationByID(gomock.Any(), integrationID).
					Return(sqlc.Integration{}, errors.New("db error"))
			},
			code: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			querier := sqlcmocks.NewMockQuerier(ctrl)
			tt.setup(querier)

			repo := newMockRepoWithQuerier(t, ctrl, querier)
			service := newTestService(t, repo)
			_, err := service.GetIntegration(context.Background(), &integrationsgrpc.GetIntegrationRequest{Id: integrationID.String()})

			if got := status.Code(err); got != tt.code {
				t.Fatalf("status code = %s, want %s", got, tt.code)
			}
		})
	}
}

func TestDeleteIntegration(t *testing.T) {
	t.Parallel()

	integrationID := uuid.New()

	tests := []struct {
		name    string
		req     *integrationsgrpc.DeleteIntegrationRequest
		setup   func(*sqlcmocks.MockQuerier)
		wantErr bool
	}{
		{
			name:    "missing request",
			req:     nil,
			wantErr: true,
		},
		{
			name:    "invalid id",
			req:     &integrationsgrpc.DeleteIntegrationRequest{Id: "invalid"},
			wantErr: true,
		},
		{
			name: "not found",
			req:  &integrationsgrpc.DeleteIntegrationRequest{Id: integrationID.String()},
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().DeleteIntegration(gomock.Any(), integrationID).Return(int64(0), nil)
			},
			wantErr: true,
		},
		{
			name: "repository error",
			req:  &integrationsgrpc.DeleteIntegrationRequest{Id: integrationID.String()},
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().DeleteIntegration(gomock.Any(), integrationID).Return(int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "success",
			req:  &integrationsgrpc.DeleteIntegrationRequest{Id: integrationID.String()},
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().DeleteIntegration(gomock.Any(), integrationID).Return(int64(1), nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			querier := sqlcmocks.NewMockQuerier(ctrl)
			if tt.setup != nil {
				tt.setup(querier)
			}

			repo := newMockRepoWithQuerier(t, ctrl, querier)
			service := newTestService(t, repo)
			got, err := service.DeleteIntegration(context.Background(), tt.req)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.GetId() != integrationID.String() {
				t.Fatalf("id = %s, want %s", got.GetId(), integrationID.String())
			}
		})
	}
}

func TestDeleteIntegrationStatusCodes(t *testing.T) {
	t.Parallel()

	integrationID := uuid.New()

	tests := []struct {
		name  string
		setup func(*sqlcmocks.MockQuerier)
		code  codes.Code
	}{
		{
			name: "not found",
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().DeleteIntegration(gomock.Any(), integrationID).Return(int64(0), nil)
			},
			code: codes.NotFound,
		},
		{
			name: "internal error",
			setup: func(q *sqlcmocks.MockQuerier) {
				q.EXPECT().DeleteIntegration(gomock.Any(), integrationID).Return(int64(0), errors.New("db error"))
			},
			code: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			querier := sqlcmocks.NewMockQuerier(ctrl)
			tt.setup(querier)

			repo := newMockRepoWithQuerier(t, ctrl, querier)
			service := newTestService(t, repo)
			_, err := service.DeleteIntegration(context.Background(), &integrationsgrpc.DeleteIntegrationRequest{Id: integrationID.String()})

			if got := status.Code(err); got != tt.code {
				t.Fatalf("status code = %s, want %s", got, tt.code)
			}
		})
	}
}

func TestGrpcIntegrationRequestToSqlc(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	if _, err := grpcIntegrationRequestToSqlc(nil); err == nil {
		t.Fatal("nil request did not return an error")
	}

	if _, err := grpcIntegrationRequestToSqlc(&integrationsgrpc.CreateIntegrationRequest{
		Provider: "highlevel",
	}); err == nil {
		t.Fatal("missing token expires at did not return an error")
	}

	params, err := grpcIntegrationRequestToSqlc(&integrationsgrpc.CreateIntegrationRequest{
		Provider:       "highlevel",
		LocationId:     "loc_123",
		AccessToken:    "access",
		RefreshToken:   "refresh",
		TokenExpiresAt: timestamppb.New(now),
	})
	if err != nil {
		t.Fatalf("valid request returned error: %v", err)
	}

	if params.Provider != "highlevel" {
		t.Fatalf("provider = %s, want highlevel", params.Provider)
	}
	if params.LocationID != "loc_123" {
		t.Fatalf("location_id = %s, want loc_123", params.LocationID)
	}
	if params.AccessToken != "access" {
		t.Fatalf("access_token = %s, want access", params.AccessToken)
	}
	if params.RefreshToken != "refresh" {
		t.Fatalf("refresh_token = %s, want refresh", params.RefreshToken)
	}
	if !params.TokenExpiresAt.Equal(now) {
		t.Fatalf("token_expires_at = %v, want %v", params.TokenExpiresAt, now)
	}
}

func TestSqlcIntegrationToGrpc(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	integrationID := uuid.New()

	got := sqlcIntegrationToGrpc(sqlc.Integration{
		ID:             integrationID,
		Provider:       "highlevel",
		LocationID:     "loc_123",
		AccessToken:    "access",
		RefreshToken:   "refresh",
		TokenExpiresAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	})

	if got.GetId() != integrationID.String() {
		t.Fatalf("id = %s, want %s", got.GetId(), integrationID.String())
	}
	if got.GetProvider() != "highlevel" {
		t.Fatalf("provider = %s, want highlevel", got.GetProvider())
	}
	if got.GetLocationId() != "loc_123" {
		t.Fatalf("location_id = %s, want loc_123", got.GetLocationId())
	}
	if got.GetAccessToken() != "access" {
		t.Fatalf("access_token = %s, want access", got.GetAccessToken())
	}
	if got.GetRefreshToken() != "refresh" {
		t.Fatalf("refresh_token = %s, want refresh", got.GetRefreshToken())
	}
	if !got.GetTokenExpiresAt().AsTime().Equal(now) {
		t.Fatalf("token_expires_at = %v, want %v", got.GetTokenExpiresAt().AsTime(), now)
	}
	if !got.GetCreatedAt().AsTime().Equal(now) {
		t.Fatalf("created_at = %v, want %v", got.GetCreatedAt().AsTime(), now)
	}
	if !got.GetUpdatedAt().AsTime().Equal(now) {
		t.Fatalf("updated_at = %v, want %v", got.GetUpdatedAt().AsTime(), now)
	}
}