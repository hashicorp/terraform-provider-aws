// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_docdb_subnet_group", name="Subnet Group")
// @Tags(identifierAttribute="arn")
func ResourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSubnetGroupCreate,
		ReadWithoutTimeout:   resourceSubnetGroupRead,
		UpdateWithoutTimeout: resourceSubnetGroupUpdate,
		DeleteWithoutTimeout: resourceSubnetGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validSubnetGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validSubnetGroupNamePrefix,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},

			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	subnetIds := flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set))

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = id.PrefixedUniqueId(v.(string))
	} else {
		groupName = id.UniqueId()
	}

	input := docdb.CreateDBSubnetGroupInput{
		DBSubnetGroupName:        aws.String(groupName),
		DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
		SubnetIds:                subnetIds,
		Tags:                     getTagsIn(ctx),
	}

	_, err := conn.CreateDBSubnetGroupWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DocumentDB Subnet Group: %s", err)
	}

	d.SetId(groupName)

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	describeOpts := docdb.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(d.Id()),
	}

	var subnetGroups []*docdb.DBSubnetGroup
	if err := conn.DescribeDBSubnetGroupsPagesWithContext(ctx, &describeOpts, func(resp *docdb.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
		subnetGroups = append(subnetGroups, resp.DBSubnetGroups...)
		return !lastPage
	}); err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBSubnetGroupNotFoundFault) {
			log.Printf("[WARN] DocumentDB Subnet Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Subnet Group (%s) parameters: %s", d.Id(), err)
	}

	if !d.IsNewResource() && (len(subnetGroups) != 1 || aws.StringValue(subnetGroups[0].DBSubnetGroupName) != d.Id()) {
		log.Printf("[WARN] DocumentDB Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	subnetGroup := subnetGroups[0]
	d.Set("name", subnetGroup.DBSubnetGroupName)
	d.Set("description", subnetGroup.DBSubnetGroupDescription)
	d.Set("arn", subnetGroup.DBSubnetGroupArn)

	subnets := make([]string, 0, len(subnetGroup.Subnets))
	for _, s := range subnetGroup.Subnets {
		subnets = append(subnets, aws.StringValue(s.SubnetIdentifier))
	}
	if err := d.Set("subnet_ids", subnets); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_ids: %s", err)
	}

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	if d.HasChanges("subnet_ids", "description") {
		_, n := d.GetChange("subnet_ids")
		if n == nil {
			n = new(schema.Set)
		}
		sIds := flex.ExpandStringSet(n.(*schema.Set))

		_, err := conn.ModifyDBSubnetGroupWithContext(ctx, &docdb.ModifyDBSubnetGroupInput{
			DBSubnetGroupName:        aws.String(d.Id()),
			DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
			SubnetIds:                sIds,
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying DocumentDB Subnet Group (%s) parameters: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	delOpts := docdb.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DocumentDB Subnet Group: %s", d.Id())

	_, err := conn.DeleteDBSubnetGroupWithContext(ctx, &delOpts)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBSubnetGroupNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Subnet Group (%s): %s", d.Id(), err)
	}

	if err := WaitForSubnetGroupDeletion(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Subnet Group (%s): %s", d.Id(), err)
	}
	return diags
}

func WaitForSubnetGroupDeletion(ctx context.Context, conn *docdb.DocDB, name string) error {
	params := &docdb.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(name),
	}

	err := retry.RetryContext(ctx, 10*time.Minute, func() *retry.RetryError {
		_, err := conn.DescribeDBSubnetGroupsWithContext(ctx, params)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBSubnetGroupNotFoundFault) {
			return nil
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return retry.RetryableError(fmt.Errorf("DocumentDB Subnet Group (%s) still exists", name))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DescribeDBSubnetGroupsWithContext(ctx, params)
		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBSubnetGroupNotFoundFault) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("deleting DocumentDB subnet group: %s", err)
	}
	return nil
}
