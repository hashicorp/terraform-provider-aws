// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_connector", name="Connector")
// @Tags(identifierAttribute="arn")
func resourceConnector() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"as2_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compression": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.CompressionEnum](),
						},
						"encryption_algorithm": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.EncryptionAlg](),
						},
						"local_profile_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mdn_response": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.MdnResponse](),
						},
						"mdn_signing_algorithm": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.MdnSigningAlg](),
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SigningAlg](),
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
			"security_policy_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 100),
					validation.StringMatch(regexache.MustCompile(`^TransferSFTPConnectorSecurityPolicy-[A-Za-z0-9-]+$`), "must be in the format matching TransferSFTPConnectorSecurityPolicy-[A-Za-z0-9-]+"),
				),
			},
			"sftp_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"trusted_host_keys": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							MaxItems: 10,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 2028),
							},
						},
						"user_secret_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 2028),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrURL: {
				Type:     schema.TypeString,
				Required: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	input := &transfer.CreateConnectorInput{
		AccessRole: aws.String(d.Get("access_role").(string)),
		Tags:       getTagsIn(ctx),
		Url:        aws.String(d.Get(names.AttrURL).(string)),
	}

	if v, ok := d.GetOk("as2_config"); ok {
		input.As2Config = expandAs2ConnectorConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("logging_role"); ok {
		input.LoggingRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_policy_name"); ok {
		input.SecurityPolicyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sftp_config"); ok {
		input.SftpConfig = expandSftpConnectorConfig(v.([]interface{}))
	}

	output, err := conn.CreateConnector(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Connector: %s", err)
	}

	d.SetId(aws.ToString(output.ConnectorId))

	return append(diags, resourceConnectorRead(ctx, d, meta)...)
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	output, err := findConnectorByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transfer Connector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Transfer Connector (%s): %s", d.Id(), err)
	}

	d.Set("access_role", output.AccessRole)
	d.Set(names.AttrARN, output.Arn)
	if err := d.Set("as2_config", flattenAs2ConnectorConfig(output.As2Config)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting as2_config: %s", err)
	}
	d.Set("connector_id", output.ConnectorId)
	d.Set("logging_role", output.LoggingRole)
	d.Set("security_policy_name", output.SecurityPolicyName)
	if err := d.Set("sftp_config", flattenSftpConnectorConfig(output.SftpConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sftp_config: %s", err)
	}
	d.Set(names.AttrURL, output.Url)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &transfer.UpdateConnectorInput{
			ConnectorId: aws.String(d.Id()),
		}

		if d.HasChange("access_role") {
			input.AccessRole = aws.String(d.Get("access_role").(string))
		}

		if d.HasChange("as2_config") {
			input.As2Config = expandAs2ConnectorConfig(d.Get("as2_config").([]interface{}))
		}

		if d.HasChange("logging_role") {
			input.LoggingRole = aws.String(d.Get("logging_role").(string))
		}

		if d.HasChange("security_policy_name") {
			input.SecurityPolicyName = aws.String(d.Get("security_policy_name").(string))
		}

		if d.HasChange("sftp_config") {
			input.SftpConfig = expandSftpConnectorConfig(d.Get("sftp_config").([]interface{}))
		}

		if d.HasChange(names.AttrURL) {
			input.Url = aws.String(d.Get(names.AttrURL).(string))
		}

		_, err := conn.UpdateConnector(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Transfer Connector (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceConnectorRead(ctx, d, meta)...)
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	log.Printf("[DEBUG] Deleting Transfer Connector: %s", d.Id())
	_, err := conn.DeleteConnector(ctx, &transfer.DeleteConnectorInput{
		ConnectorId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Connector (%s): %s", d.Id(), err)
	}

	return diags
}

func findConnectorByID(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedConnector, error) {
	input := &transfer.DescribeConnectorInput{
		ConnectorId: aws.String(id),
	}

	output, err := conn.DescribeConnector(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Connector == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Connector, nil
}

func expandAs2ConnectorConfig(tfList []interface{}) *awstypes.As2ConnectorConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &awstypes.As2ConnectorConfig{
		Compression:         awstypes.CompressionEnum(tfMap["compression"].(string)),
		EncryptionAlgorithm: awstypes.EncryptionAlg(tfMap["encryption_algorithm"].(string)),
		LocalProfileId:      aws.String(tfMap["local_profile_id"].(string)),
		MdnResponse:         awstypes.MdnResponse(tfMap["mdn_response"].(string)),
		MdnSigningAlgorithm: awstypes.MdnSigningAlg(tfMap["mdn_signing_algorithm"].(string)),
		MessageSubject:      aws.String(tfMap["message_subject"].(string)),
		PartnerProfileId:    aws.String(tfMap["partner_profile_id"].(string)),
		SigningAlgorithm:    awstypes.SigningAlg(tfMap["signing_algorithm"].(string)),
	}

	return apiObject
}

func expandSftpConnectorConfig(tfList []interface{}) *awstypes.SftpConnectorConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &awstypes.SftpConnectorConfig{
		UserSecretId: aws.String(tfMap["user_secret_id"].(string)),
	}

	if v, ok := tfMap["trusted_host_keys"].(*schema.Set); ok && len(v.List()) > 0 {
		apiObject.TrustedHostKeys = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func flattenAs2ConnectorConfig(apiObject *awstypes.As2ConnectorConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"compression":           apiObject.Compression,
		"encryption_algorithm":  apiObject.EncryptionAlgorithm,
		"mdn_response":          apiObject.MdnResponse,
		"mdn_signing_algorithm": apiObject.MdnSigningAlgorithm,
		"signing_algorithm":     apiObject.SigningAlgorithm,
	}

	if v := apiObject.LocalProfileId; v != nil {
		tfMap["local_profile_id"] = aws.ToString(v)
	}

	if v := apiObject.MessageSubject; v != nil {
		tfMap["message_subject"] = aws.ToString(v)
	}

	if v := apiObject.PartnerProfileId; v != nil {
		tfMap["partner_profile_id"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenSftpConnectorConfig(apiObject *awstypes.SftpConnectorConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"trusted_host_keys": apiObject.TrustedHostKeys,
		"user_secret_id":    aws.ToString(apiObject.UserSecretId),
	}

	return []interface{}{tfMap}
}
