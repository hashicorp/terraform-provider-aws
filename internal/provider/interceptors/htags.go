// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interceptors

import (
	"context"

	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type taggingAWSClient interface {
	Partition(context.Context) string
}

type HTags inttypes.ServicePackageResourceTags

func (h HTags) unwrap() inttypes.ServicePackageResourceTags {
	return inttypes.ServicePackageResourceTags(h)
}

// GetIdentifierFramework returns the value of the identifier attribute used in AWS tagging APIs.
func (h HTags) GetIdentifierFramework(ctx context.Context, d interface {
	GetAttribute(context.Context, path.Path, any) diag.Diagnostics
}) string {
	var identifier string

	if identifierAttribute := h.unwrap().IdentifierAttribute(); identifierAttribute != "" {
		d.GetAttribute(ctx, path.Root(identifierAttribute), &identifier)
	}

	return identifier
}

// GetIdentifierSDKv2 returns the value of the identifier attribute used in AWS tagging APIs.
func (h HTags) GetIdentifierSDKv2(_ context.Context, d sdkv2.ResourceDiffer) string {
	var identifier string

	if identifierAttribute := h.unwrap().IdentifierAttribute(); identifierAttribute != "" {
		if identifierAttribute == names.AttrID {
			identifier = d.Id()
		} else {
			identifier = d.Get(identifierAttribute).(string)
		}
	}

	return identifier
}

func (h HTags) Enabled() bool {
	return h.unwrap().Enabled()
}

// If the service package has a generic resource list tags methods, call it.
func (h HTags) ListTags(ctx context.Context, sp conns.ServicePackage, c taggingAWSClient, identifier string) error {
	var err error

	resourceType := h.unwrap().ResourceType()
	if v, ok := sp.(tftags.ServiceTagLister); ok {
		err = v.ListTags(ctx, c, identifier) // Sets tags in Context
	} else if v, ok := sp.(tftags.ResourceTypeTagLister); ok {
		if resourceType == "" {
			tflog.Error(ctx, "ListTags method requires ResourceType but none set", map[string]any{
				"ServicePackage": sp.ServicePackageName(),
			})
		} else {
			err = v.ListTags(ctx, c, identifier, resourceType) // Sets tags in Context
		}
	} else {
		tflog.Warn(ctx, "No ListTags method found", map[string]any{
			"ServicePackage": sp.ServicePackageName(),
			"ResourceType":   resourceType,
		})
	}

	switch {
	// ISO partitions may not support tagging, giving error.
	case errs.IsUnsupportedOperationInPartitionError(c.Partition(ctx), err):
		err = nil
	case sp.ServicePackageName() == names.DynamoDB && err != nil:
		// When a DynamoDB Table is `ARCHIVED`, ListTags returns `ResourceNotFoundException`.
		if retry.NotFound(err) || tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
			err = nil
		}
	}

	return err
}

// If the service package has a generic resource update tags methods, call it.
func (h HTags) UpdateTags(ctx context.Context, sp conns.ServicePackage, c taggingAWSClient, identifier string, oldTags, newTags any) error {
	var err error

	resourceType := h.unwrap().ResourceType()
	if v, ok := sp.(tftags.ServiceTagUpdater); ok {
		err = v.UpdateTags(ctx, c, identifier, oldTags, newTags)
	} else if v, ok := sp.(tftags.ResourceTypeTagUpdater); ok {
		if resourceType == "" {
			tflog.Error(ctx, "UpdateTags method requires ResourceType but none set", map[string]any{
				"ServicePackage": sp.ServicePackageName(),
			})
		} else {
			err = v.UpdateTags(ctx, c, identifier, resourceType, oldTags, newTags)
		}
	} else {
		tflog.Warn(ctx, "No UpdateTags method found", map[string]any{
			"ServicePackage": sp.ServicePackageName(),
			"ResourceType":   resourceType,
		})
	}

	// ISO partitions may not support tagging, giving error.
	if errs.IsUnsupportedOperationInPartitionError(c.Partition(ctx), err) {
		err = nil
	}

	return err
}
