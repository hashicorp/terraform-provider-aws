package provider

import (
	"context"

	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func tagsUpdateFunc(ctx context.Context, d schemaResourceData, sp conns.ServicePackage, spt *types.ServicePackageResourceTags, serviceName, resourceName string, meta any, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	configTags := make(map[string]string)
	if config := d.GetRawConfig(); !config.IsNull() && config.IsKnown() {
		c := config.GetAttr(names.AttrTags)
		if !c.IsNull() {
			for k, v := range c.AsValueMap() {
				configTags[k] = v.AsString()
			}
		}
	}

	stateTags := make(map[string]string)
	if state := d.GetRawState(); !state.IsNull() && state.IsKnown() {
		s := state.GetAttr(names.AttrTagsAll)
		for k, v := range s.AsValueMap() {
			stateTags[k] = v.AsString()
		}
	}

	tagsAll := tftags.New(ctx, stateTags)
	// if tags_all was computed because not wholly known
	// Merge the resource's configured tags with any provider configured default_tags.
	configAll := tagsInContext.DefaultConfig.MergeTags(tftags.New(ctx, configTags))
	// Remove system tags.
	configAll = configAll.IgnoreSystem(inContext.ServicePackageName)

	toAdd := configAll.Difference(tagsAll)
	toRemove := tagsAll.Difference(configAll)

	var identifier string
	if identifierAttribute := spt.IdentifierAttribute; identifierAttribute == "id" {
		identifier = d.Id()
	} else {
		identifier = d.Get(identifierAttribute).(string)
	}
	// If the service package has a generic resource update tags methods, call it.
	var err error

	if v, ok := sp.(interface {
		UpdateTags(context.Context, any, string, any, any) error
	}); ok {
		err = v.UpdateTags(ctx, meta, identifier, toRemove, toAdd)
	} else if v, ok := sp.(interface {
		UpdateTags(context.Context, any, string, string, any, any) error
	}); ok && spt.ResourceType != "" {
		err = v.UpdateTags(ctx, meta, identifier, spt.ResourceType, toRemove, toAdd)
	}

	if verify.ErrorISOUnsupported(meta.(*conns.AWSClient).Partition, err) {
		// ISO partitions may not support tagging, giving error
		tflog.Warn(ctx, "failed updating tags for resource", map[string]interface{}{
			spt.IdentifierAttribute: identifier,
			"error":                 err.Error(),
		})

		return ctx, diags
	}

	if err != nil {
		return ctx, sdkdiag.AppendErrorf(diags, "updating tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
	}

	return ctx, diags
}

func tagsReadFunc(ctx context.Context, d schemaResourceData, sp conns.ServicePackage, spt *types.ServicePackageResourceTags, serviceName, resourceName string, meta any, diags diag.Diagnostics) (context.Context, diag.Diagnostics) {
	inContext, ok := conns.FromContext(ctx)
	if !ok {
		return ctx, diags
	}

	tagsInContext, ok := tftags.FromContext(ctx)
	if !ok {
		return ctx, diags
	}
	var identifier string
	if identifierAttribute := spt.IdentifierAttribute; identifierAttribute == "id" {
		identifier = d.Id()
	} else {
		identifier = d.Get(identifierAttribute).(string)
	}

	var err error
	if v, ok := sp.(interface {
		ListTags(context.Context, any, string) error
	}); ok {
		err = v.ListTags(ctx, meta, identifier) // Sets tags in Context
	} else if v, ok := sp.(interface {
		ListTags(context.Context, any, string, string) error
	}); ok && spt.ResourceType != "" {
		err = v.ListTags(ctx, meta, identifier, spt.ResourceType) // Sets tags in Context
	}

	if verify.ErrorISOUnsupported(meta.(*conns.AWSClient).Partition, err) {
		// ISO partitions may not support tagging, giving error
		tflog.Warn(ctx, "failed listing tags for resource", map[string]interface{}{
			spt.IdentifierAttribute: d.Id(),
			"error":                 err.Error(),
		})
		return ctx, diags
	}

	if inContext.ServicePackageName == names.DynamoDB && err != nil {
		// When a DynamoDB Table is `ARCHIVED`, ListTags returns `ResourceNotFoundException`.
		if tfresource.NotFound(err) || tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
			err = nil
		}
	}

	if err != nil {
		return ctx, sdkdiag.AppendErrorf(diags, "listing tags for %s %s (%s): %s", serviceName, resourceName, identifier, err)
	}

	// Remove any provider configured ignore_tags and system tags from those returned from the service API.
	toAdd := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(inContext.ServicePackageName).IgnoreConfig(tagsInContext.IgnoreConfig)

	// The resource's configured tags do not include any provider configured default_tags.
	if err := d.Set(names.AttrTags, toAdd.RemoveDefaultConfig(tagsInContext.DefaultConfig).RemoveDuplicates(ctx, tagsInContext.DefaultConfig, d).Map()); err != nil {
		return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTags, err)
	}

	// Computed tags_all do.
	if err := d.Set(names.AttrTagsAll, toAdd.Map()); err != nil {
		return ctx, sdkdiag.AppendErrorf(diags, "setting %s: %s", names.AttrTagsAll, err)
	}

	return ctx, diags
}
