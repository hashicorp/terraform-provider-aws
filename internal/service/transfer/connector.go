// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_connector", name="Connector")
// @Tags(identifierAttribute="connector_id")
func ResourceConnector() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectorCreate,
		ReadWithoutTimeout:   resourceConnectorRead,
		UpdateWithoutTimeout: resourceConnectorUpdate,
		DeleteWithoutTimeout: resourceConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"as2_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compression": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(transfer.CompressionEnum_Values(), false),
						},
						"encryption_algorithm": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(transfer.EncryptionAlg_Values(), false),
						},
						"local_profile_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mdn_response": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(transfer.MdnResponse_Values(), false),
						},
						"mdn_signing_algorithm": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(transfer.MdnSigningAlg_Values(), false),
						},
						"message_subject": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"partner_profile_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"signing_algorithm": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(transfer.SigningAlg_Values(), false),
						},
					},
				},
			},
			"connector_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logging_role": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	input := &transfer.CreateConnectorInput{
		AccessRole: aws.String(d.Get("access_role").(string)),
		As2Config:  expandAs2Config(d.Get("as2_config").([]interface{})[0].(map[string]interface{})),
		Tags:       getTagsIn(ctx),
		Url:        aws.String(d.Get("url").(string)),
	}

	if v, ok := d.GetOk("logging_role"); ok {
		input.LoggingRole = aws.String(v.(string))
	}

	output, err := conn.CreateConnectorWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Connector: %s", err)
	}

	d.SetId(aws.StringValue(output.ConnectorId))

	return append(diags, resourceConnectorRead(ctx, d, meta)...)
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	output, err := FindConnectorByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Connector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Connector (%s): %s", d.Id(), err)
	}

	d.Set("access_role", output.AccessRole)
	if err := d.Set("as2_config", flattenAs2Config(output.As2Config)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting as2_config: %s", err)
	}
	d.Set("connector_id", output.ConnectorId)
	d.Set("logging_role", output.LoggingRole)
	d.Set("url", output.Url)
	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &transfer.UpdateConnectorInput{
			ConnectorId: aws.String(d.Id()),
		}

		if d.HasChange("access_role") {
			input.AccessRole = aws.String(d.Get("access_role").(string))
		}

		if d.HasChange("as2_config") {
			if v, ok := d.GetOk("as2_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.As2Config = expandAs2Config(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("logging_role") {
			input.LoggingRole = aws.String(d.Get("logging_role").(string))
		}

		if d.HasChange("url") {
			input.Url = aws.String(d.Get("url").(string))
		}

		_, err := conn.UpdateConnectorWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Transfer Connector (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceConnectorRead(ctx, d, meta)...)
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	log.Printf("[DEBUG] Deleting Transfer Connector: %s", d.Id())
	_, err := conn.DeleteConnectorWithContext(ctx, &transfer.DeleteConnectorInput{
		ConnectorId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Connector (%s): %s", d.Id(), err)
	}

	return diags
}

func expandAs2Config(tfMap map[string]interface{}) *transfer.As2ConnectorConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &transfer.As2ConnectorConfig{}

	if v, ok := tfMap["compression"].(string); ok && v != "" {
		apiObject.Compression = aws.String(v)
	}

	if v, ok := tfMap["encryption_algorithm"].(string); ok && v != "" {
		apiObject.EncryptionAlgorithm = aws.String(v)
	}

	if v, ok := tfMap["local_profile_id"].(string); ok && v != "" {
		apiObject.LocalProfileId = aws.String(v)
	}

	if v, ok := tfMap["mdn_response"].(string); ok && v != "" {
		apiObject.MdnResponse = aws.String(v)
	}

	if v, ok := tfMap["mdn_signing_algorithm"].(string); ok && v != "" {
		apiObject.MdnSigningAlgorithm = aws.String(v)
	}

	if v, ok := tfMap["message_subject"].(string); ok && v != "" {
		apiObject.MessageSubject = aws.String(v)
	}

	if v, ok := tfMap["partner_profile_id"].(string); ok && v != "" {
		apiObject.PartnerProfileId = aws.String(v)
	}

	if v, ok := tfMap["signing_algorithm"].(string); ok && v != "" {
		apiObject.SigningAlgorithm = aws.String(v)
	}

	return apiObject
}

func flattenAs2Config(apiObject *transfer.As2ConnectorConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Compression; v != nil {
		tfMap["compression"] = aws.StringValue(v)
	}

	if v := apiObject.EncryptionAlgorithm; v != nil {
		tfMap["encryption_algorithm"] = aws.StringValue(v)
	}

	if v := apiObject.LocalProfileId; v != nil {
		tfMap["local_profile_id"] = aws.StringValue(v)
	}

	if v := apiObject.MdnResponse; v != nil {
		tfMap["mdn_response"] = aws.StringValue(v)
	}

	if v := apiObject.MdnSigningAlgorithm; v != nil {
		tfMap["mdn_signing_algorithm"] = aws.StringValue(v)
	}

	if v := apiObject.MessageSubject; v != nil {
		tfMap["message_subject"] = aws.StringValue(v)
	}

	if v := apiObject.PartnerProfileId; v != nil {
		tfMap["partner_profile_id"] = aws.StringValue(v)
	}

	if v := apiObject.SigningAlgorithm; v != nil {
		tfMap["signing_algorithm"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}
