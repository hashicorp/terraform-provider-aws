package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a QueryLogConfig to return CREATED
	QueryLogConfigCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a QueryLogConfig to be deleted
	QueryLogConfigDeletedTimeout = 5 * time.Minute
)

// QueryLogConfigCreated waits for a QueryLogConfig to return CREATED
func QueryLogConfigCreated(conn *route53resolver.Route53Resolver, queryLogConfigID string) (*route53resolver.ResolverQueryLogConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigStatusCreating},
		Target:  []string{route53resolver.ResolverQueryLogConfigStatusCreated},
		Refresh: QueryLogConfigStatus(conn, queryLogConfigID),
		Timeout: QueryLogConfigCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfig); ok {
		return v, err
	}

	return nil, err
}

// QueryLogConfigCreated waits for a QueryLogConfig to be deleted
func QueryLogConfigDeleted(conn *route53resolver.Route53Resolver, queryLogConfigID string) (*route53resolver.ResolverQueryLogConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigStatusDeleting},
		Target:  []string{},
		Refresh: QueryLogConfigStatus(conn, queryLogConfigID),
		Timeout: QueryLogConfigDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfig); ok {
		return v, err
	}

	return nil, err
}
