// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package simpledb

import (
	"context"
	"fmt"
	"net"
	"net/url"

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

	if endpoint := c.ResolveEndpoint(ctx, servicePackageName); endpoint != "" {
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
