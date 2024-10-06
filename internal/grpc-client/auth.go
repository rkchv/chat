package grpc_client

import (
	"context"

	"github.com/rkchv/auth/pkg/user_v1"
	"google.golang.org/grpc"
)

type Auth struct {
	client user_v1.UserV1Client
}

func NewAuth(conn *grpc.ClientConn) *Auth {
	return &Auth{client: user_v1.NewUserV1Client(conn)}
}

func (a *Auth) CanDelete(ctx context.Context, userID int64) (bool, error) {
	resp, err := a.client.CanDelete(ctx, &user_v1.RightsRequest{UserID: userID})
	if err != nil {
		return false, err
	}

	return resp.Can, nil
}
