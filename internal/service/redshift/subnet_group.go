// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_subnet_group", name="Subnet Group")
// @Tags(identifierAttribute="arn")
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

	subnetIdsSet := d.Get(names.AttrSubnetIDs).(*schema.Set)
	subnetIds := make([]string, subnetIdsSet.Len())
	for i, subnetId := range subnetIdsSet.List() {
		subnetIds[i] = subnetId.(string)
	}

	name := d.Get(names.AttrName).(string)
	input := redshift.CreateClusterSubnetGroupInput{
		ClusterSubnetGroupName: aws.String(name),
		Description:            aws.String(d.Get(names.AttrDescription).(string)),
		SubnetIds:              subnetIds,
		Tags:                   getTagsIn(ctx),
	}

	log.Printf("[DEBUG] Creating Redshift Subnet Group: %#v", input)
	_, err := conn.CreateClusterSubnetGroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Subnet Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(input.ClusterSubnetGroupName))

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	subnetgroup, err := findSubnetGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Subnet Group (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Service:   names.Redshift,
		Region:    meta.(*conns.AWSClient).Region(ctx),
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("subnetgroup:%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, subnetgroup.Description)
	d.Set(names.AttrName, d.Id())
	d.Set(names.AttrSubnetIDs, subnetIdsToSlice(subnetgroup.Subnets))

	setTagsOut(ctx, subnetgroup.Tags)

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	if d.HasChanges(names.AttrSubnetIDs, names.AttrDescription) {
		_, n := d.GetChange(names.AttrSubnetIDs)
		if n == nil {
			n = new(schema.Set)
		}
		ns := n.(*schema.Set)

		var sIds []string
		for _, s := range ns.List() {
			sIds = append(sIds, s.(string))
		}

		input := &redshift.ModifyClusterSubnetGroupInput{
			ClusterSubnetGroupName: aws.String(d.Id()),
			Description:            aws.String(d.Get(names.AttrDescription).(string)),
			SubnetIds:              sIds,
		}

		log.Printf("[DEBUG] Updating Redshift Subnet Group: %#v", input)
		_, err := conn.ModifyClusterSubnetGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Redshift Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift Subnet Group: %s", d.Id())
	_, err := conn.DeleteClusterSubnetGroup(ctx, &redshift.DeleteClusterSubnetGroupInput{
		ClusterSubnetGroupName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ClusterSubnetGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Subnet Group (%s): %s", d.Id(), err)
	}

	return diags
}

func subnetIdsToSlice(subnetIds []awstypes.Subnet) []string {
	subnetsSlice := make([]string, 0, len(subnetIds))
	for _, s := range subnetIds {
		subnetsSlice = append(subnetsSlice, aws.ToString(s.SubnetIdentifier))
	}
	return subnetsSlice
}
