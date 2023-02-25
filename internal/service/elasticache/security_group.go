package elasticache

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityGroupCreate,
		ReadWithoutTimeout:   resourceSecurityGroupRead,
		DeleteWithoutTimeout: resourceSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_names": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},

		DeprecationMessage: `With the retirement of EC2-Classic the aws_elasticache_security_group resource has been deprecated and will be removed in a future version.`,
	}
}

func resourceSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return sdkdiag.AppendErrorf(diags, `with the retirement of EC2-Classic no new ElastiCache Security Groups can be created`)
}

func resourceSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()
	req := &elasticache.DescribeCacheSecurityGroupsInput{
		CacheSecurityGroupName: aws.String(d.Id()),
	}

	res, err := conn.DescribeCacheSecurityGroupsWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Cache Security Group (%s): %s", d.Id(), err)
	}
	if len(res.CacheSecurityGroups) == 0 {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Cache Security Group (%s): empty response", d.Id())
	}

	var group *elasticache.CacheSecurityGroup
	for _, g := range res.CacheSecurityGroups {
		if aws.StringValue(g.CacheSecurityGroupName) == d.Id() {
			group = g
		}
	}
	if group == nil {
		return sdkdiag.AppendErrorf(diags, "reading ElastiCache Cache Security Group (%s): not found", d.Id())
	}

	d.Set("name", group.CacheSecurityGroupName)
	d.Set("description", group.Description)

	sgNames := make([]string, 0, len(group.EC2SecurityGroups))
	for _, sg := range group.EC2SecurityGroups {
		sgNames = append(sgNames, *sg.EC2SecurityGroupName)
	}
	d.Set("security_group_names", sgNames)

	return diags
}

func resourceSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElastiCacheConn()

	log.Printf("[DEBUG] Cache security group delete: %s", d.Id())

	err := resource.RetryContext(ctx, 5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteCacheSecurityGroupWithContext(ctx, &elasticache.DeleteCacheSecurityGroupInput{
			CacheSecurityGroupName: aws.String(d.Id()),
		})

		if tfawserr.ErrCodeEquals(err, "InvalidCacheSecurityGroupState", "DependencyViolation") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.RetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteCacheSecurityGroupWithContext(ctx, &elasticache.DeleteCacheSecurityGroupInput{
			CacheSecurityGroupName: aws.String(d.Id()),
		})
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ElastiCache Cache Security Group (%s): %s", d.Id(), err)
	}
	return diags
}
