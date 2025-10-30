// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudfront_connection_group", name="Connection Group")
// @Tags(identifierAttribute="arn")
func dataSourceConnectionGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectionGroupRead,

		Schema: map[string]*schema.Schema{
			"anycast_ip_list_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"etag": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"routing_endpoint", names.AttrID},
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipv6_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"routing_endpoint": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"routing_endpoint", names.AttrID},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceConnectionGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	var identifier string
	var connectionGroup *awstypes.ConnectionGroup
	var etag *string

	if id, ok := d.GetOk(names.AttrID); ok {
		identifier = id.(string)
		output, err := findConnectionGroupByID(ctx, conn, identifier)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudFront Connection Group by ID (%s): %s", identifier, err)
		}
		connectionGroup = output.ConnectionGroup
		etag = output.ETag
	} else {
		identifier = d.Get("routing_endpoint").(string)
		output, err := findConnectionGroupByEndpoint(ctx, conn, identifier)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudFront Connection Group by endpoint (%s): %s", identifier, err)
		}
		connectionGroup = output.ConnectionGroup
		etag = output.ETag
	}

	d.SetId(aws.ToString(connectionGroup.Id))
	d.Set("anycast_ip_list_id", connectionGroup.AnycastIpListId)
	d.Set(names.AttrARN, connectionGroup.Arn)
	d.Set(names.AttrEnabled, connectionGroup.Enabled)
	d.Set("etag", etag)
	d.Set("ipv6_enabled", connectionGroup.Ipv6Enabled)
	d.Set("is_default", connectionGroup.IsDefault)
	d.Set("last_modified_time", connectionGroup.LastModifiedTime.String())
	d.Set(names.AttrName, connectionGroup.Name)
	d.Set("routing_endpoint", connectionGroup.RoutingEndpoint)
	d.Set(names.AttrStatus, connectionGroup.Status)

	return diags
}

func findConnectionGroupByEndpoint(ctx context.Context, conn *cloudfront.Client, endpoint string) (*cloudfront.GetConnectionGroupByRoutingEndpointOutput, error) {
	input := cloudfront.GetConnectionGroupByRoutingEndpointInput{
		RoutingEndpoint: aws.String(endpoint),
	}

	output, err := conn.GetConnectionGroupByRoutingEndpoint(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ConnectionGroup == nil || output.ConnectionGroup.Name == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
