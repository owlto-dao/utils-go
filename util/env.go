package util

import "context"

func GetEnv(ctx context.Context) string {
	logId := ctx.Value("env")
	if Env, ok := logId.(string); ok {
		return Env
	}
	return ""
}
