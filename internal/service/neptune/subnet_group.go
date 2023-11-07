// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_neptune_subnet_group", name="Subnet Group")
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
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
			"subnet_ids": {
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
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &neptune.CreateDBSubnetGroupInput{
		DBSubnetGroupName:        aws.String(name),
		DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
		SubnetIds:                flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:                     getTagsIn(ctx),
	}

	output, err := conn.CreateDBSubnetGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Subnet Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.DBSubnetGroup.DBSubnetGroupName))

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	subnetGroup, err := FindSubnetGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Neptune Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Subnet Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(subnetGroup.DBSubnetGroupArn)
	d.Set("arn", arn)
	d.Set("description", subnetGroup.DBSubnetGroupDescription)
	d.Set("name", subnetGroup.DBSubnetGroupName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(subnetGroup.DBSubnetGroupName)))
	var subnetIDs []string
	for _, v := range subnetGroup.Subnets {
		subnetIDs = append(subnetIDs, aws.StringValue(v.SubnetIdentifier))
	}
	d.Set("subnet_ids", subnetIDs)

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	if d.HasChanges("description", "subnet_ids") {
		input := &neptune.ModifyDBSubnetGroupInput{
			DBSubnetGroupName:        aws.String(d.Id()),
			DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
			SubnetIds:                flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		}

		_, err := conn.ModifyDBSubnetGroupWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Neptune Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	log.Printf("[DEBUG] Deleting Neptune Subnet Group: %s", d.Id())
	_, err := conn.DeleteDBSubnetGroupWithContext(ctx, &neptune.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBSubnetGroupNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Subnet Group (%s): %s", d.Id(), err)
	}

	return diags
}

func FindSubnetGroupByName(ctx context.Context, conn *neptune.Neptune, name string) (*neptune.DBSubnetGroup, error) {
	input := &neptune.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(name),
	}

	output, err := conn.DescribeDBSubnetGroupsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBSubnetGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DBSubnetGroups) == 0 || output.DBSubnetGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	dbSubnetGroup := output.DBSubnetGroups[0]

	// Eventual consistency check.
	if aws.StringValue(dbSubnetGroup.DBSubnetGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return dbSubnetGroup, nil
}
