// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package redshift

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_subnet_group", name="Subnet Group")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/redshift/types;awstypes;awstypes.ClusterSubnetGroup")
func resourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSubnetGroupCreate,
		ReadWithoutTimeout:   resourceSubnetGroupRead,
		UpdateWithoutTimeout: resourceSubnetGroupUpdate,
		DeleteWithoutTimeout: resourceSubnetGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			names.AttrName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
					validation.StringNotInSlice([]string{"default"}, false),
				),
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := redshift.CreateClusterSubnetGroupInput{
		ClusterSubnetGroupName: aws.String(name),
		Description:            aws.String(d.Get(names.AttrDescription).(string)),
		SubnetIds:              flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
		Tags:                   getTagsIn(ctx),
	}

	_, err := conn.CreateClusterSubnetGroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Subnet Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(input.ClusterSubnetGroupName))

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.RedshiftClient(ctx)

	subnetgroup, err := findSubnetGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Redshift Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Subnet Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, subnetGroupARN(ctx, c, d.Id()))
	d.Set(names.AttrDescription, subnetgroup.Description)
	d.Set(names.AttrName, d.Id())
	d.Set(names.AttrSubnetIDs, tfslices.ApplyToAll(subnetgroup.Subnets, func(v awstypes.Subnet) string {
		return aws.ToString(v.SubnetIdentifier)
	}))

	setTagsOut(ctx, subnetgroup.Tags)

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	if d.HasChanges(names.AttrDescription, names.AttrSubnetIDs) {
		input := redshift.ModifyClusterSubnetGroupInput{
			ClusterSubnetGroupName: aws.String(d.Id()),
			Description:            aws.String(d.Get(names.AttrDescription).(string)),
			SubnetIds:              flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
		}

		_, err := conn.ModifyClusterSubnetGroup(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift Subnet Group: %s", d.Id())
	input := redshift.DeleteClusterSubnetGroupInput{
		ClusterSubnetGroupName: aws.String(d.Id()),
	}
	_, err := conn.DeleteClusterSubnetGroup(ctx, &input)

	if errs.IsA[*awstypes.ClusterSubnetGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Subnet Group (%s): %s", d.Id(), err)
	}

	return diags
}

func subnetGroupARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.RegionalARN(ctx, names.Redshift, "subnetgroup:"+id)
}
