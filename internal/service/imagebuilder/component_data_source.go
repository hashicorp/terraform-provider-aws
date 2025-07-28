// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_imagebuilder_component", name="Component")
// @Tags
func dataSourceComponent() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceComponentRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"change_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"supported_os_versions": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceComponentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	arn := d.Get(names.AttrARN).(string)
	component, err := findComponentByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Component (%s): %s", arn, err)
	}

	arn = aws.ToString(component.Arn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set("change_description", component.ChangeDescription)
	d.Set("data", component.Data)
	d.Set("date_created", component.DateCreated)
	d.Set(names.AttrDescription, component.Description)
	d.Set(names.AttrEncrypted, component.Encrypted)
	d.Set(names.AttrKMSKeyID, component.KmsKeyId)
	d.Set(names.AttrName, component.Name)
	d.Set(names.AttrOwner, component.Owner)
	d.Set("platform", component.Platform)
	d.Set("supported_os_versions", component.SupportedOsVersions)
	d.Set(names.AttrType, component.Type)
	d.Set(names.AttrVersion, component.Version)

	setTagsOut(ctx, component.Tags)

	return diags
}
