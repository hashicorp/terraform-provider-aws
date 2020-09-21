package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
)

// ResolverQueryLogConfigByID returns the query logging configuration corresponding to the specified ID.
// Returns nil if no configuration is found.
func ResolverQueryLogConfigByID(conn *route53resolver.Route53Resolver, queryLogConfigID string) (*route53resolver.ResolverQueryLogConfig, error) {
	input := &route53resolver.GetResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(queryLogConfigID),
	}

	output, err := conn.GetResolverQueryLogConfig(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.ResolverQueryLogConfig, nil
}
