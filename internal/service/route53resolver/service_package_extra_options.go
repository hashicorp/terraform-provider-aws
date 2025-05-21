package route53resolver

import (
	"context"
	"errors"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	retry_sdkv2 "github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/route53resolver"
	smithy "github.com/aws/smithy-go"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// service_package_gen.go is a generated file but the behavior can be extended in another file by defining this
// function which NewClient will call. Use this to customize error retries.
func (p *servicePackage) withExtraOptions(
	_ context.Context,
	config map[string]any,
) []func(*route53resolver.Options) {
	retryer := mustFindRetryer(config)
	return []func(*route53resolver.Options){func(o *route53resolver.Options) {
		o.Retryer = conns.AddIsErrorRetryables(retryer, doNotRetryLimitExceededException())
	}}

}

func mustFindRetryer(config map[string]any) aws_sdkv2.RetryerV2 {
	if cfg, ok := config["aws_sdkv2_config"]; ok {
		if cfgp, ok := cfg.(*aws_sdkv2.Config); ok {
			if r, ok := cfgp.Retryer().(aws_sdkv2.RetryerV2); ok {
				return r
			}
		}
	}
	return nil
}

func doNotRetryLimitExceededException() retry_sdkv2.IsErrorRetryable {
	return retry_sdkv2.IsErrorRetryableFunc(func(err error) aws_sdkv2.Ternary {
		var smithyErr smithy.APIError
		if ok := errors.As(err, &smithyErr); ok {
			if smithyErr.ErrorCode() == "LimitExceededException" {
				return aws_sdkv2.FalseTernary
			}
		}
		return aws_sdkv2.UnknownTernary // Delegate to the configured Retryer.
	})
}
