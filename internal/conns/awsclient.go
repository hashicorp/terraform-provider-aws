package conns

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// PartitionHostname returns a hostname with the provider domain suffix for the partition
// e.g. PREFIX.amazonaws.com
// The prefix should not contain a trailing period.
func (client *AWSClient) PartitionHostname(prefix string) string {
	return fmt.Sprintf("%s.%s", prefix, client.DNSSuffix)
}

// RegionalHostname returns a hostname with the provider domain suffix for the region and partition
// e.g. PREFIX.us-west-2.amazonaws.com
// The prefix should not contain a trailing period.
func (client *AWSClient) RegionalHostname(prefix string) string {
	return fmt.Sprintf("%s.%s.%s", prefix, client.Region, client.DNSSuffix)
}

type clientInitFunc[T any, OT any] func(aws.Config, ...func(OT)) T

type lazyClient[T any, OT any] struct {
	config *aws.Config
	initf  clientInitFunc[T, OT]

	clientOnce sync.Once
	client     T
}

func (l *lazyClient[T, OT]) init(config *aws.Config, f clientInitFunc[T, OT]) {
	l.config = config
	l.initf = f
}

func (l *lazyClient[T, OT]) Client() T {
	l.clientOnce.Do(func() {
		l.client = l.initf(*l.config)
	})
	return l.client
}

func (client *AWSClient) SSMClient() *ssm.Client {
	return client.ssmClient.Client()
}
