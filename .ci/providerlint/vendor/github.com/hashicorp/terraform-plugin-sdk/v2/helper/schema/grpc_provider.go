// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/hashicorp/go-cty/cty"
	ctyconvert "github.com/hashicorp/go-cty/cty/convert"
	"github.com/hashicorp/go-cty/cty/msgpack"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/configschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/configs/hcl2shim"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/plans/objchange"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/plugin/convert"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	newExtraKey = "_new_extra_shim"
)

// Verify provider server interface implementation.
var _ tfprotov5.ProviderServer = (*GRPCProviderServer)(nil)

func NewGRPCProviderServer(p *Provider) *GRPCProviderServer {
	return &GRPCProviderServer{
		provider: p,
		stopCh:   make(chan struct{}),
	}
}

// GRPCProviderServer handles the server, or plugin side of the rpc connection.
type GRPCProviderServer struct {
	provider *Provider
	stopCh   chan struct{}
	stopMu   sync.Mutex
}

// mergeStop is called in a goroutine and waits for the global stop signal
// and propagates cancellation to the passed in ctx/cancel func. The ctx is
// also passed to this function and waited upon so no goroutine leak is caused.
func mergeStop(ctx context.Context, cancel context.CancelFunc, stopCh chan struct{}) {
	select {
	case <-ctx.Done():
		return
	case <-stopCh:
		cancel()
	}
}

// StopContext derives a new context from the passed in grpc context.
// It creates a goroutine to wait for the server stop and propagates
// cancellation to the derived grpc context.
func (s *GRPCProviderServer) StopContext(ctx context.Context) context.Context {
	ctx = logging.InitContext(ctx)
	s.stopMu.Lock()
	defer s.stopMu.Unlock()

	stoppable, cancel := context.WithCancel(ctx)
	go mergeStop(stoppable, cancel, s.stopCh)
	return stoppable
}

func (s *GRPCProviderServer) serverCapabilities() *tfprotov5.ServerCapabilities {
	return &tfprotov5.ServerCapabilities{
		GetProviderSchemaOptional: true,
	}
}

func (s *GRPCProviderServer) GetMetadata(ctx context.Context, req *tfprotov5.GetMetadataRequest) (*tfprotov5.GetMetadataResponse, error) {
	ctx = logging.InitContext(ctx)

	logging.HelperSchemaTrace(ctx, "Getting provider metadata")

	resp := &tfprotov5.GetMetadataResponse{
		DataSources:        make([]tfprotov5.DataSourceMetadata, 0, len(s.provider.DataSourcesMap)),
		Functions:          make([]tfprotov5.FunctionMetadata, 0),
		Resources:          make([]tfprotov5.ResourceMetadata, 0, len(s.provider.ResourcesMap)),
		ServerCapabilities: s.serverCapabilities(),
	}

	for typeName := range s.provider.DataSourcesMap {
		resp.DataSources = append(resp.DataSources, tfprotov5.DataSourceMetadata{
			TypeName: typeName,
		})
	}

	for typeName := range s.provider.ResourcesMap {
		resp.Resources = append(resp.Resources, tfprotov5.ResourceMetadata{
			TypeName: typeName,
		})
	}

	return resp, nil
}

func (s *GRPCProviderServer) GetProviderSchema(ctx context.Context, req *tfprotov5.GetProviderSchemaRequest) (*tfprotov5.GetProviderSchemaResponse, error) {
	ctx = logging.InitContext(ctx)

	logging.HelperSchemaTrace(ctx, "Getting provider schema")

	resp := &tfprotov5.GetProviderSchemaResponse{
		DataSourceSchemas:  make(map[string]*tfprotov5.Schema, len(s.provider.DataSourcesMap)),
		Functions:          make(map[string]*tfprotov5.Function, 0),
		ResourceSchemas:    make(map[string]*tfprotov5.Schema, len(s.provider.ResourcesMap)),
		ServerCapabilities: s.serverCapabilities(),
	}

	resp.Provider = &tfprotov5.Schema{
		Block: convert.ConfigSchemaToProto(ctx, s.getProviderSchemaBlock()),
	}

	resp.ProviderMeta = &tfprotov5.Schema{
		Block: convert.ConfigSchemaToProto(ctx, s.getProviderMetaSchemaBlock()),
	}

	for typ, res := range s.provider.ResourcesMap {
		logging.HelperSchemaTrace(ctx, "Found resource type", map[string]interface{}{logging.KeyResourceType: typ})

		resp.ResourceSchemas[typ] = &tfprotov5.Schema{
			Version: int64(res.SchemaVersion),
			Block:   convert.ConfigSchemaToProto(ctx, res.CoreConfigSchema()),
		}
	}

	for typ, dat := range s.provider.DataSourcesMap {
		logging.HelperSchemaTrace(ctx, "Found data source type", map[string]interface{}{logging.KeyDataSourceType: typ})

		resp.DataSourceSchemas[typ] = &tfprotov5.Schema{
			Version: int64(dat.SchemaVersion),
			Block:   convert.ConfigSchemaToProto(ctx, dat.CoreConfigSchema()),
		}
	}

	return resp, nil
}

func (s *GRPCProviderServer) getProviderSchemaBlock() *configschema.Block {
	return InternalMap(s.provider.Schema).CoreConfigSchema()
}

func (s *GRPCProviderServer) getProviderMetaSchemaBlock() *configschema.Block {
	return InternalMap(s.provider.ProviderMetaSchema).CoreConfigSchema()
}

func (s *GRPCProviderServer) getResourceSchemaBlock(name string) *configschema.Block {
	res := s.provider.ResourcesMap[name]
	return res.CoreConfigSchema()
}

func (s *GRPCProviderServer) getDatasourceSchemaBlock(name string) *configschema.Block {
	dat := s.provider.DataSourcesMap[name]
	return dat.CoreConfigSchema()
}

