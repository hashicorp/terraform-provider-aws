// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package interceptors

import (
	"context"

	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// WithTaggingMethods should be embedded in tagging interceptors.
// It provides common methods for calling a service package's generic tagging methods.
type WithTaggingMethods struct {
	ServicePackageResourceTags *types.ServicePackageResourceTags
}

func (w WithTaggingMethods) HasServicePackageResourceTags() bool {
	return w.ServicePackageResourceTags != nil
}

// If the service package has a generic resource list tags methods, call it.
func (w WithTaggingMethods) ListTags(ctx context.Context, sp conns.ServicePackage, c *conns.AWSClient, identifier string) error {
	var err error

	if v, ok := sp.(tftags.ServiceTagLister); ok {
		err = v.ListTags(ctx, c, identifier) // Sets tags in Context
	} else if v, ok := sp.(tftags.ResourceTypeTagLister); ok {
		if w.ServicePackageResourceTags.ResourceType == "" {
			tflog.Error(ctx, "ListTags method requires ResourceType but none set", map[string]any{
				"ServicePackage": sp.ServicePackageName(),
			})
		} else {
			err = v.ListTags(ctx, c, identifier, w.ServicePackageResourceTags.ResourceType) // Sets tags in Context
		}
	} else {
		tflog.Warn(ctx, "No ListTags method found", map[string]any{
			"ServicePackage": sp.ServicePackageName(),
			"ResourceType":   w.ServicePackageResourceTags.ResourceType,
		})
	}

	switch {
	// ISO partitions may not support tagging, giving error.
	case errs.IsUnsupportedOperationInPartitionError(c.Partition(ctx), err):
		err = nil
	case sp.ServicePackageName() == names.DynamoDB && err != nil:
		// When a DynamoDB Table is `ARCHIVED`, ListTags returns `ResourceNotFoundException`.
		if tfresource.NotFound(err) || tfawserr.ErrMessageContains(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
			err = nil
		}
	}

	return err
}

// If the service package has a generic resource update tags methods, call it.
func (w WithTaggingMethods) UpdateTags(ctx context.Context, sp conns.ServicePackage, c *conns.AWSClient, identifier string, oldTags, newTags any) error {
	var err error

	if v, ok := sp.(tftags.ServiceTagUpdater); ok {
		err = v.UpdateTags(ctx, c, identifier, oldTags, newTags)
	} else if v, ok := sp.(tftags.ResourceTypeTagUpdater); ok {
		if w.ServicePackageResourceTags.ResourceType == "" {
			tflog.Error(ctx, "UpdateTags method requires ResourceType but none set", map[string]any{
				"ServicePackage": sp.ServicePackageName(),
			})
		} else {
			err = v.UpdateTags(ctx, c, identifier, w.ServicePackageResourceTags.ResourceType, oldTags, newTags)
		}
	} else {
		tflog.Warn(ctx, "No UpdateTags method found", map[string]any{
			"ServicePackage": sp.ServicePackageName(),
			"ResourceType":   w.ServicePackageResourceTags.ResourceType,
		})
	}

	// ISO partitions may not support tagging, giving error.
	if errs.IsUnsupportedOperationInPartitionError(c.Partition(ctx), err) {
		err = nil
	}

	return err
}
