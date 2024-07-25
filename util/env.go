package util

import "context"

const (
	Test = "test"
	Prod = "prod"
)

func GetEnv(ctx context.Context) string {
	env := ctx.Value("env")
	if Env, ok := env.(string); ok {
		return Env
	}
	return ""
}

func SetEnv(ctx context.Context, env string) context.Context {
	ctx = context.WithValue(ctx, "env", env)
	return ctx
}

func IsTestEnv(ctx context.Context) bool {
	return GetEnv(ctx) == Test
}

func IsProdEnv(ctx context.Context) bool {
	return GetEnv(ctx) == Prod
}
