// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
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

// @SDKResource("aws_transfer_connector", name="Connector")
// @Tags(identifierAttribute="arn")
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
		Tags:       getTagsIn(ctx),
		Url:        aws.String(d.Get("url").(string)),
	}

	if v, ok := d.GetOk("as2_config"); ok {
		input.As2Config = expandAs2Config(v.([]interface{}))
	}

	if v, ok := d.GetOk("logging_role"); ok {
		input.LoggingRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_policy_name"); ok {
		input.SecurityPolicyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sftp_config"); ok {
		input.SftpConfig = expandSftpConfig(v.([]interface{}))
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
	d.Set(names.AttrARN, output.Arn)
	if err := d.Set("as2_config", flattenAs2Config(output.As2Config)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting as2_config: %s", err)
	}
	d.Set("connector_id", output.ConnectorId)
	d.Set("logging_role", output.LoggingRole)
	d.Set("security_policy_name", output.SecurityPolicyName)
	if err := d.Set("sftp_config", flattenSftpConfig(output.SftpConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sftp_config: %s", err)
	}
	d.Set("url", output.Url)
	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &transfer.UpdateConnectorInput{
			ConnectorId: aws.String(d.Id()),
		}

		if d.HasChange("access_role") {
			input.AccessRole = aws.String(d.Get("access_role").(string))
		}

		if d.HasChange("as2_config") {
			input.As2Config = expandAs2Config(d.Get("as2_config").([]interface{}))
		}

		if d.HasChange("logging_role") {
			input.LoggingRole = aws.String(d.Get("logging_role").(string))
		}

		if d.HasChange("security_policy_name") {
			input.SecurityPolicyName = aws.String(d.Get("security_policy_name").(string))
		}

		if d.HasChange("sftp_config") {
			input.SftpConfig = expandSftpConfig(d.Get("sftp_config").([]interface{}))
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

func expandAs2Config(pUser []interface{}) *transfer.As2ConnectorConfig {
	if len(pUser) < 1 || pUser[0] == nil {
		return nil
	}

	m := pUser[0].(map[string]interface{})

	as2Config := &transfer.As2ConnectorConfig{
		Compression:         aws.String(m["compression"].(string)),
		EncryptionAlgorithm: aws.String(m["encryption_algorithm"].(string)),
		LocalProfileId:      aws.String(m["local_profile_id"].(string)),
		MdnResponse:         aws.String(m["mdn_response"].(string)),
		MdnSigningAlgorithm: aws.String(m["mdn_signing_algorithm"].(string)),
		MessageSubject:      aws.String(m["message_subject"].(string)),
		PartnerProfileId:    aws.String(m["partner_profile_id"].(string)),
		SigningAlgorithm:    aws.String(m["signing_algorithm"].(string)),
	}

	return as2Config
}

func expandSftpConfig(pUser []interface{}) *transfer.SftpConnectorConfig {
	if len(pUser) < 1 || pUser[0] == nil {
		return nil
	}

	m := pUser[0].(map[string]interface{})

	sftpConfig := &transfer.SftpConnectorConfig{
		UserSecretId: aws.String(m["user_secret_id"].(string)),
	}

	if v, ok := m["trusted_host_keys"].(*schema.Set); ok && len(v.List()) > 0 {
		sftpConfig.TrustedHostKeys = flex.ExpandStringSet(v)
	}

	return sftpConfig
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

func flattenSftpConfig(posixUser *transfer.SftpConnectorConfig) []interface{} {
	if posixUser == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"trusted_host_keys": aws.StringValueSlice(posixUser.TrustedHostKeys),
		"user_secret_id":    aws.StringValue(posixUser.UserSecretId),
	}

	return []interface{}{m}
}
