// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticache_subnet_group", name="Subnet Group")
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					// ElastiCache normalizes subnet names to lowercase,
					// so we have to do this too or else we can end up
					// with non-converging diffs.
					return strings.ToLower(val.(string))
				},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: resourceSubnetGroupDiff,
	}
}

func resourceSubnetGroupDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	// Reserved ElastiCache Subnet Groups with the name "default" do not support tagging;
	// thus we must suppress the diff originating from the provider-level default_tags configuration
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19213
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	if len(defaultTagsConfig.GetTags()) > 0 && diff.Get("name").(string) == "default" {
		return nil
	}

	return verify.SetTagsDiff(ctx, diff, meta)
}

func resourceSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	name := d.Get("name").(string)
	input := &elasticache.CreateCacheSubnetGroupInput{
		CacheSubnetGroupDescription: aws.String(d.Get("description").(string)),
		CacheSubnetGroupName:        aws.String(name),
		SubnetIds:                   flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:                        getTagsIn(ctx),
	}

	output, err := conn.CreateCacheSubnetGroupWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreateCacheSubnetGroupWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ElastiCache Subnet Group (%s): %s", name, err)
	}

	// Assign the group name as the resource ID
	// ElastiCache always retains the name in lower case, so we have to
	// mimic that or else we won't be able to refresh a resource whose
	// name contained uppercase characters.
	d.SetId(strings.ToLower(name))

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, aws.StringValue(output.CacheSubnetGroup.ARN), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ElastiCache Subnet Group (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	group, err := FindCacheSubnetGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Subnet Group (%s): %s", d.Id(), err)
	}

	var subnetIDs []*string
	for _, subnet := range group.Subnets {
		subnetIDs = append(subnetIDs, subnet.SubnetIdentifier)
	}

	d.Set("arn", group.ARN)
	d.Set("name", group.CacheSubnetGroupName)
	d.Set("description", group.CacheSubnetGroupDescription)
	d.Set("subnet_ids", aws.StringValueSlice(subnetIDs))

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	if d.HasChanges("subnet_ids", "description") {
		input := &elasticache.ModifyCacheSubnetGroupInput{
			CacheSubnetGroupDescription: aws.String(d.Get("description").(string)),
			CacheSubnetGroupName:        aws.String(d.Get("name").(string)),
			SubnetIds:                   flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		}

		_, err := conn.ModifyCacheSubnetGroupWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ElastiCache Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	log.Printf("[DEBUG] Deleting ElastiCache Subnet Group: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Minute, func() (interface{}, error) {
		return conn.DeleteCacheSubnetGroupWithContext(ctx, &elasticache.DeleteCacheSubnetGroupInput{
			CacheSubnetGroupName: aws.String(d.Id()),
		})
	}, "DependencyViolation")

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheSubnetGroupNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ElastiCache Subnet Group (%s): %s", d.Id(), err)
	}

	return diags
}