func (s *GRPCProviderServer) PrepareProviderConfig(ctx context.Context, req *tfprotov5.PrepareProviderConfigRequest) (*tfprotov5.PrepareProviderConfigResponse, error) {
	ctx = logging.InitContext(ctx)
	resp := &tfprotov5.PrepareProviderConfigResponse{}

	logging.HelperSchemaTrace(ctx, "Preparing provider configuration")

	schemaBlock := s.getProviderSchemaBlock()

	configVal, err := msgpack.Unmarshal(req.Config.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// lookup any required, top-level attributes that are Null, and see if we
	// have a Default value available.
	configVal, err = cty.Transform(configVal, func(path cty.Path, val cty.Value) (cty.Value, error) {
		// we're only looking for top-level attributes
		if len(path) != 1 {
			return val, nil
		}

		// nothing to do if we already have a value
		if !val.IsNull() {
			return val, nil
		}

		// get the Schema definition for this attribute
		getAttr, ok := path[0].(cty.GetAttrStep)
		// these should all exist, but just ignore anything strange
		if !ok {
			return val, nil
		}

		attrSchema := s.provider.Schema[getAttr.Name]
		// continue to ignore anything that doesn't match
		if attrSchema == nil {
			return val, nil
		}

		// this is deprecated, so don't set it
		if attrSchema.Deprecated != "" {
			return val, nil
		}

		// find a default value if it exists
		def, err := attrSchema.DefaultValue()
		if err != nil {
			return val, fmt.Errorf("error getting default for %q: %w", getAttr.Name, err)
		}

		// no default
		if def == nil {
			return val, nil
		}

		// create a cty.Value and make sure it's the correct type
		tmpVal := hcl2shim.HCL2ValueFromConfigValue(def)

		// helper/schema used to allow setting "" to a bool
		if val.Type() == cty.Bool && tmpVal.RawEquals(cty.StringVal("")) {
			// return a warning about the conversion
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, "provider set empty string as default value for bool "+getAttr.Name)
			tmpVal = cty.False
		}

		val, err = ctyconvert.Convert(tmpVal, val.Type())
		if err != nil {
			return val, fmt.Errorf("error setting default for %q: %w", getAttr.Name, err)
		}

		return val, nil
	})
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	configVal, err = schemaBlock.CoerceValue(configVal)
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// Ensure there are no nulls that will cause helper/schema to panic.
	if err := validateConfigNulls(ctx, configVal, nil); err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	config := terraform.NewResourceConfigShimmed(configVal, schemaBlock)

	logging.HelperSchemaTrace(ctx, "Calling downstream")
	resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, s.provider.Validate(config))
	logging.HelperSchemaTrace(ctx, "Called downstream")

	preparedConfigMP, err := msgpack.Marshal(configVal, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	resp.PreparedConfig = &tfprotov5.DynamicValue{MsgPack: preparedConfigMP}

	return resp, nil
}

func (s *GRPCProviderServer) ValidateResourceTypeConfig(ctx context.Context, req *tfprotov5.ValidateResourceTypeConfigRequest) (*tfprotov5.ValidateResourceTypeConfigResponse, error) {
	ctx = logging.InitContext(ctx)
	resp := &tfprotov5.ValidateResourceTypeConfigResponse{}

	schemaBlock := s.getResourceSchemaBlock(req.TypeName)

	configVal, err := msgpack.Unmarshal(req.Config.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	config := terraform.NewResourceConfigShimmed(configVal, schemaBlock)

	logging.HelperSchemaTrace(ctx, "Calling downstream")
	resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, s.provider.ValidateResource(req.TypeName, config))
	logging.HelperSchemaTrace(ctx, "Called downstream")

	return resp, nil
}

func (s *GRPCProviderServer) ValidateDataSourceConfig(ctx context.Context, req *tfprotov5.ValidateDataSourceConfigRequest) (*tfprotov5.ValidateDataSourceConfigResponse, error) {
	ctx = logging.InitContext(ctx)
	resp := &tfprotov5.ValidateDataSourceConfigResponse{}

	schemaBlock := s.getDatasourceSchemaBlock(req.TypeName)

	configVal, err := msgpack.Unmarshal(req.Config.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// Ensure there are no nulls that will cause helper/schema to panic.
	if err := validateConfigNulls(ctx, configVal, nil); err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	config := terraform.NewResourceConfigShimmed(configVal, schemaBlock)

	logging.HelperSchemaTrace(ctx, "Calling downstream")
	resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, s.provider.ValidateDataSource(req.TypeName, config))
	logging.HelperSchemaTrace(ctx, "Called downstream")

	return resp, nil
}

func (s *GRPCProviderServer) UpgradeResourceState(ctx context.Context, req *tfprotov5.UpgradeResourceStateRequest) (*tfprotov5.UpgradeResourceStateResponse, error) {
	ctx = logging.InitContext(ctx)
	resp := &tfprotov5.UpgradeResourceStateResponse{}

	res, ok := s.provider.ResourcesMap[req.TypeName]
	if !ok {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, fmt.Errorf("unknown resource type: %s", req.TypeName))
		return resp, nil
	}
	schemaBlock := s.getResourceSchemaBlock(req.TypeName)

	version := int(req.Version)

	jsonMap := map[string]interface{}{}
	var err error

	switch {
	// We first need to upgrade a flatmap state if it exists.
	// There should never be both a JSON and Flatmap state in the request.
	case len(req.RawState.Flatmap) > 0:
		logging.HelperSchemaTrace(ctx, "Upgrading flatmap state")

		jsonMap, version, err = s.upgradeFlatmapState(ctx, version, req.RawState.Flatmap, res)
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}
	// if there's a JSON state, we need to decode it.
	case len(req.RawState.JSON) > 0:
		if res.UseJSONNumber {
			err = unmarshalJSON(req.RawState.JSON, &jsonMap)
		} else {
			err = json.Unmarshal(req.RawState.JSON, &jsonMap)
		}
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}
	default:
		logging.HelperSchemaDebug(ctx, "no state provided to upgrade")
		return resp, nil
	}

	// complete the upgrade of the JSON states
	logging.HelperSchemaTrace(ctx, "Upgrading JSON state")

	jsonMap, err = s.upgradeJSONState(ctx, version, jsonMap, res)
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// The provider isn't required to clean out removed fields
	s.removeAttributes(ctx, jsonMap, schemaBlock.ImpliedType())

	// now we need to turn the state into the default json representation, so
	// that it can be re-decoded using the actual schema.
	val, err := JSONMapToStateValue(jsonMap, schemaBlock)
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// Now we need to make sure blocks are represented correctly, which means
	// that missing blocks are empty collections, rather than null.
	// First we need to CoerceValue to ensure that all object types match.
	val, err = schemaBlock.CoerceValue(val)
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}
	// Normalize the value and fill in any missing blocks.
	val = objchange.NormalizeObjectFromLegacySDK(val, schemaBlock)

	// encode the final state to the expected msgpack format
	newStateMP, err := msgpack.Marshal(val, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	resp.UpgradedState = &tfprotov5.DynamicValue{MsgPack: newStateMP}
	return resp, nil
}

