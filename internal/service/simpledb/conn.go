// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package simpledb

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/simpledb"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	conn *simpledb.SimpleDB
)

// Adapted from
// internal/conns/awsclient_gen.go: func (c *AWSClient) SimpleDBConn(ctx context.Context) *simpledb_sdkv1.SimpleDB
// internal/conns/awsclient.go: func conn[T any](ctx context.Context, c *AWSClient, servicePackageName string, extra map[string]any) (T, error)
// internal/service/simpledb/service_package_gen.go: func (p *servicePackage) NewConn(ctx context.Context, config map[string]any) (*simpledb_sdkv1.SimpleDB, error)

func simpleDBConn(ctx context.Context, c *conns.AWSClient) *simpledb.SimpleDB { // nosemgrep:ci.simpledb-in-func-name
	const servicePackageName = names.SimpleDB
	ctx = tflog.SetField(ctx, "tf_aws.service_package", servicePackageName)

	const mutexKey = "simpledb-conn"
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	if conn != nil {
		return conn
	}

	cfg := aws.Config{}

	if endpoint := resolveEndpoint(ctx, c); endpoint != "" {
		tflog.Debug(ctx, "setting endpoint", map[string]any{
			"tf_aws.endpoint": endpoint,
		})
		cfg.Endpoint = aws.String(endpoint)
	} else {
		cfg.EndpointResolver = newEndpointResolverSDKv1(ctx)
	}

	conn = simpledb.New(c.AwsSession(ctx).Copy(&cfg))

	return conn
}

// Adapted from
//	internal/conns/awsclient_resolveendpoint_gen.go: func (c *AWSClient) ResolveEndpoint(ctx context.Context, servicePackageName string) string

func resolveEndpoint(ctx context.Context, c *conns.AWSClient) string {
	endpoint := c.Endpoints(ctx)[names.SimpleDB]
	if endpoint != "" {
		return endpoint
	}

	//endpoint = aws.StringValue(c.AwsConfig(ctx).BaseEndpoint)
	svc := os.Getenv("AWS_ENDPOINT_URL_SIMPLEDB")
	if svc != "" {
		return svc
	}

	if base := os.Getenv("AWS_ENDPOINT_URL"); base != "" {
		return base
	}

	endpoint, found, err := resolveServiceBaseEndpoint(ctx, "SimpleDB", c.AwsConfig(ctx).ConfigSources)
	if found && err == nil {
		return endpoint
	}

	return endpoint
}

// Copied from internal/conns/awsclient.go.

// serviceBaseEndpointProvider is needed to search for all providers
// that provide a configured service endpoint
type serviceBaseEndpointProvider interface {
	GetServiceBaseEndpoint(ctx context.Context, sdkID string) (string, bool, error)
}

// resolveServiceBaseEndpoint is used to retrieve service endpoints from configured sources
// while allowing for configured endpoints to be disabled
func resolveServiceBaseEndpoint(ctx context.Context, sdkID string, configs []any) (value string, found bool, err error) {
	if val, found, _ := config.GetIgnoreConfiguredEndpoints(ctx, configs); found && val {
		return "", false, nil
	}

	for _, cs := range configs {
		if p, ok := cs.(serviceBaseEndpointProvider); ok {
			value, found, err = p.GetServiceBaseEndpoint(ctx, sdkID)
			if err != nil || found {
				break
			}
		}
	}
	return
}

// Copied from internal/service/simpledb/service_endpoint_resolver_gen.go.

var _ endpoints.Resolver = resolverSDKv1{}

type resolverSDKv1 struct {
	ctx context.Context //nolint:containedctx // Was in generated code
}

func newEndpointResolverSDKv1(ctx context.Context) resolverSDKv1 {
	return resolverSDKv1{
		ctx: ctx,
	}
}

func (r resolverSDKv1) EndpointFor(service, region string, opts ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
	ctx := r.ctx

	var opt endpoints.Options
	opt.Set(opts...)

	useFIPS := opt.UseFIPSEndpoint == endpoints.FIPSEndpointStateEnabled

	defaultResolver := endpoints.DefaultResolver()

	if useFIPS {
		ctx = tflog.SetField(ctx, "tf_aws.use_fips", useFIPS)

		endpoint, err := defaultResolver.EndpointFor(service, region, opts...)
		if err != nil {
			return endpoint, err
		}

		tflog.Debug(ctx, "endpoint resolved", map[string]any{
			"tf_aws.endpoint": endpoint.URL,
		})

		var endpointURL *url.URL
		endpointURL, err = url.Parse(endpoint.URL)
		if err != nil {
			return endpoint, err
		}

		hostname := endpointURL.Hostname()
		_, err = net.LookupHost(hostname)
		if err != nil {
			if dnsErr, ok := errs.As[*net.DNSError](err); ok && dnsErr.IsNotFound {
				tflog.Debug(ctx, "default endpoint host not found, disabling FIPS", map[string]any{
					"tf_aws.hostname": hostname,
				})
				opts = append(opts, func(o *endpoints.Options) {
					o.UseFIPSEndpoint = endpoints.FIPSEndpointStateDisabled
				})
			} else {
				err := fmt.Errorf("looking up simpledb endpoint %q: %s", hostname, err)
				return endpoints.ResolvedEndpoint{}, err
			}
		} else {
			return endpoint, err
		}
	}

	return defaultResolver.EndpointFor(service, region, opts...)
}
