// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"unique"

	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	ephemeralschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	tfunique "github.com/hashicorp/terraform-provider-aws/internal/unique"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestProviderInit(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	primary, err := sdkv2.NewProvider(ctx)
	if err != nil {
		t.Fatalf("Initializing SDKv2 provider: %s", err)
	}

	p, err := NewProvider(ctx, primary)
	if err != nil {
		t.Fatalf("Initializing Framework provider: %s", err)
	}

	provider := p.(*frameworkProvider)

	validateResourceSchemas(ctx, t, provider)
}

// To run these benchmarks:
// go test -bench=. -benchmem -run=^$ -v ./internal/provider/framework

// This logs Initialization an annoying number of times
// func BenchmarkFrameworkProviderInitialization(b *testing.B) {
// 	ctx := b.Context()
// 	primary, err := sdkv2.NewProvider(ctx)
// 	if err != nil {
// 		b.Fatalf("Initializing SDKv2 provider: %s", err)
// 	}

// 	// Reset memory counters to zero, so that we only measure the Framework provider initialization.
// 	b.ResetTimer()
// 	for b.Loop() {
// 		_, err := NewProvider(ctx, primary)
// 		if err != nil {
// 			b.Fatalf("Initializing Framework provider: %s", err)
// 		}
// 	}
// }

func BenchmarkFrameworkProvider_SchemaInitialization_DataSource(b *testing.B) {
	ctx := b.Context()
	primary, err := sdkv2.NewProvider(ctx)
	if err != nil {
		b.Fatalf("Initializing SDKv2 provider: %s", err)
	}
	provider, err := NewProvider(ctx, primary)
	if err != nil {
		b.Fatalf("Initializing Framework provider: %s", err)
	}

	// Reset memory counters to zero, so that we only measure the Data Source schema initialization.
	b.ResetTimer()
	for b.Loop() {
		datasources := provider.DataSources(ctx)
		for _, f := range datasources {
			ds := f()
			ds.Schema(ctx, datasource.SchemaRequest{}, &datasource.SchemaResponse{})
		}
	}
}

func BenchmarkFrameworkProvider_SchemaInitialization_Resource(b *testing.B) {
	ctx := b.Context()
	primary, err := sdkv2.NewProvider(ctx)
	if err != nil {
		b.Fatalf("Initializing SDKv2 provider: %s", err)
	}
	provider, err := NewProvider(ctx, primary)
	if err != nil {
		b.Fatalf("Initializing Framework provider: %s", err)
	}

	// Reset memory counters to zero, so that we only measure the Resource schema initialization.
	b.ResetTimer()
	for b.Loop() {
		resources := provider.Resources(ctx)
		for _, f := range resources {
			r := f()
			r.Schema(ctx, resource.SchemaRequest{}, &resource.SchemaResponse{})
		}
	}
}

// validateResourceSchemas is called in a unit test to validate Terraform Plugin Framework-style resource schemas.
func validateResourceSchemas(ctx context.Context, t *testing.T, p *frameworkProvider) {
	t.Helper()

	for sp := range p.servicePackages {
		for _, dataSourceSpec := range sp.FrameworkDataSources(ctx) {
			typeName := dataSourceSpec.TypeName
			inner, err := dataSourceSpec.Factory(ctx)

			if err != nil {
				t.Errorf("creating data source type (%s): %s", typeName, err)
				continue
			}

			schemaResponse := datasource.SchemaResponse{}
			inner.Schema(ctx, datasource.SchemaRequest{}, &schemaResponse)

			if err := validateSchemaRegionForDataSource(dataSourceSpec.Region, schemaResponse.Schema); err != nil {
				t.Errorf("data source type %q: %s", typeName, err)
				continue
			}

			if err := validateSchemaTagsForDataSource(dataSourceSpec.Tags, schemaResponse.Schema); err != nil {
				t.Errorf("data source type %q: %s", typeName, err)
				continue
			}
		}

		if v, ok := sp.(conns.ServicePackageWithEphemeralResources); ok {
			for _, ephemeralResourceSpec := range v.EphemeralResources(ctx) {
				typeName := ephemeralResourceSpec.TypeName
				inner, err := ephemeralResourceSpec.Factory(ctx)

				if err != nil {
					t.Errorf("creating ephemeral resource type (%s): %s", typeName, err)
					continue
				}

				schemaResponse := ephemeral.SchemaResponse{}
				inner.Schema(ctx, ephemeral.SchemaRequest{}, &schemaResponse)

				if err := validateSchemaRegionForEphemeralResource(ephemeralResourceSpec.Region, schemaResponse.Schema); err != nil {
					t.Errorf("ephemeral resource type %q: %s", typeName, err)
					continue
				}
			}
		}

		if v, ok := sp.(conns.ServicePackageWithActions); ok {
			for _, actionSpec := range v.Actions(ctx) {
				typeName := actionSpec.TypeName
				inner, err := actionSpec.Factory(ctx)

				if err != nil {
					t.Errorf("creating action type (%s): %s", typeName, err)
					continue
				}

				schemaResponse := action.SchemaResponse{}
				inner.Schema(ctx, action.SchemaRequest{}, &schemaResponse)

				if err := validateSchemaRegionForAction(actionSpec.Region, schemaResponse.Schema); err != nil {
					t.Errorf("action type %q: %s", typeName, err)
					continue
				}
			}
		}

		for _, resourceSpec := range sp.FrameworkResources(ctx) {
			typeName := resourceSpec.TypeName
			inner, err := resourceSpec.Factory(ctx)

			if err != nil {
				t.Errorf("creating resource type (%s): %s", typeName, err)
				continue
			}

			schemaResponse := resource.SchemaResponse{}
			inner.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

			if err := validateSchemaRegionForResource(resourceSpec.Region, schemaResponse.Schema); err != nil {
				t.Errorf("resource type %q: %s", typeName, err)
				continue
			}

			if err := validateSchemaTagsForResource(resourceSpec.Tags, schemaResponse.Schema); err != nil {
				t.Errorf("resource type %q: %s", typeName, err)
				continue
			}

			if resourceSpec.Import.WrappedImport {
				if resourceSpec.Import.SetIDAttr {
					if _, ok := resourceSpec.Import.ImportID.(inttypes.FrameworkImportIDCreator); !ok {
						t.Errorf("resource type %q: importer sets `%s` attribute, but creator isn't configured", resourceSpec.TypeName, names.AttrID)
						continue
					}
				}

				if _, ok := inner.(framework.ImportByIdentityer); !ok {
					t.Errorf("resource type %q: cannot configure importer, does not implement %q", resourceSpec.TypeName, reflect.TypeFor[framework.ImportByIdentityer]())
					continue
				}
			}
		}
	}
}