// upgradeFlatmapState takes a legacy flatmap state, upgrades it using Migrate
// state if necessary, and converts it to the new JSON state format decoded as a
// map[string]interface{}.
// upgradeFlatmapState returns the json map along with the corresponding schema
// version.
func (s *GRPCProviderServer) upgradeFlatmapState(_ context.Context, version int, m map[string]string, res *Resource) (map[string]interface{}, int, error) {
	// this will be the version we've upgraded so, defaulting to the given
	// version in case no migration was called.
	upgradedVersion := version

	// first determine if we need to call the legacy MigrateState func
	requiresMigrate := version < res.SchemaVersion

	schemaType := res.CoreConfigSchema().ImpliedType()

	// if there are any StateUpgraders, then we need to only compare
	// against the first version there
	if len(res.StateUpgraders) > 0 {
		requiresMigrate = version < res.StateUpgraders[0].Version
	}

	if requiresMigrate && res.MigrateState == nil {
		// Providers were previously allowed to bump the version
		// without declaring MigrateState.
		// If there are further upgraders, then we've only updated that far.
		if len(res.StateUpgraders) > 0 {
			schemaType = res.StateUpgraders[0].Type
			upgradedVersion = res.StateUpgraders[0].Version
		}
	} else if requiresMigrate {
		is := &terraform.InstanceState{
			ID:         m["id"],
			Attributes: m,
			Meta: map[string]interface{}{
				"schema_version": strconv.Itoa(version),
			},
		}
		is, err := res.MigrateState(version, is, s.provider.Meta())
		if err != nil {
			return nil, 0, err
		}

		// re-assign the map in case there was a copy made, making sure to keep
		// the ID
		m := is.Attributes
		m["id"] = is.ID

		// if there are further upgraders, then we've only updated that far
		if len(res.StateUpgraders) > 0 {
			schemaType = res.StateUpgraders[0].Type
			upgradedVersion = res.StateUpgraders[0].Version
		}
	} else {
		// the schema version may be newer than the MigrateState functions
		// handled and older than the current, but still stored in the flatmap
		// form. If that's the case, we need to find the correct schema type to
		// convert the state.
		for _, upgrader := range res.StateUpgraders {
			if upgrader.Version == version {
				schemaType = upgrader.Type
				break
			}
		}
	}

	// now we know the state is up to the latest version that handled the
	// flatmap format state. Now we can upgrade the format and continue from
	// there.
	newConfigVal, err := hcl2shim.HCL2ValueFromFlatmap(m, schemaType)
	if err != nil {
		return nil, 0, err
	}

	jsonMap, err := stateValueToJSONMap(newConfigVal, schemaType, res.UseJSONNumber)
	return jsonMap, upgradedVersion, err
}

func (s *GRPCProviderServer) upgradeJSONState(ctx context.Context, version int, m map[string]interface{}, res *Resource) (map[string]interface{}, error) {
	var err error

	for _, upgrader := range res.StateUpgraders {
		if version != upgrader.Version {
			continue
		}

		m, err = upgrader.Upgrade(ctx, m, s.provider.Meta())
		if err != nil {
			return nil, err
		}
		version++
	}

	return m, nil
}

// Remove any attributes no longer present in the schema, so that the json can
// be correctly decoded.
func (s *GRPCProviderServer) removeAttributes(ctx context.Context, v interface{}, ty cty.Type) {
	// we're only concerned with finding maps that corespond to object
	// attributes
	switch v := v.(type) {
	case []interface{}:
		// If these aren't blocks the next call will be a noop
		if ty.IsListType() || ty.IsSetType() {
			eTy := ty.ElementType()
			for _, eV := range v {
				s.removeAttributes(ctx, eV, eTy)
			}
		}
		return
	case map[string]interface{}:
		// map blocks aren't yet supported, but handle this just in case
		if ty.IsMapType() {
			eTy := ty.ElementType()
			for _, eV := range v {
				s.removeAttributes(ctx, eV, eTy)
			}
			return
		}

		if ty == cty.DynamicPseudoType {
			logging.HelperSchemaDebug(ctx, "ignoring dynamic block", map[string]interface{}{"block": v})
			return
		}

		if !ty.IsObjectType() {
			// This shouldn't happen, and will fail to decode further on, so
			// there's no need to handle it here.
			logging.HelperSchemaWarn(ctx, "unexpected type for map in JSON state", map[string]interface{}{"type": ty})
			return
		}

		attrTypes := ty.AttributeTypes()
		for attr, attrV := range v {
			attrTy, ok := attrTypes[attr]
			if !ok {
				logging.HelperSchemaDebug(ctx, "attribute no longer present in schema", map[string]interface{}{"attribute": attr})
				delete(v, attr)
				continue
			}

			s.removeAttributes(ctx, attrV, attrTy)
		}
	}
}

func (s *GRPCProviderServer) StopProvider(ctx context.Context, _ *tfprotov5.StopProviderRequest) (*tfprotov5.StopProviderResponse, error) {
	ctx = logging.InitContext(ctx)

	logging.HelperSchemaTrace(ctx, "Stopping provider")

	s.stopMu.Lock()
	defer s.stopMu.Unlock()

	// stop
	close(s.stopCh)
	// reset the stop signal
	s.stopCh = make(chan struct{})

	logging.HelperSchemaTrace(ctx, "Stopped provider")

	return &tfprotov5.StopProviderResponse{}, nil
}

func (s *GRPCProviderServer) ConfigureProvider(ctx context.Context, req *tfprotov5.ConfigureProviderRequest) (*tfprotov5.ConfigureProviderResponse, error) {
	ctx = logging.InitContext(ctx)
	resp := &tfprotov5.ConfigureProviderResponse{}

	schemaBlock := s.getProviderSchemaBlock()

	configVal, err := msgpack.Unmarshal(req.Config.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	s.provider.TerraformVersion = req.TerraformVersion

	// Ensure there are no nulls that will cause helper/schema to panic.
	if err := validateConfigNulls(ctx, configVal, nil); err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	config := terraform.NewResourceConfigShimmed(configVal, schemaBlock)

	// CtyValue is the raw protocol configuration data from newer APIs.
	//
	// This field was only added as a targeted fix for passing raw protocol data
	// through the existing (helper/schema.Provider).Configure() exported method
	// and is only populated in that situation. The data could theoretically be
	// set in the NewResourceConfigShimmed() function, however the consequences
	// of doing this were not investigated at the time the fix was introduced.
	//
	// Reference: https://github.com/hashicorp/terraform-plugin-sdk/issues/1270
	config.CtyValue = configVal

	// TODO: remove global stop context hack
	// This attaches a global stop synchro'd context onto the provider.Configure
	// request scoped context. This provides a substitute for the removed provider.StopContext()
	// function. Ideally a provider should migrate to the context aware API that receives
	// request scoped contexts, however this is a large undertaking for very large providers.
	ctxHack := context.WithValue(ctx, StopContextKey, s.StopContext(context.Background()))

	// NOTE: This is a hack to pass the deferral_allowed field from the Terraform client to the
	// underlying (provider).Configure function, which cannot be changed because the function
	// signature is public. (╯°□°)╯︵ ┻━┻
	s.provider.deferralAllowed = configureDeferralAllowed(req.ClientCapabilities)

	logging.HelperSchemaTrace(ctx, "Calling downstream")
	diags := s.provider.Configure(ctxHack, config)
	logging.HelperSchemaTrace(ctx, "Called downstream")

	resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, diags)

	if s.provider.providerDeferred != nil {
		// Check if a deferred response was incorrectly set on the provider. This would cause an error during later RPCs.
		if !s.provider.deferralAllowed {
			resp.Diagnostics = append(resp.Diagnostics, &tfprotov5.Diagnostic{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Invalid Deferred Provider Response",
				Detail: "Provider configured a deferred response for all resources and data sources but the Terraform request " +
					"did not indicate support for deferred actions. This is an issue with the provider and should be reported to the provider developers.",
			})
		} else {
			logging.HelperSchemaDebug(
				ctx,
				"Provider has configured a deferred response, all associated resources and data sources will automatically return a deferred response.",
				map[string]interface{}{
					logging.KeyDeferredReason: s.provider.providerDeferred.Reason.String(),
				},
			)
		}
	}

	return resp, nil
}

