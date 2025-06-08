package biz

import (
	"context"
	plog "demo/pkg/log"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport"
	jwtV5 "github.com/golang-jwt/jwt/v5"

	"demo/internal/conf"

	pb "demo/api/v1"
)

type Auth struct {
	jwtConf *conf.Middleware_Token
}

func NewAuth(c *conf.Server) *Auth {
	return &Auth{
		jwtConf: c.Middleware.Token,
	}
}

type MyCustomClaims struct {
	UserId uint `json:"userId"`
	jwtV5.RegisteredClaims
}

const (
	RoleCommon        uint8 = 1
	RoleOperator      uint8 = 2
	RolePlatformAdmin uint8 = 3
	RoleSuperAdmin    uint8 = 4
)

type User struct {
	RoleId uint `json:"roleId"`
}

func (a *Auth) CurrentUserId(ctx context.Context) (uint, error) {
	if claims, ok := jwt.FromContext(ctx); ok {
		if m, ok := claims.(*MyCustomClaims); ok {
			return m.UserId, nil
		}
	}
	return 0, errors.New(401, "Unauthorized", "jwt claim missing")
}

func (a *Auth) CurrentUser(ctx context.Context) (*User, error) {
	userId, err := a.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	plog.Infof(ctx, "user id is: %d", userId)
	// todo query RoleId by userId
	UserMap := map[uint]uint{
		496: uint(4),
	}
	users := make([]*User, 0)
	users = append(users, &User{
		RoleId: UserMap[userId],
	})
	if len(users) == 0 {
		return nil, errors.New(401, "Unauthorized", "Not exist User")
	}
	return users[0], nil
}

func (a *Auth) Middleware() middleware.Middleware {
	return selector.Server(
		a.jwtMiddleware(),
		a.roleMiddleware(),
	).Match(newWhiteListMatcher()).Build()
}

func (a *Auth) jwtMiddleware() middleware.Middleware {
	return jwt.Server(
		func(token *jwtV5.Token) (interface{}, error) {
			return []byte(a.jwtConf.GetJwt()), nil
		},
		jwt.WithSigningMethod(jwtV5.SigningMethodHS256),
		jwt.WithClaims(
			func() jwtV5.Claims {
				return &MyCustomClaims{}
			}))
}

func (a *Auth) roleMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			user, err := a.CurrentUser(ctx)
			if err != nil {
				return nil, err
			}
			// check角色权限
			if tr, ok := transport.FromServerContext(ctx); ok {
				op := tr.Operation()
				opMinRole := getOpMinRole()
				if minRole, ok := opMinRole[op]; ok {
					roleId := uint8(user.RoleId)
					// check role
					if roleId < minRole {
						return nil, errors.New(403, "Forbidden", "No Permission")
					}
				}
			}
			return handler(ctx, req)
		}
	}
}

func newWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]struct{})
	// white list that don't auth
	//whiteList[pb.OperationGreeterSvcListGreeter] = struct{}{}

	var whiteListPattern []string
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		for _, pat := range whiteListPattern {
			if strings.Contains(operation, pat) {
				return false
			}
		}
		return true
	}
}

func getOpMinRole() map[string]uint8 {
	return map[string]uint8{
		// set the role for operation
		pb.OperationGreeterSvcListGreeter: RolePlatformAdmin,
	}
}
