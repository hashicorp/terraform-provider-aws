package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/finder"
)

const (
	resolverQueryLogConfigStatusNotFound = "NotFound"
	resolverQueryLogConfigStatusUnknown  = "Unknown"
)

// QueryLogConfigStatus fetches the QueryLogConfig and its Status
func QueryLogConfigStatus(conn *route53resolver.Route53Resolver, queryLogConfigID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		queryLogConfig, err := finder.ResolverQueryLogConfigByID(conn, queryLogConfigID)

		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return nil, resolverQueryLogConfigStatusNotFound, nil
		}

		if err != nil {
			return nil, resolverQueryLogConfigStatusUnknown, err
		}

		if queryLogConfig == nil {
			return nil, resolverQueryLogConfigStatusNotFound, nil
		}

		return queryLogConfig, aws.StringValue(queryLogConfig.Status), nil
	}
}