func (s *GRPCProviderServer) ReadResource(ctx context.Context, req *tfprotov5.ReadResourceRequest) (*tfprotov5.ReadResourceResponse, error) {
	ctx = logging.InitContext(ctx)
	resp := &tfprotov5.ReadResourceResponse{
		// helper/schema did previously handle private data during refresh, but
		// core is now going to expect this to be maintained in order to
		// persist it in the state.
		Private: req.Private,
	}

	res, ok := s.provider.ResourcesMap[req.TypeName]
	if !ok {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, fmt.Errorf("unknown resource type: %s", req.TypeName))
		return resp, nil
	}
	schemaBlock := s.getResourceSchemaBlock(req.TypeName)

	if s.provider.providerDeferred != nil {
		logging.HelperSchemaDebug(
			ctx,
			"Provider has deferred response configured, automatically returning deferred response.",
			map[string]interface{}{
				logging.KeyDeferredReason: s.provider.providerDeferred.Reason.String(),
			},
		)

		resp.NewState = req.CurrentState
		resp.Deferred = &tfprotov5.Deferred{
			Reason: tfprotov5.DeferredReason(s.provider.providerDeferred.Reason),
		}
		return resp, nil
	}

	stateVal, err := msgpack.Unmarshal(req.CurrentState.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	instanceState, err := res.ShimInstanceStateFromValue(stateVal)
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}
	instanceState.RawState = stateVal

	private := make(map[string]interface{})
	if len(req.Private) > 0 {
		if err := json.Unmarshal(req.Private, &private); err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}
	}
	instanceState.Meta = private

	pmSchemaBlock := s.getProviderMetaSchemaBlock()
	if pmSchemaBlock != nil && req.ProviderMeta != nil {
		providerSchemaVal, err := msgpack.Unmarshal(req.ProviderMeta.MsgPack, pmSchemaBlock.ImpliedType())
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}
		instanceState.ProviderMeta = providerSchemaVal
	}

	newInstanceState, diags := res.RefreshWithoutUpgrade(ctx, instanceState, s.provider.Meta())
	resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, diags)
	if diags.HasError() {
		return resp, nil
	}

	if newInstanceState == nil || newInstanceState.ID == "" {
		// The old provider API used an empty id to signal that the remote
		// object appears to have been deleted, but our new protocol expects
		// to see a null value (in the cty sense) in that case.
		newStateMP, err := msgpack.Marshal(cty.NullVal(schemaBlock.ImpliedType()), schemaBlock.ImpliedType())
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		}
		resp.NewState = &tfprotov5.DynamicValue{
			MsgPack: newStateMP,
		}
		return resp, nil
	}

	// helper/schema should always copy the ID over, but do it again just to be safe
	newInstanceState.Attributes["id"] = newInstanceState.ID

	newStateVal, err := hcl2shim.HCL2ValueFromFlatmap(newInstanceState.Attributes, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	newStateVal = normalizeNullValues(newStateVal, stateVal, false)
	newStateVal = copyTimeoutValues(newStateVal, stateVal)

	newStateMP, err := msgpack.Marshal(newStateVal, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	resp.NewState = &tfprotov5.DynamicValue{
		MsgPack: newStateMP,
	}

	return resp, nil
}

func (s *GRPCProviderServer) PlanResourceChange(ctx context.Context, req *tfprotov5.PlanResourceChangeRequest) (*tfprotov5.PlanResourceChangeResponse, error) {
	ctx = logging.InitContext(ctx)
	resp := &tfprotov5.PlanResourceChangeResponse{}

	res, ok := s.provider.ResourcesMap[req.TypeName]
	if !ok {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, fmt.Errorf("unknown resource type: %s", req.TypeName))
		return resp, nil
	}
	schemaBlock := s.getResourceSchemaBlock(req.TypeName)

	// This is a signal to Terraform Core that we're doing the best we can to
	// shim the legacy type system of the SDK onto the Terraform type system
	// but we need it to cut us some slack. This setting should not be taken
	// forward to any new SDK implementations, since setting it prevents us
	// from catching certain classes of provider bug that can lead to
	// confusing downstream errors.
	if !res.EnableLegacyTypeSystemPlanErrors {
		//nolint:staticcheck // explicitly for this SDK
		resp.UnsafeToUseLegacyTypeSystem = true
	}

	// Provider deferred response is present and the resource hasn't opted-in to CustomizeDiff being called, return early
	// with proposed new state as a best effort for PlannedState.
	if s.provider.providerDeferred != nil && !res.ResourceBehavior.ProviderDeferred.EnablePlanModification {
		logging.HelperSchemaDebug(
			ctx,
			"Provider has deferred response configured, automatically returning deferred response.",
			map[string]interface{}{
				logging.KeyDeferredReason: s.provider.providerDeferred.Reason.String(),
			},
		)

		resp.PlannedState = req.ProposedNewState
		resp.PlannedPrivate = req.PriorPrivate
		resp.Deferred = &tfprotov5.Deferred{
			Reason: tfprotov5.DeferredReason(s.provider.providerDeferred.Reason),
		}
		return resp, nil
	}

	priorStateVal, err := msgpack.Unmarshal(req.PriorState.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	create := priorStateVal.IsNull()

	proposedNewStateVal, err := msgpack.Unmarshal(req.ProposedNewState.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// We don't usually plan destroys, but this can return early in any case.
	if proposedNewStateVal.IsNull() {
		resp.PlannedState = req.ProposedNewState
		resp.PlannedPrivate = req.PriorPrivate
		return resp, nil
	}

	configVal, err := msgpack.Unmarshal(req.Config.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	priorState, err := res.ShimInstanceStateFromValue(priorStateVal)
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}
	priorState.RawState = priorStateVal
	priorState.RawPlan = proposedNewStateVal
	priorState.RawConfig = configVal
	priorPrivate := make(map[string]interface{})
	if len(req.PriorPrivate) > 0 {
		if err := json.Unmarshal(req.PriorPrivate, &priorPrivate); err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}
	}

	priorState.Meta = priorPrivate

	pmSchemaBlock := s.getProviderMetaSchemaBlock()
	if pmSchemaBlock != nil && req.ProviderMeta != nil {
		providerSchemaVal, err := msgpack.Unmarshal(req.ProviderMeta.MsgPack, pmSchemaBlock.ImpliedType())
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}
		priorState.ProviderMeta = providerSchemaVal
	}

	// Ensure there are no nulls that will cause helper/schema to panic.
	if err := validateConfigNulls(ctx, proposedNewStateVal, nil); err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// turn the proposed state into a legacy configuration
	cfg := terraform.NewResourceConfigShimmed(proposedNewStateVal, schemaBlock)

	diff, err := res.SimpleDiff(ctx, priorState, cfg, s.provider.Meta())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// if this is a new instance, we need to make sure ID is going to be computed
	if create {
		if diff == nil {
			diff = terraform.NewInstanceDiff()
		}

		diff.Attributes["id"] = &terraform.ResourceAttrDiff{
			NewComputed: true,
		}
	}

	if diff == nil || len(diff.Attributes) == 0 {
		// schema.Provider.Diff returns nil if it ends up making a diff with no
		// changes, but our new interface wants us to return an actual change
		// description that _shows_ there are no changes. This is always the
		// prior state, because we force a diff above if this is a new instance.
		resp.PlannedState = req.PriorState
		resp.PlannedPrivate = req.PriorPrivate
		return resp, nil
	}

	if priorState == nil {
		priorState = &terraform.InstanceState{}
	}

	// now we need to apply the diff to the prior state, so get the planned state
	plannedAttrs, err := diff.Apply(priorState.Attributes, schemaBlock)

	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	plannedStateVal, err := hcl2shim.HCL2ValueFromFlatmap(plannedAttrs, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	plannedStateVal, err = schemaBlock.CoerceValue(plannedStateVal)
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	plannedStateVal = normalizeNullValues(plannedStateVal, proposedNewStateVal, false)

	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	plannedStateVal = copyTimeoutValues(plannedStateVal, proposedNewStateVal)

	// The old SDK code has some imprecisions that cause it to sometimes
	// generate differences that the SDK itself does not consider significant
	// but Terraform Core would. To avoid producing weird do-nothing diffs
	// in that case, we'll check if the provider as produced something we
	// think is "equivalent" to the prior state and just return the prior state
	// itself if so, thus ensuring that Terraform Core will treat this as
	// a no-op. See the docs for ValuesSDKEquivalent for some caveats on its
	// accuracy.
	forceNoChanges := false
	if hcl2shim.ValuesSDKEquivalent(priorStateVal, plannedStateVal) {
		plannedStateVal = priorStateVal
		forceNoChanges = true
	}

	// if this was creating the resource, we need to set any remaining computed
	// fields
	if create {
		plannedStateVal = SetUnknowns(plannedStateVal, schemaBlock)
	}

	plannedMP, err := msgpack.Marshal(plannedStateVal, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}
	resp.PlannedState = &tfprotov5.DynamicValue{
		MsgPack: plannedMP,
	}

	// encode any timeouts into the diff Meta
	t := &ResourceTimeout{}
	if err := t.ConfigDecode(res, cfg); err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	if err := t.DiffEncode(diff); err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// Now we need to store any NewExtra values, which are where any actual
	// StateFunc modified config fields are hidden.
	privateMap := diff.Meta
	if privateMap == nil {
		privateMap = map[string]interface{}{}
	}

	newExtra := map[string]interface{}{}

	for k, v := range diff.Attributes {
		if v.NewExtra != nil {
			newExtra[k] = v.NewExtra
		}
	}
	privateMap[newExtraKey] = newExtra

	// the Meta field gets encoded into PlannedPrivate
	plannedPrivate, err := json.Marshal(privateMap)
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}
	resp.PlannedPrivate = plannedPrivate

	// collect the attributes that require instance replacement, and convert
	// them to cty.Paths.
	var requiresNew []string
	if !forceNoChanges {
		for attr, d := range diff.Attributes {
			if d.RequiresNew {
				requiresNew = append(requiresNew, attr)
			}
		}
	}

	// If anything requires a new resource already, or the "id" field indicates
	// that we will be creating a new resource, then we need to add that to
	// RequiresReplace so that core can tell if the instance is being replaced
	// even if changes are being suppressed via "ignore_changes".
	id := plannedStateVal.GetAttr("id")
	if len(requiresNew) > 0 || id.IsNull() || !id.IsKnown() {
		requiresNew = append(requiresNew, "id")
	}

	requiresReplace, err := hcl2shim.RequiresReplace(requiresNew, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// convert these to the protocol structures
	for _, p := range requiresReplace {
		resp.RequiresReplace = append(resp.RequiresReplace, pathToAttributePath(p))
	}

	// Provider deferred response is present, add the deferred response alongside the provider-modified plan
	if s.provider.providerDeferred != nil {
		logging.HelperSchemaDebug(
			ctx,
			"Provider has deferred response configured, returning deferred response with modified plan.",
			map[string]interface{}{
				logging.KeyDeferredReason: s.provider.providerDeferred.Reason.String(),
			},
		)

		resp.Deferred = &tfprotov5.Deferred{
			Reason: tfprotov5.DeferredReason(s.provider.providerDeferred.Reason),
		}
	}

	return resp, nil
}

func (s *GRPCProviderServer) ApplyResourceChange(ctx context.Context, req *tfprotov5.ApplyResourceChangeRequest) (*tfprotov5.ApplyResourceChangeResponse, error) {
	ctx = logging.InitContext(ctx)
	resp := &tfprotov5.ApplyResourceChangeResponse{
		// Start with the existing state as a fallback
		NewState: req.PriorState,
	}

	res, ok := s.provider.ResourcesMap[req.TypeName]
	if !ok {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, fmt.Errorf("unknown resource type: %s", req.TypeName))
		return resp, nil
	}
	schemaBlock := s.getResourceSchemaBlock(req.TypeName)

	priorStateVal, err := msgpack.Unmarshal(req.PriorState.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	plannedStateVal, err := msgpack.Unmarshal(req.PlannedState.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	configVal, err := msgpack.Unmarshal(req.Config.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	priorState, err := res.ShimInstanceStateFromValue(priorStateVal)
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	private := make(map[string]interface{})
	if len(req.PlannedPrivate) > 0 {
		if err := json.Unmarshal(req.PlannedPrivate, &private); err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}
	}

	var diff *terraform.InstanceDiff
	destroy := false

	// a null state means we are destroying the instance
	if plannedStateVal.IsNull() {
		destroy = true
		diff = &terraform.InstanceDiff{
			Attributes: make(map[string]*terraform.ResourceAttrDiff),
			Meta:       make(map[string]interface{}),
			Destroy:    true,
			RawPlan:    plannedStateVal,
			RawState:   priorStateVal,
			RawConfig:  configVal,
		}
	} else {
		diff, err = DiffFromValues(ctx, priorStateVal, plannedStateVal, configVal, stripResourceModifiers(res))
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}
	}

	if diff == nil {
		diff = &terraform.InstanceDiff{
			Attributes: make(map[string]*terraform.ResourceAttrDiff),
			Meta:       make(map[string]interface{}),
			RawPlan:    plannedStateVal,
			RawState:   priorStateVal,
			RawConfig:  configVal,
		}
	}

	// add NewExtra Fields that may have been stored in the private data
	if newExtra := private[newExtraKey]; newExtra != nil {
		for k, v := range newExtra.(map[string]interface{}) {
			d := diff.Attributes[k]

			if d == nil {
				d = &terraform.ResourceAttrDiff{}
			}

			d.NewExtra = v
			diff.Attributes[k] = d
		}
	}

	if private != nil {
		diff.Meta = private
	}

	for k, d := range diff.Attributes {
		// We need to turn off any RequiresNew. There could be attributes
		// without changes in here inserted by helper/schema, but if they have
		// RequiresNew then the state will be dropped from the ResourceData.
		d.RequiresNew = false

		// Check that any "removed" attributes that don't actually exist in the
		// prior state, or helper/schema will confuse itself
		if d.NewRemoved {
			if _, ok := priorState.Attributes[k]; !ok {
				delete(diff.Attributes, k)
			}
		}
	}

	pmSchemaBlock := s.getProviderMetaSchemaBlock()
	if pmSchemaBlock != nil && req.ProviderMeta != nil {
		providerSchemaVal, err := msgpack.Unmarshal(req.ProviderMeta.MsgPack, pmSchemaBlock.ImpliedType())
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}
		priorState.ProviderMeta = providerSchemaVal
	}

	newInstanceState, diags := res.Apply(ctx, priorState, diff, s.provider.Meta())
	resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, diags)

	newStateVal := cty.NullVal(schemaBlock.ImpliedType())

	// Always return a null value for destroy.
	// While this is usually indicated by a nil state, check for missing ID or
	// attributes in the case of a provider failure.
	if destroy || newInstanceState == nil || newInstanceState.Attributes == nil || newInstanceState.ID == "" {
		newStateMP, err := msgpack.Marshal(newStateVal, schemaBlock.ImpliedType())
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}
		resp.NewState = &tfprotov5.DynamicValue{
			MsgPack: newStateMP,
		}
		return resp, nil
	}

	// We keep the null val if we destroyed the resource, otherwise build the
	// entire object, even if the new state was nil.
	newStateVal, err = StateValueFromInstanceState(newInstanceState, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	newStateVal = normalizeNullValues(newStateVal, plannedStateVal, true)

	newStateVal = copyTimeoutValues(newStateVal, plannedStateVal)

	newStateMP, err := msgpack.Marshal(newStateVal, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}
	resp.NewState = &tfprotov5.DynamicValue{
		MsgPack: newStateMP,
	}

	meta, err := json.Marshal(newInstanceState.Meta)
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}
	resp.Private = meta

	// This is a signal to Terraform Core that we're doing the best we can to
	// shim the legacy type system of the SDK onto the Terraform type system
	// but we need it to cut us some slack. This setting should not be taken
	// forward to any new SDK implementations, since setting it prevents us
	// from catching certain classes of provider bug that can lead to
	// confusing downstream errors.
	if !res.EnableLegacyTypeSystemApplyErrors {
		//nolint:staticcheck // explicitly for this SDK
		resp.UnsafeToUseLegacyTypeSystem = true
	}

	return resp, nil
}

