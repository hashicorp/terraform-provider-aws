// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validSubnetGroupName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validSubnetGroupNamePrefix,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &docdb.CreateDBSubnetGroupInput{
		DBSubnetGroupDescription: aws.String(d.Get(names.AttrDescription).(string)),
		DBSubnetGroupName:        aws.String(name),
		SubnetIds:                flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
		Tags:                     getTagsIn(ctx),
	}

	_, err := conn.CreateDBSubnetGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DocumentDB Subnet Group (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	subnetGroup, err := findDBSubnetGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DocumentDB Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Subnet Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, subnetGroup.DBSubnetGroupArn)
	d.Set(names.AttrDescription, subnetGroup.DBSubnetGroupDescription)
	d.Set(names.AttrName, subnetGroup.DBSubnetGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(subnetGroup.DBSubnetGroupName)))
	var subnetIDs []string
	for _, v := range subnetGroup.Subnets {
		subnetIDs = append(subnetIDs, aws.ToString(v.SubnetIdentifier))
	}
	d.Set(names.AttrSubnetIDs, subnetIDs)

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	if d.HasChanges(names.AttrDescription, names.AttrSubnetIDs) {
		input := &docdb.ModifyDBSubnetGroupInput{
			DBSubnetGroupName:        aws.String(d.Id()),
			DBSubnetGroupDescription: aws.String(d.Get(names.AttrDescription).(string)),
			SubnetIds:                flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
		}

		_, err := conn.ModifyDBSubnetGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying DocumentDB Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	log.Printf("[DEBUG] Deleting DocumentDB Subnet Group: %s", d.Id())
	_, err := conn.DeleteDBSubnetGroup(ctx, &docdb.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.DBSubnetGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Subnet Group (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 10*time.Minute, func() (interface{}, error) {
		return findDBSubnetGroupByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Subnet Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findDBSubnetGroupByName(ctx context.Context, conn *docdb.Client, name string) (*awstypes.DBSubnetGroup, error) {
	input := &docdb.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(name),
	}
	output, err := findDBSubnetGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBSubnetGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBSubnetGroup(ctx context.Context, conn *docdb.Client, input *docdb.DescribeDBSubnetGroupsInput) (*awstypes.DBSubnetGroup, error) {
	output, err := findDBSubnetGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBSubnetGroups(ctx context.Context, conn *docdb.Client, input *docdb.DescribeDBSubnetGroupsInput) ([]awstypes.DBSubnetGroup, error) {
	var output []awstypes.DBSubnetGroup

	pages := docdb.NewDescribeDBSubnetGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.DBSubnetGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.DBSubnetGroups...)
	}

	return output, nil
}
