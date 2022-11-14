package conns

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type clientInitFunc[T any] func() T

type lazyClient[T any] struct {
	initf clientInitFunc[T]

	once   sync.Once
	client T
}

func (l *lazyClient[T]) init(config *aws.Config, f clientInitFunc[T]) {
	l.initf = f
}

func (l *lazyClient[T]) Client() T {
	l.once.Do(func() {
		l.client = l.initf()
	})
	return l.client
}