func (s *GRPCProviderServer) ImportResourceState(ctx context.Context, req *tfprotov5.ImportResourceStateRequest) (*tfprotov5.ImportResourceStateResponse, error) {
	ctx = logging.InitContext(ctx)
	resp := &tfprotov5.ImportResourceStateResponse{}

	info := &terraform.InstanceInfo{
		Type: req.TypeName,
	}

	if s.provider.providerDeferred != nil {
		logging.HelperSchemaDebug(
			ctx,
			"Provider has deferred response configured, automatically returning deferred response.",
			map[string]interface{}{
				logging.KeyDeferredReason: s.provider.providerDeferred.Reason.String(),
			},
		)

		// The logic for ensuring the resource type is supported by this provider is inside of (provider).ImportState
		// We need to check to ensure the resource type is supported before using the schema
		_, ok := s.provider.ResourcesMap[req.TypeName]
		if !ok {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, fmt.Errorf("unknown resource type: %s", req.TypeName))
			return resp, nil
		}

		// Since we are automatically deferring, send back an unknown value for the imported object
		schemaBlock := s.getResourceSchemaBlock(req.TypeName)
		unknownVal := cty.UnknownVal(schemaBlock.ImpliedType())
		unknownStateMp, err := msgpack.Marshal(unknownVal, schemaBlock.ImpliedType())
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}

		resp.ImportedResources = []*tfprotov5.ImportedResource{
			{
				TypeName: req.TypeName,
				State: &tfprotov5.DynamicValue{
					MsgPack: unknownStateMp,
				},
			},
		}

		resp.Deferred = &tfprotov5.Deferred{
			Reason: tfprotov5.DeferredReason(s.provider.providerDeferred.Reason),
		}

		return resp, nil
	}

	newInstanceStates, err := s.provider.ImportState(ctx, info, req.ID)
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	for _, is := range newInstanceStates {
		// copy the ID again just to be sure it wasn't missed
		is.Attributes["id"] = is.ID

		resourceType := is.Ephemeral.Type
		if resourceType == "" {
			resourceType = req.TypeName
		}

		schemaBlock := s.getResourceSchemaBlock(resourceType)
		newStateVal, err := hcl2shim.HCL2ValueFromFlatmap(is.Attributes, schemaBlock.ImpliedType())
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}

		// Normalize the value and fill in any missing blocks.
		newStateVal = objchange.NormalizeObjectFromLegacySDK(newStateVal, schemaBlock)

		// Ensure any timeouts block is null in the imported state. There is no
		// configuration to read from during import, so it is never valid to
		// return a known value for the block.
		//
		// This is done without modifying HCL2ValueFromFlatmap or
		// NormalizeObjectFromLegacySDK to prevent other unexpected changes.
		//
		// Reference: https://github.com/hashicorp/terraform-plugin-sdk/issues/1145
		newStateType := newStateVal.Type()

		if newStateVal != cty.NilVal && !newStateVal.IsNull() && newStateType.IsObjectType() && newStateType.HasAttribute(TimeoutsConfigKey) {
			newStateValueMap := newStateVal.AsValueMap()
			newStateValueMap[TimeoutsConfigKey] = cty.NullVal(newStateType.AttributeType(TimeoutsConfigKey))
			newStateVal = cty.ObjectVal(newStateValueMap)
		}

		newStateMP, err := msgpack.Marshal(newStateVal, schemaBlock.ImpliedType())
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}

		meta, err := json.Marshal(is.Meta)
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}

		importedResource := &tfprotov5.ImportedResource{
			TypeName: resourceType,
			State: &tfprotov5.DynamicValue{
				MsgPack: newStateMP,
			},
			Private: meta,
		}

		resp.ImportedResources = append(resp.ImportedResources, importedResource)
	}

	return resp, nil
}

