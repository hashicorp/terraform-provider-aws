// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package transfer

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
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
			"egress_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_lattice": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port_number": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 65535),
									},
									"resource_configuration_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
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
				Optional: true,
			},
		},
	}
}

func resourceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	input := transfer.CreateConnectorInput{
		AccessRole: aws.String(d.Get("access_role").(string)),
		Tags:       getTagsIn(ctx),
	}

	if v, ok := d.GetOk("as2_config"); ok {
		input.As2Config = expandAs2ConnectorConfig(v.([]any))
	}

	if v, ok := d.GetOk("egress_config"); ok {
		input.EgressConfig = expandConnectorEgressConfig(v.([]any))
	}

	if v, ok := d.GetOk("logging_role"); ok {
		input.LoggingRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_policy_name"); ok {
		input.SecurityPolicyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sftp_config"); ok {
		input.SftpConfig = expandSftpConnectorConfig(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrURL); ok {
		input.Url = aws.String(v.(string))
	}

	output, err := conn.CreateConnector(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Connector: %s", err)
	}

	d.SetId(aws.ToString(output.ConnectorId))

	if _, err := waitConnectorCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Transfer Connector (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConnectorRead(ctx, d, meta)...)
}

func resourceConnectorRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	output, err := findConnectorByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
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
	if err := d.Set("egress_config", flattenDescribedConnectorEgressConfig(output.EgressConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting egress_config: %s", err)
	}
	d.Set("logging_role", output.LoggingRole)
	d.Set("security_policy_name", output.SecurityPolicyName)
	if err := d.Set("sftp_config", flattenSftpConnectorConfig(output.SftpConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sftp_config: %s", err)
	}
	d.Set(names.AttrURL, output.Url)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := transfer.UpdateConnectorInput{
			ConnectorId: aws.String(d.Id()),
		}

		if d.HasChange("access_role") {
			input.AccessRole = aws.String(d.Get("access_role").(string))
		}

		if d.HasChange("as2_config") {
			input.As2Config = expandAs2ConnectorConfig(d.Get("as2_config").([]any))
		}

		if d.HasChange("egress_config") {
			input.EgressConfig = expandUpdateConnectorEgressConfig(d.Get("egress_config").([]any))
		}

		if d.HasChange("logging_role") {
			input.LoggingRole = aws.String(d.Get("logging_role").(string))
		}

		if d.HasChange("security_policy_name") {
			input.SecurityPolicyName = aws.String(d.Get("security_policy_name").(string))
		}

		if d.HasChange("sftp_config") {
			input.SftpConfig = expandSftpConnectorConfig(d.Get("sftp_config").([]any))
		}

		if d.HasChange(names.AttrURL) {
			input.Url = aws.String(d.Get(names.AttrURL).(string))
		}

		_, err := conn.UpdateConnector(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Transfer Connector (%s): %s", d.Id(), err)
		}

		if _, err := waitConnectorUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Transfer Connector (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceConnectorRead(ctx, d, meta)...)
}

func resourceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferClient(ctx)

	log.Printf("[DEBUG] Deleting Transfer Connector: %s", d.Id())
	input := transfer.DeleteConnectorInput{
		ConnectorId: aws.String(d.Id()),
	}
	_, err := conn.DeleteConnector(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Transfer Connector (%s): %s", d.Id(), err)
	}

	if _, err := waitConnectorDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Transfer Connector (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findConnectorByID(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedConnector, error) {
	input := transfer.DescribeConnectorInput{
		ConnectorId: aws.String(id),
	}

	return findConnector(ctx, conn, &input)
}

func findConnector(ctx context.Context, conn *transfer.Client, input *transfer.DescribeConnectorInput) (*awstypes.DescribedConnector, error) {
	output, err := conn.DescribeConnector(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Connector == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Connector, nil
}

func expandAs2ConnectorConfig(tfList []any) *awstypes.As2ConnectorConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

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

func expandSftpConnectorConfig(tfList []any) *awstypes.SftpConnectorConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &awstypes.SftpConnectorConfig{
		UserSecretId: aws.String(tfMap["user_secret_id"].(string)),
	}

	if v, ok := tfMap["trusted_host_keys"].(*schema.Set); ok && len(v.List()) > 0 {
		apiObject.TrustedHostKeys = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func flattenAs2ConnectorConfig(apiObject *awstypes.As2ConnectorConfig) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
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

	return []any{tfMap}
}

func flattenSftpConnectorConfig(apiObject *awstypes.SftpConnectorConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"trusted_host_keys": apiObject.TrustedHostKeys,
		"user_secret_id":    aws.ToString(apiObject.UserSecretId),
	}

	return []any{tfMap}
}

func expandConnectorEgressConfig(tfList []any) awstypes.ConnectorEgressConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap["vpc_lattice"].([]any); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]any)

		apiObject := awstypes.ConnectorVpcLatticeEgressConfig{
			ResourceConfigurationArn: aws.String(tfMap["resource_configuration_arn"].(string)),
		}

		if v, ok := tfMap["port_number"].(int); ok && v > 0 {
			apiObject.PortNumber = aws.Int32(int32(v))
		}

		return &awstypes.ConnectorEgressConfigMemberVpcLattice{
			Value: apiObject,
		}
	}

	return nil
}

