// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"unique"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/interceptors"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type ListResourceWithSDKv2Tags struct {
	tagSpec interceptors.HTags
}

func (r *ListResourceWithSDKv2Tags) SetTagsSpec(tags unique.Handle[inttypes.ServicePackageResourceTags]) {
	r.tagSpec = interceptors.HTags(tags)
}

func (r *ListResourceWithSDKv2Tags) SetTags(ctx context.Context, client *conns.AWSClient, d *schema.ResourceData) error {
	sp, _, _, _, tagsInContext, ok := interceptors.InfoFromContext(ctx, client)
	if !ok {
		return nil
	}

	// If the R handler didn't set tags, try and read them from the service API.
	if tagsInContext.TagsOut.IsNone() {
		// Some old resources may not have the required attribute set after Read:
		// https://github.com/hashicorp/terraform-provider-aws/issues/31180
		if identifier := r.tagSpec.GetIdentifierSDKv2(ctx, d); identifier != "" {
			if err := r.tagSpec.ListTags(ctx, sp, client, identifier); err != nil {
				return err
			}
		}
	}

	// Remove any provider configured ignore_tags and system tags from those returned from the service API.
	tags := tagsInContext.TagsOut.UnwrapOrDefault().IgnoreSystem(sp.ServicePackageName()).IgnoreConfig(client.IgnoreTagsConfig(ctx))

	// The resource's configured tags can now include duplicate tags that have been configured on the provider.
	if err := d.Set(names.AttrTags, tags.ResolveDuplicates(ctx, client.DefaultTagsConfig(ctx), client.IgnoreTagsConfig(ctx), d, names.AttrTags, nil).Map()); err != nil {
		return err
	}

	// Computed tags_all do.
	if err := d.Set(names.AttrTagsAll, tags.Map()); err != nil {
		return err
	}

	// reset tags in context for next resource
	tagsInContext.TagsOut = nil

	return nil
}
