// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_emr_block_public_access_configuration", name="Block Public Access Configuration")
func resourceBlockPublicAccessConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBlockPublicAccessConfigurationCreate,
		ReadWithoutTimeout:   resourceBlockPublicAccessConfigurationRead,
		DeleteWithoutTimeout: resourceBlockPublicAccessConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"block_public_security_group_rules": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
			"permitted_public_security_group_rule_range": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_range": {
							Type:             schema.TypeInt,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsPortNumber),
						},
						"min_range": {
							Type:             schema.TypeInt,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsPortNumber),
						},
					},
				},
			},
		},
	}
}

func resourceBlockPublicAccessConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	blockPublicAccessConfiguration := &awstypes.BlockPublicAccessConfiguration{}

	blockPublicSecurityGroupRules := d.Get("block_public_security_group_rules")
	blockPublicAccessConfiguration.BlockPublicSecurityGroupRules = aws.Bool(blockPublicSecurityGroupRules.(bool))
	if v, ok := d.GetOk("permitted_public_security_group_rule_range"); ok {
		blockPublicAccessConfiguration.PermittedPublicSecurityGroupRuleRanges = expandPortRanges(v.([]any))
	}

	input := &emr.PutBlockPublicAccessConfigurationInput{
		BlockPublicAccessConfiguration: blockPublicAccessConfiguration,
	}

	_, err := conn.PutBlockPublicAccessConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting EMR Block Public Access Configuration: %s", err)
	}

	d.SetId("emr-block-public-access-configuration")

	return append(diags, resourceBlockPublicAccessConfigurationRead(ctx, d, meta)...)
}

func resourceBlockPublicAccessConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	bpa, err := findBlockPublicAccessConfiguration(ctx, conn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Block Public Access Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Block Public Access Configuration (%s): %s", d.Id(), err)
	}

	d.Set("block_public_security_group_rules", bpa.BlockPublicSecurityGroupRules)
	if err := d.Set("permitted_public_security_group_rule_range", flattenPortRanges(bpa.PermittedPublicSecurityGroupRuleRanges)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting permitted_public_security_group_rule_range: %s", err)
	}

	return diags
}

func resourceBlockPublicAccessConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	log.Printf("[DEBUG] Deleting EMR Block Public Access Configuration: %s", d.Id())
	_, err := conn.PutBlockPublicAccessConfiguration(ctx, &emr.PutBlockPublicAccessConfigurationInput{
		BlockPublicAccessConfiguration: defaultBlockPublicAccessConfiguration(),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "resetting EMR Block Public Access Configuration: %s", err)
	}

	return diags
}

func findBlockPublicAccessConfiguration(ctx context.Context, conn *emr.Client) (*awstypes.BlockPublicAccessConfiguration, error) {
	input := &emr.GetBlockPublicAccessConfigurationInput{}
	output, err := conn.GetBlockPublicAccessConfiguration(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.BlockPublicAccessConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.BlockPublicAccessConfiguration, nil
}

func defaultBlockPublicAccessConfiguration() *awstypes.BlockPublicAccessConfiguration {
	const port = 22

	return &awstypes.BlockPublicAccessConfiguration{
		BlockPublicSecurityGroupRules: aws.Bool(true),
		PermittedPublicSecurityGroupRuleRanges: []awstypes.PortRange{{
			MinRange: aws.Int32(port),
			MaxRange: aws.Int32(port),
		}},
	}
}

func flattenPortRange(apiObject *awstypes.PortRange) map[string]any {
	tfMap := map[string]any{}

	if v := apiObject.MinRange; v != nil {
		tfMap["min_range"] = aws.ToInt32(v)
	}

	if v := apiObject.MaxRange; v != nil {
		tfMap["max_range"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenPortRanges(apiObjects []awstypes.PortRange) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfMap []any

	for _, apiObject := range apiObjects {
		tfMap = append(tfMap, flattenPortRange(&apiObject))
	}

	return tfMap
}

func expandPortRange(tfMap map[string]any) *awstypes.PortRange {
	apiObject := &awstypes.PortRange{}

	apiObject.MinRange = aws.Int32(int32(tfMap["min_range"].(int)))
	apiObject.MaxRange = aws.Int32(int32(tfMap["max_range"].(int)))

	return apiObject
}

func expandPortRanges(tfList []any) []awstypes.PortRange {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.PortRange

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObjects = append(apiObjects, *expandPortRange(tfMap))
	}

	return apiObjects
}
