package resourcegroups

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_resourcegroups_resource", name="Resource")
func ResourceResource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceCreate,
		ReadWithoutTimeout:   resourceResourceRead,
		DeleteWithoutTimeout: resourceResourceDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"group_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

const (
	ResNameResource = "Resource"
)

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn(ctx)

	group := d.Get("group_arn").(string)
	resourceArn := d.Get("resource_arn").(string)

	in := &resourcegroups.GroupResourcesInput{
		Group:        aws.String(group),
		ResourceArns: []*string{&resourceArn},
	}

	_, err := conn.GroupResourcesWithContext(ctx, in)

	if err != nil {
		return create.DiagError(names.ResourceGroups, create.ErrActionCreating, ResNameResource, d.Get("name").(string), err)
	}

	vars := []string{
		strings.Split(strings.ToLower(d.Get("group_arn").(string)), "/")[1],
		strings.Split(d.Get("resource_arn").(string), "/")[1],
	}

	d.SetId(strings.Join(vars, "_"))

	return resourceResourceRead(ctx, d, meta)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn(ctx)

	out, err := FindResourceByARN(ctx, conn, d.Get("group_arn").(string), d.Get("resource_arn").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ResourceGroups Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.ResourceGroups, create.ErrActionReading, ResNameResource, d.Id(), err)
	}

	d.Set("resource_arn", out.Identifier.ResourceArn)
	d.Set("resource_type", out.Identifier.ResourceType)

	return nil
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn(ctx)

	log.Printf("[INFO] Deleting ResourceGroups Resource %s", d.Id())

	group := d.Get("group_arn").(string)
	resourceArn := d.Get("resource_arn").(string)

	_, err := conn.UngroupResourcesWithContext(ctx, &resourcegroups.UngroupResourcesInput{
		Group:        aws.String(group),
		ResourceArns: []*string{&resourceArn},
	})

	if err != nil {
		return create.DiagError(names.ResourceGroups, create.ErrActionDeleting, ResNameResource, d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return FindResourceByARN(ctx, conn, d.Get("group_arn").(string), d.Get("resource_arn").(string))
	})

	if err != nil {
		return create.DiagError(names.ResourceGroups, create.ErrActionDeleting, ResNameResource, d.Id(), err)
	}

	return nil
}

func FindResourceByARN(ctx context.Context, conn *resourcegroups.ResourceGroups, groupArn, resourceArn string) (*resourcegroups.ListGroupResourcesItem, error) {
	input := &resourcegroups.ListGroupResourcesInput{
		Group: aws.String(groupArn),
	}

	output, err := conn.ListGroupResourcesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, resourcegroups.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	for _, resourceItem := range output.Resources {
		if aws.StringValue(resourceItem.Identifier.ResourceArn) == resourceArn {
			return resourceItem, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}
