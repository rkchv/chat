package interceptors

import (
	"context"
	"slices"
	"strings"

	"github.com/rkchv/auth/pkg/user_v1/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authHeader = "Authorization"
const authPrefix = "Bearer "

var secureMethodsMap map[string]struct{}
var secretKey string

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

// NewStreamAccessInterceptor для заданных методов проверяет наличие access-токена и наличие соответствующего scope в нем
// так же при успешной проверке записывает данные из токена в контекст
func NewStreamAccessInterceptor(secureMethods []string, jwtSecretKey string) grpc.StreamServerInterceptor {
	secretKey = jwtSecretKey

	if len(secureMethods) > 0 {
		secureMethodsMap = make(map[string]struct{}, len(secureMethods))
		for _, m := range secureMethods {
			secureMethodsMap[m] = struct{}{}
		}
	}

	return accessInterceptor
}

func accessInterceptor(srv any, ss grpc.ServerStream, i *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	//смотрим требует ли метод проверки доступа
	//если да, то смотрим наличие метода в scope разделе токена
	ctx := ss.Context()
	if _, needCheck := secureMethodsMap[i.FullMethod]; needCheck {
		meta, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		token := meta.Get(authHeader)
		if len(token) == 0 {
			return status.Error(codes.Unauthenticated, "token is not provided")
		}

		if !strings.HasPrefix(token[0], authPrefix) {
			return status.Error(codes.Unauthenticated, "invalid auth header format")
		}

		accessToken := strings.TrimPrefix(token[0], authPrefix)
		user, err := auth.ParseToken(accessToken, []byte(secretKey))
		if err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}

		if !slices.Contains(user.Scope, i.FullMethod) {
			return status.Error(codes.PermissionDenied, "нет доступа")
		}

		ctx = auth.AddUserToContext(ctx, user)
	}

	return handler(srv, &serverStream{ServerStream: ss, ctx: ctx})
}
