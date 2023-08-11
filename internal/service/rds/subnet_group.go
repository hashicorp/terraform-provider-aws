// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

// @SDKResource("aws_db_subnet_group", name="DB Subnet Group")
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
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"supported_network_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &rds.CreateDBSubnetGroupInput{
		DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
		DBSubnetGroupName:        aws.String(name),
		SubnetIds:                flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:                     getTagsIn(ctx),
	}

	log.Printf("[DEBUG] Creating RDS DB Subnet Group: %s", input)
	output, err := conn.CreateDBSubnetGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Subnet Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.DBSubnetGroup.DBSubnetGroupName))

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	v, err := FindDBSubnetGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Subnet Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(v.DBSubnetGroupArn)
	d.Set("arn", arn)
	d.Set("description", v.DBSubnetGroupDescription)
	d.Set("name", v.DBSubnetGroupName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(v.DBSubnetGroupName)))
	var subnetIDs []string
	for _, v := range v.Subnets {
		subnetIDs = append(subnetIDs, aws.StringValue(v.SubnetIdentifier))
	}
	d.Set("subnet_ids", subnetIDs)
	d.Set("supported_network_types", aws.StringValueSlice(v.SupportedNetworkTypes))
	d.Set("vpc_id", v.VpcId)

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	if d.HasChanges("description", "subnet_ids") {
		input := &rds.ModifyDBSubnetGroupInput{
			DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
			DBSubnetGroupName:        aws.String(d.Id()),
			SubnetIds:                flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		}

		log.Printf("[DEBUG] Modifying RDS DB Subnet Group: %s", input)
		_, err := conn.ModifyDBSubnetGroupWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS DB Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	log.Printf("[DEBUG] Deleting RDS DB Subnet Group: %s", d.Id())
	_, err := conn.DeleteDBSubnetGroupWithContext(ctx, &rds.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBSubnetGroupNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Subnet Group (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 3*time.Minute, func() (interface{}, error) {
		return FindDBSubnetGroupByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS Subnet Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}