func (s *GRPCProviderServer) MoveResourceState(ctx context.Context, req *tfprotov5.MoveResourceStateRequest) (*tfprotov5.MoveResourceStateResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("MoveResourceState request is nil")
	}

	ctx = logging.InitContext(ctx)

	logging.HelperSchemaTrace(ctx, "Returning error for MoveResourceState")

	resp := &tfprotov5.MoveResourceStateResponse{}

	_, ok := s.provider.ResourcesMap[req.TargetTypeName]

	if !ok {
		resp.Diagnostics = []*tfprotov5.Diagnostic{
			{
				Severity: tfprotov5.DiagnosticSeverityError,
				Summary:  "Unknown Resource Type",
				Detail:   fmt.Sprintf("The %q resource type is not supported by this provider.", req.TargetTypeName),
			},
		}

		return resp, nil
	}

	resp.Diagnostics = []*tfprotov5.Diagnostic{
		{
			Severity: tfprotov5.DiagnosticSeverityError,
			Summary:  "Move Resource State Not Supported",
			Detail:   fmt.Sprintf("The %q resource type does not support moving resource state across resource types.", req.TargetTypeName),
		},
	}

	return resp, nil
}

func (s *GRPCProviderServer) ReadDataSource(ctx context.Context, req *tfprotov5.ReadDataSourceRequest) (*tfprotov5.ReadDataSourceResponse, error) {
	ctx = logging.InitContext(ctx)
	resp := &tfprotov5.ReadDataSourceResponse{}

	schemaBlock := s.getDatasourceSchemaBlock(req.TypeName)

	if s.provider.providerDeferred != nil {
		logging.HelperSchemaDebug(
			ctx,
			"Provider has deferred response configured, automatically returning deferred response.",
			map[string]interface{}{
				logging.KeyDeferredReason: s.provider.providerDeferred.Reason.String(),
			},
		)

		// Send an unknown value for the data source
		unknownVal := cty.UnknownVal(schemaBlock.ImpliedType())
		unknownStateMp, err := msgpack.Marshal(unknownVal, schemaBlock.ImpliedType())
		if err != nil {
			resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
			return resp, nil
		}

		resp.State = &tfprotov5.DynamicValue{
			MsgPack: unknownStateMp,
		}
		resp.Deferred = &tfprotov5.Deferred{
			Reason: tfprotov5.DeferredReason(s.provider.providerDeferred.Reason),
		}
		return resp, nil
	}

	configVal, err := msgpack.Unmarshal(req.Config.MsgPack, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// Ensure there are no nulls that will cause helper/schema to panic.
	if err := validateConfigNulls(ctx, configVal, nil); err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	config := terraform.NewResourceConfigShimmed(configVal, schemaBlock)

	// we need to still build the diff separately with the Read method to match
	// the old behavior
	res, ok := s.provider.DataSourcesMap[req.TypeName]
	if !ok {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, fmt.Errorf("unknown data source: %s", req.TypeName))
		return resp, nil
	}
	diff, err := res.Diff(ctx, nil, config, s.provider.Meta())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	// Not setting RawConfig here is okay, as ResourceData.GetRawConfig()
	// will return a NullVal of the schema if there is no InstanceDiff.
	if diff != nil {
		diff.RawConfig = configVal
	}

	// now we can get the new complete data source
	newInstanceState, diags := res.ReadDataApply(ctx, diff, s.provider.Meta())
	resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, diags)
	if diags.HasError() {
		return resp, nil
	}

	newStateVal, err := StateValueFromInstanceState(newInstanceState, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}

	newStateVal = copyTimeoutValues(newStateVal, configVal)

	newStateMP, err := msgpack.Marshal(newStateVal, schemaBlock.ImpliedType())
	if err != nil {
		resp.Diagnostics = convert.AppendProtoDiag(ctx, resp.Diagnostics, err)
		return resp, nil
	}
	resp.State = &tfprotov5.DynamicValue{
		MsgPack: newStateMP,
	}
	return resp, nil
}

