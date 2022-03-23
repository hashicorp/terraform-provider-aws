package logging

import (
	"context"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

// DataSourceContext injects the data source type into logger contexts.
func DataSourceContext(ctx context.Context, dataSource string) context.Context {
	ctx = tfsdklog.With(ctx, KeyDataSourceType, dataSource)
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemProto, KeyDataSourceType, dataSource)
	ctx = tflog.With(ctx, KeyDataSourceType, dataSource)

	return ctx
}

// InitContext creates SDK and provider logger contexts.
func InitContext(ctx context.Context, sdkOpts tfsdklog.Options, providerOpts tflog.Options) context.Context {
	ctx = tfsdklog.NewRootSDKLogger(ctx, append(tfsdklog.Options{
		tfsdklog.WithLevelFromEnv(EnvTfLogSdk),
	}, sdkOpts...)...)
	ctx = tfsdklog.NewSubsystem(ctx, SubsystemProto, append(tfsdklog.Options{
		tfsdklog.WithLevelFromEnv(EnvTfLogSdkProto),
	}, sdkOpts...)...)
	ctx = tfsdklog.NewRootProviderLogger(ctx, providerOpts...)

	return ctx
}

// ProtocolVersionContext injects the protocol version into logger contexts.
func ProtocolVersionContext(ctx context.Context, protocolVersion string) context.Context {
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemProto, KeyProtocolVersion, protocolVersion)

	return ctx
}

// ProviderAddressContext injects the provider address into logger contexts.
func ProviderAddressContext(ctx context.Context, providerAddress string) context.Context {
	ctx = tfsdklog.With(ctx, KeyProviderAddress, providerAddress)
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemProto, KeyProviderAddress, providerAddress)
	ctx = tflog.With(ctx, KeyProviderAddress, providerAddress)

	return ctx
}

// RequestIdContext injects a unique request ID into logger contexts.
func RequestIdContext(ctx context.Context) context.Context {
	reqID, err := uuid.GenerateUUID()

	if err != nil {
		reqID = "unable to assign request ID: " + err.Error()
	}

	ctx = tfsdklog.With(ctx, KeyRequestID, reqID)
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemProto, KeyRequestID, reqID)
	ctx = tflog.With(ctx, KeyRequestID, reqID)

	return ctx
}

// ResourceContext injects the resource type into logger contexts.
func ResourceContext(ctx context.Context, resource string) context.Context {
	ctx = tfsdklog.With(ctx, KeyResourceType, resource)
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemProto, KeyResourceType, resource)
	ctx = tflog.With(ctx, KeyResourceType, resource)

	return ctx
}

// RpcContext injects the RPC name into logger contexts.
func RpcContext(ctx context.Context, rpc string) context.Context {
	ctx = tfsdklog.With(ctx, KeyRPC, rpc)
	ctx = tfsdklog.SubsystemWith(ctx, SubsystemProto, KeyRPC, rpc)
	ctx = tflog.With(ctx, KeyRPC, rpc)

	return ctx
}
