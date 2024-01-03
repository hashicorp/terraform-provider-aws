// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devicefarm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/devicefarm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_devicefarm_test_grid_project", name="Test Grid Project")
// @Tags(identifierAttribute="arn")
func ResourceTestGridProject() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTestGridProjectCreate,
		ReadWithoutTimeout:   resourceTestGridProjectRead,
		UpdateWithoutTimeout: resourceTestGridProjectUpdate,
		DeleteWithoutTimeout: resourceTestGridProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTestGridProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn(ctx)

	name := d.Get("name").(string)
	input := &devicefarm.CreateTestGridProjectInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		input.VpcConfig = expandTestGridProjectVPCConfig(v.([]interface{}))
	}

	output, err := conn.CreateTestGridProjectWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DeviceFarm Test Grid Project (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.TestGridProject.Arn))

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting DeviceFarm Test Grid Project (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceTestGridProjectRead(ctx, d, meta)...)
}

func resourceTestGridProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn(ctx)

	project, err := FindTestGridProjectByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Test Grid Project (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DeviceFarm Test Grid Project (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(project.Arn)
	d.Set("name", project.Name)
	d.Set("arn", arn)
	d.Set("description", project.Description)
	if err := d.Set("vpc_config", flattenTestGridProjectVPCConfig(project.VpcConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
	}

	return diags
}

func resourceTestGridProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &devicefarm.UpdateTestGridProjectInput{
			ProjectArn: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		_, err := conn.UpdateTestGridProjectWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DeviceFarm Test Grid Project (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTestGridProjectRead(ctx, d, meta)...)
}

func resourceTestGridProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmConn(ctx)

	log.Printf("[DEBUG] Deleting DeviceFarm Test Grid Project: %s", d.Id())
	_, err := conn.DeleteTestGridProjectWithContext(ctx, &devicefarm.DeleteTestGridProjectInput{
		ProjectArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, devicefarm.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DeviceFarm Test Grid Project (%s): %s", d.Id(), err)
	}

	return diags
}

func expandTestGridProjectVPCConfig(l []interface{}) *devicefarm.TestGridVpcConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &devicefarm.TestGridVpcConfig{
		VpcId:            aws.String(m["vpc_id"].(string)),
		SubnetIds:        flex.ExpandStringSet(m["subnet_ids"].(*schema.Set)),
		SecurityGroupIds: flex.ExpandStringSet(m["security_group_ids"].(*schema.Set)),
	}

	return config
}

func flattenTestGridProjectVPCConfig(conf *devicefarm.TestGridVpcConfig) []interface{} {
	if conf == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"vpc_id":             aws.StringValue(conf.VpcId),
		"subnet_ids":         flex.FlattenStringSet(conf.SubnetIds),
		"security_group_ids": flex.FlattenStringSet(conf.SecurityGroupIds),
	}

	return []interface{}{m}
}