func (s *GRPCProviderServer) CallFunction(ctx context.Context, req *tfprotov5.CallFunctionRequest) (*tfprotov5.CallFunctionResponse, error) {
	ctx = logging.InitContext(ctx)

	logging.HelperSchemaTrace(ctx, "Returning error for provider function call")

	resp := &tfprotov5.CallFunctionResponse{
		Error: &tfprotov5.FunctionError{
			Text: fmt.Sprintf("Function Not Found: No function named %q was found in the provider.", req.Name),
		},
	}

	return resp, nil
}

func (s *GRPCProviderServer) GetFunctions(ctx context.Context, req *tfprotov5.GetFunctionsRequest) (*tfprotov5.GetFunctionsResponse, error) {
	ctx = logging.InitContext(ctx)

	logging.HelperSchemaTrace(ctx, "Getting provider functions")

	resp := &tfprotov5.GetFunctionsResponse{
		Functions: make(map[string]*tfprotov5.Function, 0),
	}

	return resp, nil
}

func pathToAttributePath(path cty.Path) *tftypes.AttributePath {
	var steps []tftypes.AttributePathStep

	for _, step := range path {
		switch s := step.(type) {
		case cty.GetAttrStep:
			steps = append(steps,
				tftypes.AttributeName(s.Name),
			)
		case cty.IndexStep:
			ty := s.Key.Type()
			switch ty {
			case cty.Number:
				i, _ := s.Key.AsBigFloat().Int64()
				steps = append(steps,
					tftypes.ElementKeyInt(i),
				)
			case cty.String:
				steps = append(steps,
					tftypes.ElementKeyString(s.Key.AsString()),
				)
			}
		}
	}

	if len(steps) < 1 {
		return nil
	}
	return tftypes.NewAttributePathWithSteps(steps)
}

// helper/schema throws away timeout values from the config and stores them in
// the Private/Meta fields. we need to copy those values into the planned state
// so that core doesn't see a perpetual diff with the timeout block.
func copyTimeoutValues(to cty.Value, from cty.Value) cty.Value {
	// if `to` is null we are planning to remove it altogether.
	if to.IsNull() {
		return to
	}
	toAttrs := to.AsValueMap()
	// We need to remove the key since the hcl2shims will add a non-null block
	// because we can't determine if a single block was null from the flatmapped
	// values. This needs to conform to the correct schema for marshaling, so
	// change the value to null rather than deleting it from the object map.
	timeouts, ok := toAttrs[TimeoutsConfigKey]
	if ok {
		toAttrs[TimeoutsConfigKey] = cty.NullVal(timeouts.Type())
	}

	// if from is null then there are no timeouts to copy
	if from.IsNull() {
		return cty.ObjectVal(toAttrs)
	}

	fromAttrs := from.AsValueMap()
	timeouts, ok = fromAttrs[TimeoutsConfigKey]

	// timeouts shouldn't be unknown, but don't copy possibly invalid values either
	if !ok || timeouts.IsNull() || !timeouts.IsWhollyKnown() {
		// no timeouts block to copy
		return cty.ObjectVal(toAttrs)
	}

	toAttrs[TimeoutsConfigKey] = timeouts

	return cty.ObjectVal(toAttrs)
}

// stripResourceModifiers takes a *schema.Resource and returns a deep copy with all
// StateFuncs and CustomizeDiffs removed. This will be used during apply to
// create a diff from a planned state where the diff modifications have already
// been applied.
func stripResourceModifiers(r *Resource) *Resource {
	if r == nil {
		return nil
	}
	// start with a shallow copy
	newResource := new(Resource)
	*newResource = *r

	newResource.CustomizeDiff = nil
	newResource.Schema = map[string]*Schema{}

	for k, s := range r.SchemaMap() {
		newResource.Schema[k] = stripSchema(s)
	}

	return newResource
}

func stripSchema(s *Schema) *Schema {
	if s == nil {
		return nil
	}
	// start with a shallow copy
	newSchema := new(Schema)
	*newSchema = *s

	newSchema.StateFunc = nil

	switch e := newSchema.Elem.(type) {
	case *Schema:
		newSchema.Elem = stripSchema(e)
	case *Resource:
		newSchema.Elem = stripResourceModifiers(e)
	}

	return newSchema
}

