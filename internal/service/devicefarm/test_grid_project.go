// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devicefarm

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devicefarm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devicefarm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_devicefarm_test_grid_project", name="Test Grid Project")
// @Tags(identifierAttribute="arn")
func resourceTestGridProject() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTestGridProjectCreate,
		ReadWithoutTimeout:   resourceTestGridProjectRead,
		UpdateWithoutTimeout: resourceTestGridProjectUpdate,
		DeleteWithoutTimeout: resourceTestGridProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCConfig: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Required: true,
						},
						names.AttrVPCID: {
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
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &devicefarm.CreateTestGridProjectInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVPCConfig); ok {
		input.VpcConfig = expandTestGridProjectVPCConfig(v.([]interface{}))
	}

	output, err := conn.CreateTestGridProject(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DeviceFarm Test Grid Project (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.TestGridProject.Arn))

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting DeviceFarm Test Grid Project (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceTestGridProjectRead(ctx, d, meta)...)
}

func resourceTestGridProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	project, err := findTestGridProjectByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DeviceFarm Test Grid Project (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DeviceFarm Test Grid Project (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(project.Arn)
	d.Set(names.AttrName, project.Name)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, project.Description)
	if err := d.Set(names.AttrVPCConfig, flattenTestGridProjectVPCConfig(project.VpcConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
	}

	return diags
}

func resourceTestGridProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &devicefarm.UpdateTestGridProjectInput{
			ProjectArn: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		_, err := conn.UpdateTestGridProject(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DeviceFarm Test Grid Project (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTestGridProjectRead(ctx, d, meta)...)
}

func resourceTestGridProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeviceFarmClient(ctx)

	log.Printf("[DEBUG] Deleting DeviceFarm Test Grid Project: %s", d.Id())
	_, err := conn.DeleteTestGridProject(ctx, &devicefarm.DeleteTestGridProjectInput{
		ProjectArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DeviceFarm Test Grid Project (%s): %s", d.Id(), err)
	}

	return diags
}

func findTestGridProjectByARN(ctx context.Context, conn *devicefarm.Client, arn string) (*awstypes.TestGridProject, error) {
	input := &devicefarm.GetTestGridProjectInput{
		ProjectArn: aws.String(arn),
	}
	output, err := conn.GetTestGridProject(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TestGridProject == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TestGridProject, nil
}

func expandTestGridProjectVPCConfig(l []interface{}) *awstypes.TestGridVpcConfig {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.TestGridVpcConfig{
		VpcId:            aws.String(m[names.AttrVPCID].(string)),
		SubnetIds:        flex.ExpandStringValueSet(m[names.AttrSubnetIDs].(*schema.Set)),
		SecurityGroupIds: flex.ExpandStringValueSet(m[names.AttrSecurityGroupIDs].(*schema.Set)),
	}

	return config
}

func flattenTestGridProjectVPCConfig(conf *awstypes.TestGridVpcConfig) []interface{} {
	if conf == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrVPCID:            aws.ToString(conf.VpcId),
		names.AttrSubnetIDs:        flex.FlattenStringValueSet(conf.SubnetIds),
		names.AttrSecurityGroupIDs: flex.FlattenStringValueSet(conf.SecurityGroupIds),
	}

	return []interface{}{m}
}
