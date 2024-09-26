// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
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
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)
	partition := meta.(*conns.AWSClient).Partition

	name := d.Get(names.AttrName).(string)
	input := &elasticache.CreateCacheSubnetGroupInput{
		CacheSubnetGroupDescription: aws.String(d.Get(names.AttrDescription).(string)),
		CacheSubnetGroupName:        aws.String(name),
		SubnetIds:                   flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
		Tags:                        getTagsIn(ctx),
	}

	output, err := conn.CreateCacheSubnetGroup(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreateCacheSubnetGroup(ctx, input)
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
		err := createTags(ctx, conn, aws.ToString(output.CacheSubnetGroup.ARN), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
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
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

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
	d.Set(names.AttrSubnetIDs, tfslices.ApplyToAll(group.Subnets, func(v awstypes.Subnet) string {
		return aws.ToString(v.SubnetIdentifier)
	}))
	d.Set(names.AttrVPCID, group.VpcId)

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	if d.HasChanges(names.AttrSubnetIDs, names.AttrDescription) {
		input := &elasticache.ModifyCacheSubnetGroupInput{
			CacheSubnetGroupDescription: aws.String(d.Get(names.AttrDescription).(string)),
			CacheSubnetGroupName:        aws.String(d.Get(names.AttrName).(string)),
			SubnetIds:                   flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
		}

		_, err := conn.ModifyCacheSubnetGroup(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ElastiCache Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	log.Printf("[DEBUG] Deleting ElastiCache Subnet Group: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 5*time.Minute, func() (interface{}, error) {
		return conn.DeleteCacheSubnetGroup(ctx, &elasticache.DeleteCacheSubnetGroupInput{
			CacheSubnetGroupName: aws.String(d.Id()),
		})
	}, "DependencyViolation")

	if errs.IsA[*awstypes.CacheSubnetGroupNotFoundFault](err) {
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

func findCacheSubnetGroupByName(ctx context.Context, conn *elasticache.Client, name string) (*awstypes.CacheSubnetGroup, error) {
	input := &elasticache.DescribeCacheSubnetGroupsInput{
		CacheSubnetGroupName: aws.String(name),
	}

	return findCacheSubnetGroup(ctx, conn, input, tfslices.PredicateTrue[*awstypes.CacheSubnetGroup]())
}

func findCacheSubnetGroup(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeCacheSubnetGroupsInput, filter tfslices.Predicate[*awstypes.CacheSubnetGroup]) (*awstypes.CacheSubnetGroup, error) {
	output, err := findCacheSubnetGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findCacheSubnetGroups(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeCacheSubnetGroupsInput, filter tfslices.Predicate[*awstypes.CacheSubnetGroup]) ([]awstypes.CacheSubnetGroup, error) {
	var output []awstypes.CacheSubnetGroup

	pages := elasticache.NewDescribeCacheSubnetGroupsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.CacheSubnetGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.CacheSubnetGroups {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
