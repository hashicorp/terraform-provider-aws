package xray

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_xray_group", name="Group")
// @Tags(identifierAttribute="arn")
func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"filter_expression": {
				Type:     schema.TypeString,
				Required: true,
			},
			"insights_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"insights_enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"notifications_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayConn()

	name := d.Get("group_name").(string)
	input := &xray.CreateGroupInput{
		GroupName:        aws.String(name),
		FilterExpression: aws.String(d.Get("filter_expression").(string)),
		Tags:             GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("insights_configuration"); ok {
		input.InsightsConfiguration = expandInsightsConfig(v.([]interface{}))
	}

	output, err := conn.CreateGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating XRay Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Group.GroupARN))

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayConn()

	input := &xray.GetGroupInput{
		GroupARN: aws.String(d.Id()),
	}

	group, err := conn.GetGroupWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, xray.ErrCodeInvalidRequestException, "Group not found") {
		log.Printf("[WARN] XRay Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading XRay Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(group.Group.GroupARN)
	d.Set("arn", arn)
	d.Set("group_name", group.Group.GroupName)
	d.Set("filter_expression", group.Group.FilterExpression)
	if err := d.Set("insights_configuration", flattenInsightsConfig(group.Group.InsightsConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting insights_configuration: %s", err)
	}

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &xray.UpdateGroupInput{GroupARN: aws.String(d.Id())}

		if v, ok := d.GetOk("filter_expression"); ok {
			input.FilterExpression = aws.String(v.(string))
		}

		if v, ok := d.GetOk("insights_configuration"); ok {
			input.InsightsConfiguration = expandInsightsConfig(v.([]interface{}))
		}

		_, err := conn.UpdateGroupWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating XRay Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).XRayConn()

	log.Printf("[INFO] Deleting XRay Group: %s", d.Id())
	_, err := conn.DeleteGroupWithContext(ctx, &xray.DeleteGroupInput{
		GroupARN: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting XRay Group (%s): %s", d.Id(), err)
	}

	return diags
}

func expandInsightsConfig(l []interface{}) *xray.InsightsConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	config := xray.InsightsConfiguration{}

	if v, ok := m["insights_enabled"]; ok {
		config.InsightsEnabled = aws.Bool(v.(bool))
	}
	if v, ok := m["notifications_enabled"]; ok {
		config.NotificationsEnabled = aws.Bool(v.(bool))
	}

	return &config
}

func flattenInsightsConfig(config *xray.InsightsConfiguration) []interface{} {
	if config == nil {
		return nil
	}

	m := map[string]interface{}{}

	if config.InsightsEnabled != nil {
		m["insights_enabled"] = config.InsightsEnabled
	}
	if config.NotificationsEnabled != nil {
		m["notifications_enabled"] = config.NotificationsEnabled
	}

	return []interface{}{m}
}
