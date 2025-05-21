package process

import "context"

type Runner interface {
	Run(ctx context.Context, path, target string) error
}