func expandUpdateConnectorEgressConfig(tfList []any) awstypes.UpdateConnectorEgressConfig {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap["vpc_lattice"].([]any); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]any)

		apiObject := awstypes.UpdateConnectorVpcLatticeEgressConfig{
			ResourceConfigurationArn: aws.String(tfMap["resource_configuration_arn"].(string)),
		}

		if v, ok := tfMap["port_number"].(int); ok && v > 0 {
			apiObject.PortNumber = aws.Int32(int32(v))
		}

		return &awstypes.UpdateConnectorEgressConfigMemberVpcLattice{
			Value: apiObject,
		}
	}

	return nil
}

func flattenDescribedConnectorEgressConfig(apiObject awstypes.DescribedConnectorEgressConfig) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	switch v := apiObject.(type) {
	case *awstypes.DescribedConnectorEgressConfigMemberVpcLattice:
		tfMap["vpc_lattice"] = flattenDescribedConnectorVPCLatticeEgressConfig(&v.Value)
	}

	if len(tfMap) == 0 {
		return nil
	}

	return []any{tfMap}
}

func flattenDescribedConnectorVPCLatticeEgressConfig(apiObject *awstypes.DescribedConnectorVpcLatticeEgressConfig) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.ResourceConfigurationArn; v != nil {
		tfMap["resource_configuration_arn"] = aws.ToString(v)
	}

	if v := apiObject.PortNumber; v != nil {
		tfMap["port_number"] = aws.ToInt32(v)
	}

	return []any{tfMap}
}

func statusConnector(conn *transfer.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findConnectorByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitConnectorCreated(ctx context.Context, conn *transfer.Client, id string, timeout time.Duration) (*awstypes.DescribedConnector, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectorStatusPending),
		Target:  enum.Slice(awstypes.ConnectorStatusActive),
		Refresh: statusConnector(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribedConnector); ok {
		if output.Status == awstypes.ConnectorStatusErrored {
			retry.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitConnectorUpdated(ctx context.Context, conn *transfer.Client, id string, timeout time.Duration) (*awstypes.DescribedConnector, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectorStatusPending),
		Target:  enum.Slice(awstypes.ConnectorStatusActive),
		Refresh: statusConnector(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribedConnector); ok {
		if output.Status == awstypes.ConnectorStatusErrored {
			retry.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitConnectorDeleted(ctx context.Context, conn *transfer.Client, id string, timeout time.Duration) (*awstypes.DescribedConnector, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectorStatusActive, awstypes.ConnectorStatusPending),
		Target:  []string{},
		Refresh: statusConnector(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribedConnector); ok {
		if output.Status == awstypes.ConnectorStatusErrored {
			retry.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}
