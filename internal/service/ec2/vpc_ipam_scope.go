// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_ipam_scope", name="IPAM Scope")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceIPAMScope() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPAMScopeCreate,
		ReadWithoutTimeout:   resourceIPAMScopeRead,
		UpdateWithoutTimeout: resourceIPAMScopeUpdate,
		DeleteWithoutTimeout: resourceIPAMScopeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
				},
				"external_authority_configuration": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"external_resource_identifier": {
								Type:     schema.TypeString,
								Required: true,
							},
							names.AttrType: {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.IpamScopeExternalAuthorityType](),
							},
						},
					},
				},
				"ipam_arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"ipam_id": {
					Type:     schema.TypeString,
					Required: true,
				},
				"ipam_scope_type": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"is_default": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				"pool_count": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},
	}
}

func resourceIPAMScopeCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateIpamScopeInput{
		ClientToken:       aws.String(create.UniqueId(ctx)),
		IpamId:            aws.String(d.Get("ipam_id").(string)),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeIpamScope),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("external_authority_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ExternalAuthorityConfiguration = expandIPAMScopeExternalAuthorityConfiguration(v.([]any)[0].(map[string]any))
	}

	output, err := conn.CreateIpamScope(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM Scope: %s", err)
	}

	d.SetId(aws.ToString(output.IpamScope.IpamScopeId))

	if _, err := waitIPAMScopeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Scope (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceIPAMScopeRead(ctx, d, meta)...)
}

func resourceIPAMScopeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	scope, err := findIPAMScopeByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] IPAM Scope (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Scope (%s): %s", d.Id(), err)
	}

	ipamID := strings.Split(aws.ToString(scope.IpamArn), "/")[1]
	d.Set(names.AttrARN, scope.IpamScopeArn)
	d.Set(names.AttrDescription, scope.Description)
	if v := scope.ExternalAuthorityConfiguration; v != nil {
		if err := d.Set("external_authority_configuration", []any{flattenIPAMScopeExternalAuthorityConfiguration(v)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting external_authority_configuration: %s", err)
		}
	} else {
		d.Set("external_authority_configuration", nil)
	}
	d.Set("ipam_arn", scope.IpamArn)
	d.Set("ipam_id", ipamID)
	d.Set("ipam_scope_type", scope.IpamScopeType)
	d.Set("is_default", scope.IsDefault)
	d.Set("pool_count", scope.PoolCount)

	setTagsOut(ctx, scope.Tags)

	return diags
}

func resourceIPAMScopeUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChanges(names.AttrDescription, "external_authority_configuration") {
		input := &ec2.ModifyIpamScopeInput{
			IpamScopeId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if d.HasChange("external_authority_configuration") {
			if v, ok := d.GetOk("external_authority_configuration"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.ExternalAuthorityConfiguration = expandIPAMScopeExternalAuthorityConfiguration(v.([]any)[0].(map[string]any))
			} else {
				input.RemoveExternalAuthorityConfiguration = aws.Bool(true)
			}
		}

		_, err := conn.ModifyIpamScope(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IPAM Scope (%s): %s", d.Id(), err)
		}

		if _, err := waitIPAMScopeUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IPAM Scope (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceIPAMScopeRead(ctx, d, meta)...)
}

func resourceIPAMScopeDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting IPAM Scope: %s", d.Id())
	input := ec2.DeleteIpamScopeInput{
		IpamScopeId: aws.String(d.Id()),
	}
	_, err := conn.DeleteIpamScope(ctx, &input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMScopeIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM Scope: (%s): %s", d.Id(), err)
	}

	if _, err := waitIPAMScopeDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Scope (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandIPAMScopeExternalAuthorityConfiguration(tfMap map[string]any) *awstypes.ExternalAuthorityConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ExternalAuthorityConfiguration{}

	if v, ok := tfMap["external_resource_identifier"].(string); ok && v != "" {
		apiObject.ExternalResourceIdentifier = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.IpamScopeExternalAuthorityType(v)
	}

	return apiObject
}

func flattenIPAMScopeExternalAuthorityConfiguration(apiObject *awstypes.IpamScopeExternalAuthorityConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrType: string(apiObject.Type),
	}

	if v := apiObject.ExternalResourceIdentifier; v != nil {
		tfMap["external_resource_identifier"] = aws.ToString(v)
	}

	return tfMap
}
