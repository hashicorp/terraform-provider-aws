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
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
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
						"min_range": {
							Type:             schema.TypeInt,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsPortNumber),
						},
						"max_range": {
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

const (
	ResNameBlockPublicAccessConfiguration       = "Block Public Access Configuration"
	dummyIDBlockPublicAccessConfiguration       = "emr-block-public-access-configuration"
	defaultPermittedPublicSecurityGroupRulePort = 22
)

func resourceBlockPublicAccessConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	blockPublicAccessConfiguration := &awstypes.BlockPublicAccessConfiguration{}

	blockPublicSecurityGroupRules := d.Get("block_public_security_group_rules")
	blockPublicAccessConfiguration.BlockPublicSecurityGroupRules = aws.Bool(blockPublicSecurityGroupRules.(bool))
	if v, ok := d.GetOk("permitted_public_security_group_rule_range"); ok {
		blockPublicAccessConfiguration.PermittedPublicSecurityGroupRuleRanges = expandPermittedPublicSecurityGroupRuleRanges(v.([]interface{}))
	}

	in := &emr.PutBlockPublicAccessConfigurationInput{
		BlockPublicAccessConfiguration: blockPublicAccessConfiguration,
	}

	_, err := conn.PutBlockPublicAccessConfiguration(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.EMR, create.ErrActionCreating, ResNameBlockPublicAccessConfiguration, d.Id(), err)
	}
	d.SetId(dummyIDBlockPublicAccessConfiguration)

	return append(diags, resourceBlockPublicAccessConfigurationRead(ctx, d, meta)...)
}

func resourceBlockPublicAccessConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	out, err := findBlockPublicAccessConfiguration(ctx, conn)

	if err != nil {
		return create.AppendDiagError(diags, names.EMR, create.ErrActionReading, ResNameBlockPublicAccessConfiguration, d.Id(), err)
	}

	d.Set("block_public_security_group_rules", out.BlockPublicAccessConfiguration.BlockPublicSecurityGroupRules)
	if err := d.Set("permitted_public_security_group_rule_range", flattenPermittedPublicSecurityGroupRuleRanges(out.BlockPublicAccessConfiguration.PermittedPublicSecurityGroupRuleRanges)); err != nil {
		return create.AppendDiagError(diags, names.EMR, create.ErrActionSetting, ResNameBlockPublicAccessConfiguration, d.Id(), err)
	}

	return diags
}

func resourceBlockPublicAccessConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	log.Print("[INFO] Restoring EMR Block Public Access Configuration to default settings")

	blockPublicAccessConfiguration := findDefaultBlockPublicAccessConfiguration()
	in := &emr.PutBlockPublicAccessConfigurationInput{
		BlockPublicAccessConfiguration: blockPublicAccessConfiguration,
	}

	_, err := conn.PutBlockPublicAccessConfiguration(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.EMR, create.ErrActionDeleting, ResNameBlockPublicAccessConfiguration, d.Id(), err)
	}

	return diags
}

func findBlockPublicAccessConfiguration(ctx context.Context, conn *emr.Client) (*emr.GetBlockPublicAccessConfigurationOutput, error) {
	input := &emr.GetBlockPublicAccessConfigurationInput{}
	output, err := conn.GetBlockPublicAccessConfiguration(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil || output.BlockPublicAccessConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findDefaultBlockPublicAccessConfiguration() *awstypes.BlockPublicAccessConfiguration {
	defaultPort := int32(defaultPermittedPublicSecurityGroupRulePort)
	defaultPortPointer := &defaultPort
	portRange := awstypes.PortRange{MinRange: defaultPortPointer, MaxRange: defaultPortPointer}
	permittedPublicSecurityGroupRuleRanges := []awstypes.PortRange{portRange}
	blockPublicAccessConfiguration := &awstypes.BlockPublicAccessConfiguration{
		BlockPublicSecurityGroupRules:          aws.Bool(true),
		PermittedPublicSecurityGroupRuleRanges: permittedPublicSecurityGroupRuleRanges,
	}
	return blockPublicAccessConfiguration
}

func flattenPermittedPublicSecurityGroupRuleRange(apiObject awstypes.PortRange) map[string]interface{} {
	m := map[string]interface{}{}

	if v := apiObject.MinRange; v != nil {
		m["min_range"] = aws.ToInt32(v)
	}

	if v := apiObject.MaxRange; v != nil {
		m["max_range"] = aws.ToInt32(v)
	}

	return m
}

func flattenPermittedPublicSecurityGroupRuleRanges(apiObjects []awstypes.PortRange) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenPermittedPublicSecurityGroupRuleRange(apiObject))
	}

	return l
}

func expandPermittedPublicSecurityGroupRuleRange(tfMap map[string]interface{}) awstypes.PortRange {
	a := awstypes.PortRange{}

	a.MinRange = aws.Int32(int32(tfMap["min_range"].(int)))

	a.MaxRange = aws.Int32(int32(tfMap["max_range"].(int)))

	return a
}

func expandPermittedPublicSecurityGroupRuleRanges(tfList []interface{}) []awstypes.PortRange {
	if len(tfList) == 0 {
		return nil
	}

	var s []awstypes.PortRange

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandPermittedPublicSecurityGroupRuleRange(m)

		s = append(s, a)
	}

	return s
}
