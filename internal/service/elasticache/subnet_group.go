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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
	conn := meta.(*conns.AWSClient).ElastiCacheConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &elasticache.CreateCacheSubnetGroupInput{
		CacheSubnetGroupDescription: aws.String(d.Get("description").(string)),
		CacheSubnetGroupName:        aws.String(name),
		SubnetIds:                   flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateCacheSubnetGroupWithContext(ctx, input)

	if input.Tags != nil && verify.ErrorISOUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating ElastiCache Subnet Group with tags: %s. Trying create without tags.", err)

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

	// In some partitions, only post-create tagging supported
	if input.Tags == nil && len(tags) > 0 {
		err := UpdateTags(ctx, conn, aws.StringValue(output.CacheSubnetGroup.ARN), nil, tags)

		if err != nil {
			if v, ok := d.GetOk("tags"); (ok && len(v.(map[string]interface{})) > 0) || !verify.ErrorISOUnsupported(conn.PartitionID, err) {
				// explicitly setting tags or not an iso-unsupported error
				return sdkdiag.AppendErrorf(diags, "adding tags after create for ElastiCache Subnet Group (%s): %s", d.Id(), err)
			}

			log.Printf("[WARN] failed adding tags after create for ElastiCache Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil && !verify.ErrorISOUnsupported(conn.PartitionID, err) {
		return sdkdiag.AppendErrorf(diags, "listing tags for ElastiCache Subnet Group (%s): %s", d.Id(), err)
	}

	// tags not supported in all partitions
	if err != nil {
		log.Printf("[WARN] failed listing tags for ElastiCache Subnet Group (%s): %s", d.Id(), err)
	}

	if tags != nil {
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
		}
	}

	return diags
}

func resourceSubnetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()

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

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n)

		if err != nil {
			if v, ok := d.GetOk("tags"); (ok && len(v.(map[string]interface{})) > 0) || !verify.ErrorISOUnsupported(conn.PartitionID, err) {
				// explicitly setting tags or not an iso-unsupported error
				return sdkdiag.AppendErrorf(diags, "updating ElastiCache Subnet Group (%s) tags: %s", d.Id(), err)
			}

			log.Printf("[WARN] failed updating tags for ElastiCache Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSubnetGroupRead(ctx, d, meta)...)
}

func resourceSubnetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()

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
