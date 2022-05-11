package elasticache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceSubnetGroupCreate,
		Read:   resourceSubnetGroupRead,
		Update: resourceSubnetGroupUpdate,
		Delete: resourceSubnetGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
					// Elasticache normalizes subnet names to lowercase,
					// so we have to do this too or else we can end up
					// with non-converging diffs.
					return strings.ToLower(val.(string))
				},
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: resourceSubnetGroupDiff,
	}
}

func resourceSubnetGroupDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	// Reserved ElastiCache Subnet Groups with the name "default" do not support tagging;
	// thus we must suppress the diff originating from the provider-level default_tags configuration
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19213
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	if len(defaultTagsConfig.GetTags()) > 0 && diff.Get("name").(string) == "default" {
		return nil
	}

	return verify.SetTagsDiff(context.Background(), diff, meta)
}

func resourceSubnetGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// Get the group properties
	name := d.Get("name").(string)
	desc := d.Get("description").(string)
	subnetIdsSet := d.Get("subnet_ids").(*schema.Set)

	log.Printf("[DEBUG] Cache subnet group create: name: %s, description: %s", name, desc)

	subnetIds := flex.ExpandStringSet(subnetIdsSet)

	req := &elasticache.CreateCacheSubnetGroupInput{
		CacheSubnetGroupDescription: aws.String(desc),
		CacheSubnetGroupName:        aws.String(name),
		SubnetIds:                   subnetIds,
	}

	if len(tags) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateCacheSubnetGroup(req)

	if req.Tags != nil && verify.CheckISOErrorTagsUnsupported(err) {
		log.Printf("[WARN] failed creating ElastiCache Subnet Group with tags: %s. Trying create without tags.", err)

		req.Tags = nil
		output, err = conn.CreateCacheSubnetGroup(req)
	}

	if err != nil {
		return fmt.Errorf("creating ElastiCache Subnet Group (%s): %w", name, err)
	}

	// Assign the group name as the resource ID
	// Elasticache always retains the name in lower case, so we have to
	// mimic that or else we won't be able to refresh a resource whose
	// name contained uppercase characters.
	d.SetId(strings.ToLower(name))

	// In some partitions, only post-create tagging supported
	if req.Tags == nil && len(tags) > 0 {
		err := UpdateTags(conn, aws.StringValue(output.CacheSubnetGroup.ARN), nil, tags)

		if err != nil {
			if v, ok := d.GetOk("tags"); (ok && len(v.(map[string]interface{})) > 0) || !verify.CheckISOErrorTagsUnsupported(err) {
				// explicitly setting tags or not an iso-unsupported error
				return fmt.Errorf("failed adding tags after create for ElastiCache Subnet Group (%s): %w", d.Id(), err)
			}

			log.Printf("[WARN] failed adding tags after create for ElastiCache Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return resourceSubnetGroupRead(d, meta)
}

func resourceSubnetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &elasticache.DescribeCacheSubnetGroupsInput{
		CacheSubnetGroupName: aws.String(d.Get("name").(string)),
	}

	res, err := conn.DescribeCacheSubnetGroups(req)
	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "CacheSubnetGroupNotFoundFault" {
			// Update state to indicate the db subnet no longer exists.
			log.Printf("[WARN] Elasticache Subnet Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	if len(res.CacheSubnetGroups) == 0 {
		return fmt.Errorf("Error missing %v", d.Get("name"))
	}

	var group *elasticache.CacheSubnetGroup
	for _, g := range res.CacheSubnetGroups {
		log.Printf("[DEBUG] %v %v", g.CacheSubnetGroupName, d.Id())
		if aws.StringValue(g.CacheSubnetGroupName) == d.Id() {
			group = g
		}
	}
	if group == nil {
		return fmt.Errorf("Error retrieving cache subnet group: %v", res)
	}

	ids := make([]string, len(group.Subnets))
	for i, s := range group.Subnets {
		ids[i] = aws.StringValue(s.SubnetIdentifier)
	}

	d.Set("arn", group.ARN)
	d.Set("name", group.CacheSubnetGroupName)
	d.Set("description", group.CacheSubnetGroupDescription)
	d.Set("subnet_ids", ids)

	tags, err := ListTags(conn, d.Get("arn").(string))

	if err != nil && !verify.CheckISOErrorTagsUnsupported(err) {
		return fmt.Errorf("listing tags for ElastiCache Subnet Group (%s): %w", d.Id(), err)
	}

	// tags not supported in all partitions
	if err != nil {
		log.Printf("[WARN] failed listing tags for Elasticache Subnet Group (%s): %s", d.Id(), err)
	}

	if tags != nil {
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return fmt.Errorf("error setting tags: %w", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return fmt.Errorf("error setting tags_all: %w", err)
		}
	}

	return nil
}

func resourceSubnetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	if d.HasChanges("subnet_ids", "description") {
		var subnets []*string
		if v := d.Get("subnet_ids"); v != nil {
			for _, v := range v.(*schema.Set).List() {
				subnets = append(subnets, aws.String(v.(string)))
			}
		}
		log.Printf("[DEBUG] Updating ElastiCache Subnet Group")

		_, err := conn.ModifyCacheSubnetGroup(&elasticache.ModifyCacheSubnetGroupInput{
			CacheSubnetGroupName:        aws.String(d.Get("name").(string)),
			CacheSubnetGroupDescription: aws.String(d.Get("description").(string)),
			SubnetIds:                   subnets,
		})

		if err != nil {
			return fmt.Errorf("error updating ElastiCache Subnet Group (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, d.Get("arn").(string), o, n)
		if err != nil {
			if v, ok := d.GetOk("tags"); (ok && len(v.(map[string]interface{})) > 0) || !verify.CheckISOErrorTagsUnsupported(err) {
				// explicitly setting tags or not an iso-unsupported error
				return fmt.Errorf("failed updating ElastiCache Subnet Group (%s) tags: %w", d.Id(), err)
			}

			log.Printf("[WARN] failed updating tags for ElastiCache Subnet Group (%s): %s", d.Id(), err)
		}
	}

	return resourceSubnetGroupRead(d, meta)
}
func resourceSubnetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElastiCacheConn

	log.Printf("[DEBUG] Cache subnet group delete: %s", d.Id())

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteCacheSubnetGroup(&elasticache.DeleteCacheSubnetGroupInput{
			CacheSubnetGroupName: aws.String(d.Id()),
		})
		if err != nil {
			apierr, ok := err.(awserr.Error)
			if !ok {
				return resource.RetryableError(err)
			}
			log.Printf("[DEBUG] APIError.Code: %v", apierr.Code())
			switch apierr.Code() {
			case "DependencyViolation":
				// If it is a dependency violation, we want to retry
				return resource.RetryableError(err)
			default:
				return resource.NonRetryableError(err)
			}
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteCacheSubnetGroup(&elasticache.DeleteCacheSubnetGroupInput{
			CacheSubnetGroupName: aws.String(d.Id()),
		})
	}

	if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheSubnetGroupNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting ElastiCache Subnet Group (%s): %w", d.Id(), err)
	}

	return nil
}
