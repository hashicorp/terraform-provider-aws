// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package logs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudwatch_log_group", name="Log Group")
// @Tags(identifierAttribute="arn")
func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGroupRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreationTime: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"deletion_protection_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_group_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"retention_in_days": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	name := d.Get(names.AttrName).(string)
	logGroup, err := findLogGroupByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Log Group (%s): %s", name, err)
	}

	d.SetId(name)
	d.Set(names.AttrARN, trimLogGroupARNWildcardSuffix(aws.ToString(logGroup.Arn)))
	d.Set(names.AttrCreationTime, logGroup.CreationTime)
	d.Set("deletion_protection_enabled", logGroup.DeletionProtectionEnabled)
	d.Set(names.AttrKMSKeyID, logGroup.KmsKeyId)
	d.Set("log_group_class", logGroup.LogGroupClass)
	d.Set("retention_in_days", logGroup.RetentionInDays)

	return diags
}
