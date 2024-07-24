package util

import "context"

func GetEnv(ctx context.Context) string {
	env := ctx.Value("env")
	if Env, ok := env.(string); ok {
		return Env
	}
	return ""
}
