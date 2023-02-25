package inspector

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceGroupCreate,
		ReadWithoutTimeout:   resourceResourceGroupRead,
		DeleteWithoutTimeout: resourceResourceGroupDelete,

		Schema: map[string]*schema.Schema{
			"tags": {
				ForceNew: true,
				Required: true,
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceResourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorConn()

	req := &inspector.CreateResourceGroupInput{
		ResourceGroupTags: expandResourceGroupTags(d.Get("tags").(map[string]interface{})),
	}
	log.Printf("[DEBUG] Creating Inspector resource group: %#v", req)
	resp, err := conn.CreateResourceGroupWithContext(ctx, req)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Inspector resource group: %s", err)
	}

	d.SetId(aws.StringValue(resp.ResourceGroupArn))

	return append(diags, resourceResourceGroupRead(ctx, d, meta)...)
}

func resourceResourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InspectorConn()

	resp, err := conn.DescribeResourceGroupsWithContext(ctx, &inspector.DescribeResourceGroupsInput{
		ResourceGroupArns: aws.StringSlice([]string{d.Id()}),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Inspector resource group (%s): %s", d.Id(), err)
	}

	if len(resp.ResourceGroups) == 0 {
		if failedItem, ok := resp.FailedItems[d.Id()]; ok {
			failureCode := aws.StringValue(failedItem.FailureCode)
			if failureCode == inspector.FailedItemErrorCodeItemDoesNotExist {
				log.Printf("[WARN] Inspector resource group (%s) not found, removing from state", d.Id())
				d.SetId("")
				return diags
			}

			return sdkdiag.AppendErrorf(diags, "reading Inspector resource group (%s): %s", d.Id(), failureCode)
		}

		return sdkdiag.AppendErrorf(diags, "reading Inspector resource group (%s): %v", d.Id(), resp.FailedItems)
	}

	resourceGroup := resp.ResourceGroups[0]
	d.Set("arn", resourceGroup.Arn)

	//lintignore:AWSR002
	if err := d.Set("tags", flattenResourceGroupTags(resourceGroup.Tags)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func resourceResourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func expandResourceGroupTags(m map[string]interface{}) []*inspector.ResourceGroupTag {
	var result []*inspector.ResourceGroupTag

	for k, v := range m {
		result = append(result, &inspector.ResourceGroupTag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		})
	}

	return result
}

func flattenResourceGroupTags(tags []*inspector.ResourceGroupTag) map[string]interface{} {
	m := map[string]interface{}{}

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
	}

	return m
}
