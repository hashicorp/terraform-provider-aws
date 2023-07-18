// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

import (
	"context"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

// DataSourceContext injects the data source type into logger contexts.
func DataSourceContext(ctx context.Context, dataSource string) context.Context {
	ctx = tfsdklog.SetField(ctx, KeyDataSourceType, dataSource)
	ctx = tfsdklog.SubsystemSetField(ctx, SubsystemProto, KeyDataSourceType, dataSource)
	ctx = tflog.SetField(ctx, KeyDataSourceType, dataSource)

	return ctx
}

// InitContext creates SDK and provider logger contexts.
func InitContext(ctx context.Context, sdkOpts tfsdklog.Options, providerOpts tflog.Options) context.Context {
	ctx = tfsdklog.NewRootSDKLogger(ctx, append(tfsdklog.Options{
		tfsdklog.WithLevelFromEnv(EnvTfLogSdk),
	}, sdkOpts...)...)
	ctx = ProtoSubsystemContext(ctx, sdkOpts)
	ctx = tfsdklog.NewRootProviderLogger(ctx, providerOpts...)

	return ctx
}

// ProtoSubsystemContext adds the proto subsystem to the SDK logger context.
func ProtoSubsystemContext(ctx context.Context, sdkOpts tfsdklog.Options) context.Context {
	ctx = tfsdklog.NewSubsystem(ctx, SubsystemProto, append(tfsdklog.Options{
		// All calls are through the Protocol* helper functions
		tfsdklog.WithAdditionalLocationOffset(1),
		tfsdklog.WithLevelFromEnv(EnvTfLogSdkProto),
	}, sdkOpts...)...)

	return ctx
}

// ProtocolVersionContext injects the protocol version into logger contexts.
func ProtocolVersionContext(ctx context.Context, protocolVersion string) context.Context {
	ctx = tfsdklog.SubsystemSetField(ctx, SubsystemProto, KeyProtocolVersion, protocolVersion)

	return ctx
}

// ProviderAddressContext injects the provider address into logger contexts.
func ProviderAddressContext(ctx context.Context, providerAddress string) context.Context {
	ctx = tfsdklog.SetField(ctx, KeyProviderAddress, providerAddress)
	ctx = tfsdklog.SubsystemSetField(ctx, SubsystemProto, KeyProviderAddress, providerAddress)
	ctx = tflog.SetField(ctx, KeyProviderAddress, providerAddress)

	return ctx
}

// RequestIdContext injects a unique request ID into logger contexts.
func RequestIdContext(ctx context.Context) context.Context {
	reqID, err := uuid.GenerateUUID()

	if err != nil {
		reqID = "unable to assign request ID: " + err.Error()
	}

	ctx = tfsdklog.SetField(ctx, KeyRequestID, reqID)
	ctx = tfsdklog.SubsystemSetField(ctx, SubsystemProto, KeyRequestID, reqID)
	ctx = tflog.SetField(ctx, KeyRequestID, reqID)

	return ctx
}

// ResourceContext injects the resource type into logger contexts.
func ResourceContext(ctx context.Context, resource string) context.Context {
	ctx = tfsdklog.SetField(ctx, KeyResourceType, resource)
	ctx = tfsdklog.SubsystemSetField(ctx, SubsystemProto, KeyResourceType, resource)
	ctx = tflog.SetField(ctx, KeyResourceType, resource)

	return ctx
}

// RpcContext injects the RPC name into logger contexts.
func RpcContext(ctx context.Context, rpc string) context.Context {
	ctx = tfsdklog.SetField(ctx, KeyRPC, rpc)
	ctx = tfsdklog.SubsystemSetField(ctx, SubsystemProto, KeyRPC, rpc)
	ctx = tflog.SetField(ctx, KeyRPC, rpc)

	return ctx
}
