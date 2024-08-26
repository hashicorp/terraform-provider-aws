// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
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

	conn := meta.(*conns.AWSClient).EMRConn(ctx)

	blockPublicAccessConfiguration := &emr.BlockPublicAccessConfiguration{}

	blockPublicSecurityGroupRules := d.Get("block_public_security_group_rules")
	blockPublicAccessConfiguration.BlockPublicSecurityGroupRules = aws.Bool(blockPublicSecurityGroupRules.(bool))
	if v, ok := d.GetOk("permitted_public_security_group_rule_range"); ok {
		blockPublicAccessConfiguration.PermittedPublicSecurityGroupRuleRanges = expandPermittedPublicSecurityGroupRuleRanges(v.([]interface{}))
	}

	in := &emr.PutBlockPublicAccessConfigurationInput{
		BlockPublicAccessConfiguration: blockPublicAccessConfiguration,
	}

	_, err := conn.PutBlockPublicAccessConfigurationWithContext(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.EMR, create.ErrActionCreating, ResNameBlockPublicAccessConfiguration, d.Id(), err)
	}
	d.SetId(dummyIDBlockPublicAccessConfiguration)

	return append(diags, resourceBlockPublicAccessConfigurationRead(ctx, d, meta)...)
}

func resourceBlockPublicAccessConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EMRConn(ctx)

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

	conn := meta.(*conns.AWSClient).EMRConn(ctx)

	log.Print("[INFO] Restoring EMR Block Public Access Configuration to default settings")

	blockPublicAccessConfiguration := findDefaultBlockPublicAccessConfiguration()
	in := &emr.PutBlockPublicAccessConfigurationInput{
		BlockPublicAccessConfiguration: blockPublicAccessConfiguration,
	}

	_, err := conn.PutBlockPublicAccessConfigurationWithContext(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.EMR, create.ErrActionDeleting, ResNameBlockPublicAccessConfiguration, d.Id(), err)
	}

	return diags
}

func findBlockPublicAccessConfiguration(ctx context.Context, conn *emr.EMR) (*emr.GetBlockPublicAccessConfigurationOutput, error) {
	input := &emr.GetBlockPublicAccessConfigurationInput{}
	output, err := conn.GetBlockPublicAccessConfigurationWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil || output.BlockPublicAccessConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findDefaultBlockPublicAccessConfiguration() *emr.BlockPublicAccessConfiguration {
	defaultPort := int64(defaultPermittedPublicSecurityGroupRulePort)
	defaultPortPointer := &defaultPort
	portRange := &emr.PortRange{MinRange: defaultPortPointer, MaxRange: defaultPortPointer}
	permittedPublicSecurityGroupRuleRanges := []*emr.PortRange{portRange}
	blockPublicAccessConfiguration := &emr.BlockPublicAccessConfiguration{
		BlockPublicSecurityGroupRules:          aws.Bool(true),
		PermittedPublicSecurityGroupRuleRanges: permittedPublicSecurityGroupRuleRanges,
	}
	return blockPublicAccessConfiguration
}

func flattenPermittedPublicSecurityGroupRuleRange(apiObject *emr.PortRange) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.MinRange; v != nil {
		m["min_range"] = aws.Int64Value(v)
	}

	if v := apiObject.MaxRange; v != nil {
		m["max_range"] = aws.Int64Value(v)
	}

	return m
}

func flattenPermittedPublicSecurityGroupRuleRanges(apiObjects []*emr.PortRange) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		l = append(l, flattenPermittedPublicSecurityGroupRuleRange(apiObject))
	}

	return l
}

func expandPermittedPublicSecurityGroupRuleRange(tfMap map[string]interface{}) *emr.PortRange {
	a := &emr.PortRange{}

	a.MinRange = aws.Int64(int64(tfMap["min_range"].(int)))

	a.MaxRange = aws.Int64(int64(tfMap["max_range"].(int)))

	return a
}

func expandPermittedPublicSecurityGroupRuleRanges(tfList []interface{}) []*emr.PortRange {
	if len(tfList) == 0 {
		return nil
	}

	var s []*emr.PortRange

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandPermittedPublicSecurityGroupRuleRange(m)

		if a == nil {
			continue
		}

		s = append(s, a)
	}

	return s
}
