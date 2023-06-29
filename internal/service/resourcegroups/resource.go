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
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn(ctx)

	groupARN := d.Get("group_arn").(string)
	resourceARN := d.Get("resource_arn").(string)
	id := strings.Join([]string{strings.Split(strings.ToLower(groupARN), "/")[1], strings.Split(resourceARN, "/")[1]}, "_")
	input := &resourcegroups.GroupResourcesInput{
		Group:        aws.String(groupARN),
		ResourceArns: aws.StringSlice([]string{resourceARN}),
	}

	_, err := conn.GroupResourcesWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Resource Groups Resource (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceResourceRead(ctx, d, meta)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn(ctx)

	output, err := FindResourceByTwoPartKey(ctx, conn, d.Get("group_arn").(string), d.Get("resource_arn").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ResourceGroups Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Resource Groups Resource (%s): %s", d.Id(), err)
	}

	d.Set("resource_arn", output.Identifier.ResourceArn)
	d.Set("resource_type", output.Identifier.ResourceType)

	return nil
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResourceGroupsConn(ctx)

	groupARN := d.Get("group_arn").(string)
	resourceARN := d.Get("resource_arn").(string)
	log.Printf("[INFO] Deleting Resource Groups Resource: %s", d.Id())
	_, err := conn.UngroupResourcesWithContext(ctx, &resourcegroups.UngroupResourcesInput{
		Group:        aws.String(groupARN),
		ResourceArns: aws.StringSlice([]string{resourceARN}),
	})

	if err != nil {
		return diag.Errorf("deleting Resource Groups Resource (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return FindResourceByTwoPartKey(ctx, conn, groupARN, resourceARN)
	})

	if err != nil {
		return diag.Errorf("waiting for Resource Groups Resource (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindResourceByTwoPartKey(ctx context.Context, conn *resourcegroups.ResourceGroups, groupARN, resourceARN string) (*resourcegroups.ListGroupResourcesItem, error) {
	input := &resourcegroups.ListGroupResourcesInput{
		Group: aws.String(groupARN),
	}
	var output []*resourcegroups.ListGroupResourcesItem

	err := conn.ListGroupResourcesPagesWithContext(ctx, input, func(page *resourcegroups.ListGroupResourcesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Resources...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, resourcegroups.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	output = slices.Filter(output, func(v *resourcegroups.ListGroupResourcesItem) bool {
		return v.Identifier != nil && aws.StringValue(v.Identifier.ResourceArn) == resourceARN
	})

	return tfresource.AssertSinglePtrResult(output)
}
