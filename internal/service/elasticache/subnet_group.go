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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticache_subnet_group", name="Subnet Group")
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
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					// ElastiCache normalizes subnet names to lowercase,
					// so we have to do this too or else we can end up
					// with non-converging diffs.
					return strings.ToLower(val.(string))
				},
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.All(
			resourceSubnetGroupCustomizeDiff,
			verify.SetTagsDiff,
		),
	}
}

func resourceSubnetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &elasticache.CreateCacheSubnetGroupInput{
		CacheSubnetGroupDescription: aws.String(d.Get(names.AttrDescription).(string)),
		CacheSubnetGroupName:        aws.String(name),
		SubnetIds:                   flex.ExpandStringSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
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

	// Assign the group name as the resource ID.
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

	group, err := findCacheSubnetGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ElastiCache Subnet Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Subnet Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, group.ARN)
	d.Set(names.AttrDescription, group.CacheSubnetGroupDescription)
	d.Set(names.AttrName, group.CacheSubnetGroupName)
	d.Set(names.AttrSubnetIDs, tfslices.ApplyToAll(group.Subnets, func(v *elasticache.Subnet) string {
		return aws.StringValue(v.SubnetIdentifier)
	}))
	d.Set(names.AttrVPCID, group.VpcId)

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn(ctx)

	if d.HasChanges(names.AttrSubnetIDs, names.AttrDescription) {
		input := &elasticache.ModifyCacheSubnetGroupInput{
			CacheSubnetGroupDescription: aws.String(d.Get(names.AttrDescription).(string)),
			CacheSubnetGroupName:        aws.String(d.Get(names.AttrName).(string)),
			SubnetIds:                   flex.ExpandStringSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
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

func resourceSubnetGroupCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	// Reserved ElastiCache Subnet Groups with the name "default" do not support tagging,
	// thus we must suppress the diff originating from the provider-level default_tags configuration.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19213.
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	if len(defaultTagsConfig.GetTags()) > 0 && diff.Get(names.AttrName).(string) == "default" {
		return nil
	}

	return nil
}

func findCacheSubnetGroupByName(ctx context.Context, conn *elasticache.ElastiCache, name string) (*elasticache.CacheSubnetGroup, error) {
	input := &elasticache.DescribeCacheSubnetGroupsInput{
		CacheSubnetGroupName: aws.String(name),
	}

	return findCacheSubnetGroup(ctx, conn, input, tfslices.PredicateTrue[*elasticache.CacheSubnetGroup]())
}

func findCacheSubnetGroup(ctx context.Context, conn *elasticache.ElastiCache, input *elasticache.DescribeCacheSubnetGroupsInput, filter tfslices.Predicate[*elasticache.CacheSubnetGroup]) (*elasticache.CacheSubnetGroup, error) {
	output, err := findCacheSubnetGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findCacheSubnetGroups(ctx context.Context, conn *elasticache.ElastiCache, input *elasticache.DescribeCacheSubnetGroupsInput, filter tfslices.Predicate[*elasticache.CacheSubnetGroup]) ([]*elasticache.CacheSubnetGroup, error) {
	var output []*elasticache.CacheSubnetGroup

	err := conn.DescribeCacheSubnetGroupsPagesWithContext(ctx, input, func(page *elasticache.DescribeCacheSubnetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CacheSubnetGroups {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheSubnetGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