// Zero values and empty containers may be interchanged by the apply process.
// When there is a discrepency between src and dst value being null or empty,
// prefer the src value. This takes a little more liberty with set types, since
// we can't correlate modified set values. In the case of sets, if the src set
// was wholly known we assume the value was correctly applied and copy that
// entirely to the new value.
// While apply prefers the src value, during plan we prefer dst whenever there
// is an unknown or a set is involved, since the plan can alter the value
// however it sees fit. This however means that a CustomizeDiffFunction may not
// be able to change a null to an empty value or vice versa, but that should be
// very uncommon nor was it reliable before 0.12 either.
func normalizeNullValues(dst, src cty.Value, apply bool) cty.Value {
	ty := dst.Type()
	if !src.IsNull() && !src.IsKnown() {
		// Return src during plan to retain unknown interpolated placeholders,
		// which could be lost if we're only updating a resource. If this is a
		// read scenario, then there shouldn't be any unknowns at all.
		if dst.IsNull() && !apply {
			return src
		}
		return dst
	}

	// Handle null/empty changes for collections during apply.
	// A change between null and empty values prefers src to make sure the state
	// is consistent between plan and apply.
	if ty.IsCollectionType() && apply {
		dstEmpty := !dst.IsNull() && dst.IsKnown() && dst.LengthInt() == 0
		srcEmpty := !src.IsNull() && src.IsKnown() && src.LengthInt() == 0

		if (src.IsNull() && dstEmpty) || (srcEmpty && dst.IsNull()) {
			return src
		}
	}

	// check the invariants that we need below, to ensure we are working with
	// non-null and known values.
	if src.IsNull() || !src.IsKnown() || !dst.IsKnown() {
		return dst
	}

	switch {
	case ty.IsMapType(), ty.IsObjectType():
		var dstMap map[string]cty.Value
		if !dst.IsNull() {
			dstMap = dst.AsValueMap()
		}
		if dstMap == nil {
			dstMap = map[string]cty.Value{}
		}

		srcMap := src.AsValueMap()
		for key, v := range srcMap {
			dstVal, ok := dstMap[key]
			if !ok && apply && ty.IsMapType() {
				// don't transfer old map values to dst during apply
				continue
			}

			if dstVal == cty.NilVal {
				if !apply && ty.IsMapType() {
					// let plan shape this map however it wants
					continue
				}
				dstVal = cty.NullVal(v.Type())
			}

			dstMap[key] = normalizeNullValues(dstVal, v, apply)
		}

		// you can't call MapVal/ObjectVal with empty maps, but nothing was
		// copied in anyway. If the dst is nil, and the src is known, assume the
		// src is correct.
		if len(dstMap) == 0 {
			if dst.IsNull() && src.IsWhollyKnown() && apply {
				return src
			}
			return dst
		}

		if ty.IsMapType() {
			// helper/schema will populate an optional+computed map with
			// unknowns which we have to fixup here.
			// It would be preferable to simply prevent any known value from
			// becoming unknown, but concessions have to be made to retain the
			// broken legacy behavior when possible.
			for k, srcVal := range srcMap {
				if !srcVal.IsNull() && srcVal.IsKnown() {
					dstVal, ok := dstMap[k]
					if !ok {
						continue
					}

					if !dstVal.IsNull() && !dstVal.IsKnown() {
						dstMap[k] = srcVal
					}
				}
			}

			return cty.MapVal(dstMap)
		}

		return cty.ObjectVal(dstMap)

	case ty.IsSetType():
		// If the original was wholly known, then we expect that is what the
		// provider applied. The apply process loses too much information to
		// reliably re-create the set.
		if src.IsWhollyKnown() && apply {
			return src
		}

	case ty.IsListType(), ty.IsTupleType():
		// If the dst is null, and the src is known, then we lost an empty value
		// so take the original.
		if dst.IsNull() {
			if src.IsWhollyKnown() && src.LengthInt() == 0 && apply {
				return src
			}

			// if dst is null and src only contains unknown values, then we lost
			// those during a read or plan.
			if !apply && !src.IsNull() {
				allUnknown := true
				for _, v := range src.AsValueSlice() {
					if v.IsKnown() {
						allUnknown = false
						break
					}
				}
				if allUnknown {
					return src
				}
			}

			return dst
		}

		// if the lengths are identical, then iterate over each element in succession.
		srcLen := src.LengthInt()
		dstLen := dst.LengthInt()
		if srcLen == dstLen && srcLen > 0 {
			srcs := src.AsValueSlice()
			dsts := dst.AsValueSlice()

			for i := 0; i < srcLen; i++ {
				dsts[i] = normalizeNullValues(dsts[i], srcs[i], apply)
			}

			if ty.IsTupleType() {
				return cty.TupleVal(dsts)
			}
			return cty.ListVal(dsts)
		}

	case ty == cty.String:
		// The legacy SDK should not be able to remove a value during plan or
		// apply, however we are only going to overwrite this if the source was
		// an empty string, since that is what is often equated with unset and
		// lost in the diff process.
		if dst.IsNull() && src.AsString() == "" {
			return src
		}
	}

	return dst
}

// validateConfigNulls checks a config value for unsupported nulls before
// attempting to shim the value. While null values can mostly be ignored in the
// configuration, since they're not supported in HCL1, the case where a null
// appears in a list-like attribute (list, set, tuple) will present a nil value
// to helper/schema which can panic. Return an error to the user in this case,
// indicating the attribute with the null value.
func validateConfigNulls(ctx context.Context, v cty.Value, path cty.Path) []*tfprotov5.Diagnostic {
	var diags []*tfprotov5.Diagnostic
	if v.IsNull() || !v.IsKnown() {
		return diags
	}

	switch {
	case v.Type().IsListType() || v.Type().IsSetType() || v.Type().IsTupleType():
		it := v.ElementIterator()
		for it.Next() {
			kv, ev := it.Element()
			if ev.IsNull() {
				// if this is a set, the kv is also going to be null which
				// isn't a valid path element, so we can't append it to the
				// diagnostic.
				p := path
				if !kv.IsNull() {
					p = append(p, cty.IndexStep{Key: kv})
				}

				diags = append(diags, &tfprotov5.Diagnostic{
					Severity:  tfprotov5.DiagnosticSeverityError,
					Summary:   "Null value found in list",
					Detail:    "Null values are not allowed for this attribute value.",
					Attribute: convert.PathToAttributePath(p),
				})
				continue
			}

			d := validateConfigNulls(ctx, ev, append(path, cty.IndexStep{Key: kv}))
			diags = convert.AppendProtoDiag(ctx, diags, d)
		}

	case v.Type().IsMapType() || v.Type().IsObjectType():
		it := v.ElementIterator()
		for it.Next() {
			kv, ev := it.Element()
			var step cty.PathStep
			switch {
			case v.Type().IsMapType():
				step = cty.IndexStep{Key: kv}
			case v.Type().IsObjectType():
				step = cty.GetAttrStep{Name: kv.AsString()}
			}
			d := validateConfigNulls(ctx, ev, append(path, step))
			diags = convert.AppendProtoDiag(ctx, diags, d)
		}
	}

	return diags
}

// Helper function that check a ConfigureProviderClientCapabilities struct to determine if a deferred response can be
// returned to the Terraform client. If no ConfigureProviderClientCapabilities have been passed from the client, then false
// is returned.
func configureDeferralAllowed(in *tfprotov5.ConfigureProviderClientCapabilities) bool {
	if in == nil {
		return false
	}

	return in.DeferralAllowed
}
