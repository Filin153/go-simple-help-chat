package oneass

import "context"

type OneAssClientLogin struct{}

func NewOneAssClientLogin() *OneAssClientLogin {
	return &OneAssClientLogin{}
}

func (o *OneAssClientLogin) Login(ctx context.Context, args ...any) (userUUID string, userInfo map[string]any, err error) {
	return "test-uuid", map[string]any{
		"test": "data",
	}, nil
}