func validateSchemaRegionForDataSource(regionSpec unique.Handle[inttypes.ServicePackageResourceRegion], schema datasourceschema.Schema) error {
	if !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		if _, ok := schema.Attributes[names.AttrRegion]; ok {
			return fmt.Errorf("configured for enhanced regions but defines `%s` attribute in schema", names.AttrRegion)
		}
	}
	return nil
}

func validateSchemaRegionForEphemeralResource(regionSpec unique.Handle[inttypes.ServicePackageResourceRegion], schema ephemeralschema.Schema) error {
	if !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		if _, ok := schema.Attributes[names.AttrRegion]; ok {
			return fmt.Errorf("configured for enhanced regions but defines `%s` attribute in schema", names.AttrRegion)
		}
	}
	return nil
}

func validateSchemaRegionForAction(regionSpec unique.Handle[inttypes.ServicePackageResourceRegion], schema actionschema.Schema) error {
	if !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		if _, ok := schema.Attributes[names.AttrRegion]; ok {
			return fmt.Errorf("configured for enhanced regions but defines `%s` attribute in schema", names.AttrRegion)
		}
	}
	return nil
}

func validateSchemaRegionForResource(regionSpec unique.Handle[inttypes.ServicePackageResourceRegion], schema resourceschema.Schema) error {
	if !tfunique.IsHandleNil(regionSpec) && regionSpec.Value().IsOverrideEnabled {
		if _, ok := schema.Attributes[names.AttrRegion]; ok {
			return fmt.Errorf("configured for enhanced regions but defines `%s` attribute in schema", names.AttrRegion)
		}
	}
	return nil
}

func validateSchemaTagsForDataSource(tagsSpec unique.Handle[inttypes.ServicePackageResourceTags], schema datasourceschema.Schema) error {
	if !tfunique.IsHandleNil(tagsSpec) {
		if v, ok := schema.Attributes[names.AttrTags]; ok {
			if !v.IsComputed() {
				return fmt.Errorf("`%s` attribute must be Computed", names.AttrTags)
			}
		} else {
			return fmt.Errorf("configured for tags but no `%s` attribute defined in schema", names.AttrTags)
		}
	}
	return nil
}

func validateSchemaTagsForResource(tagsSpec unique.Handle[inttypes.ServicePackageResourceTags], schema resourceschema.Schema) error {
	if !tfunique.IsHandleNil(tagsSpec) {
		if v, ok := schema.Attributes[names.AttrTags]; ok {
			if v.IsComputed() {
				return fmt.Errorf("`%s` attribute cannot be Computed", names.AttrTags)
			}
		} else {
			return fmt.Errorf("configured for tags but no `%s` attribute defined in schema", names.AttrTags)
		}
		if v, ok := schema.Attributes[names.AttrTagsAll]; ok {
			if !v.IsComputed() {
				return fmt.Errorf("`%s` attribute must be Computed", names.AttrTagsAll)
			}
		} else {
			return fmt.Errorf("configured for tags but no `%s` attribute defined in schema", names.AttrTagsAll)
		}
	}
	return nil
}
