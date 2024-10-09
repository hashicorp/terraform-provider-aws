// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amplify"
	"github.com/aws/aws-sdk-go-v2/service/amplify/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_amplify_domain_association", name="Domain Association")
func resourceDomainAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainAssociationCreate,
		ReadWithoutTimeout:   resourceDomainAssociationRead,
		UpdateWithoutTimeout: resourceDomainAssociationUpdate,
		DeleteWithoutTimeout: resourceDomainAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_verification_dns_record": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.CertificateType](),
						},
						"custom_certificate_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"certificate_verification_dns_record": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"enable_auto_sub_domain": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"sub_domain": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"branch_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"dns_record": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrPrefix: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"verified": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"wait_for_verification": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceDomainAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	appID := d.Get("app_id").(string)
	domainName := d.Get(names.AttrDomainName).(string)
	id := domainAssociationCreateResourceID(appID, domainName)
	input := &amplify.CreateDomainAssociationInput{
		AppId:               aws.String(appID),
		DomainName:          aws.String(domainName),
		EnableAutoSubDomain: aws.Bool(d.Get("enable_auto_sub_domain").(bool)),
		SubDomainSettings:   expandSubDomainSettings(d.Get("sub_domain").(*schema.Set).List()),
	}

	if v, ok := d.GetOk("certificate_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CertificateSettings = expandCertificateSettings(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.CreateDomainAssociation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Amplify Domain Association (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitDomainAssociationCreated(ctx, conn, appID, domainName); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Amplify Domain Association (%s) create: %s", d.Id(), err)
	}

	if d.Get("wait_for_verification").(bool) {
		if _, err := waitDomainAssociationVerified(ctx, conn, appID, domainName); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Amplify Domain Association (%s) verification: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainAssociationRead(ctx, d, meta)...)
}

func resourceDomainAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	appID, domainName, err := domainAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	domainAssociation, err := findDomainAssociationByTwoPartKey(ctx, conn, appID, domainName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Amplify Domain Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Amplify Domain Association (%s): %s", d.Id(), err)
	}

	d.Set("app_id", appID)
	d.Set(names.AttrARN, domainAssociation.DomainAssociationArn)
	d.Set("certificate_verification_dns_record", domainAssociation.CertificateVerificationDNSRecord)
	d.Set(names.AttrDomainName, domainAssociation.DomainName)
	d.Set("enable_auto_sub_domain", domainAssociation.EnableAutoSubDomain)
	if err := d.Set("sub_domain", flattenSubDomains(domainAssociation.SubDomains)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sub_domain: %s", err)
	}
	if err := d.Set("certificate_settings", flattenCertificateSettings(domainAssociation.Certificate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting certificate_settings: %s", err)
	}

	return diags
}

func resourceDomainAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	appID, domainName, err := domainAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges("enable_auto_sub_domain", "sub_domain") {
		input := &amplify.UpdateDomainAssociationInput{
			AppId:      aws.String(appID),
			DomainName: aws.String(domainName),
		}

		if d.HasChange("certificate_settings") {
			input.CertificateSettings = expandCertificateSettings(d.Get("certificate_settings").([]interface{})[0].(map[string]interface{}))
		}

		if d.HasChange("enable_auto_sub_domain") {
			input.EnableAutoSubDomain = aws.Bool(d.Get("enable_auto_sub_domain").(bool))
		}

		if d.HasChange("sub_domain") {
			input.SubDomainSettings = expandSubDomainSettings(d.Get("sub_domain").(*schema.Set).List())
		}

		_, err := conn.UpdateDomainAssociation(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Amplify Domain Association (%s): %s", d.Id(), err)
		}
	}

	if d.Get("wait_for_verification").(bool) {
		if _, err := waitDomainAssociationVerified(ctx, conn, appID, domainName); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Amplify Domain Association (%s) verification: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainAssociationRead(ctx, d, meta)...)
}

func resourceDomainAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	appID, domainName, err := domainAssociationParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Amplify Domain Association: %s", d.Id())
	_, err = conn.DeleteDomainAssociation(ctx, &amplify.DeleteDomainAssociationInput{
		AppId:      aws.String(appID),
		DomainName: aws.String(domainName),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Amplify Domain Association (%s): %s", d.Id(), err)
	}

	return diags
}

func findDomainAssociationByTwoPartKey(ctx context.Context, conn *amplify.Client, appID, domainName string) (*types.DomainAssociation, error) {
	input := &amplify.GetDomainAssociationInput{
		AppId:      aws.String(appID),
		DomainName: aws.String(domainName),
	}

	output, err := conn.GetDomainAssociation(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DomainAssociation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DomainAssociation, nil
}

func statusDomainAssociation(ctx context.Context, conn *amplify.Client, appID, domainName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		domainAssociation, err := findDomainAssociationByTwoPartKey(ctx, conn, appID, domainName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return domainAssociation, string(domainAssociation.DomainStatus), nil
	}
}

func waitDomainAssociationCreated(ctx context.Context, conn *amplify.Client, appID, domainName string) (*types.DomainAssociation, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DomainStatusCreating, types.DomainStatusInProgress, types.DomainStatusRequestingCertificate),
		Target:  enum.Slice(types.DomainStatusPendingVerification, types.DomainStatusPendingDeployment, types.DomainStatusAvailable),
		Refresh: statusDomainAssociation(ctx, conn, appID, domainName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*types.DomainAssociation); ok {
		if v.DomainStatus == types.DomainStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(v.StatusReason)))
		}

		return v, err
	}

	return nil, err
}

func waitDomainAssociationVerified(ctx context.Context, conn *amplify.Client, appID, domainName string) (*types.DomainAssociation, error) { //nolint:unparam
	const (
		timeout = 15 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DomainStatusUpdating, types.DomainStatusInProgress, types.DomainStatusPendingVerification),
		Target:  enum.Slice(types.DomainStatusPendingDeployment, types.DomainStatusAvailable),
		Refresh: statusDomainAssociation(ctx, conn, appID, domainName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*types.DomainAssociation); ok {
		if v.DomainStatus == types.DomainStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(v.StatusReason)))
		}

		return v, err
	}

	return nil, err
}

const domainAssociationResourceIDSeparator = "/"

func domainAssociationCreateResourceID(appID, domainName string) string {
	parts := []string{appID, domainName}
	id := strings.Join(parts, domainAssociationResourceIDSeparator)

	return id
}

func domainAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, domainAssociationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected APPID%[2]sDOMAINNAME", id, domainAssociationResourceIDSeparator)
}

func expandSubDomainSetting(tfMap map[string]interface{}) *types.SubDomainSetting {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.SubDomainSetting{}

	if v, ok := tfMap["branch_name"].(string); ok && v != "" {
		apiObject.BranchName = aws.String(v)
	}

	// Empty prefix is allowed.
	if v, ok := tfMap[names.AttrPrefix].(string); ok {
		apiObject.Prefix = aws.String(v)
	}

	return apiObject
}

func expandSubDomainSettings(tfList []interface{}) []types.SubDomainSetting {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.SubDomainSetting

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandSubDomainSetting(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandCertificateSettings(tfMap map[string]interface{}) *types.CertificateSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CertificateSettings{
		Type: types.CertificateType(tfMap[names.AttrType].(string)),
	}

	if v, ok := tfMap["custom_certificate_arn"].(string); ok {
		apiObject.CustomCertificateArn = aws.String(v)
	}

	return apiObject
}

func flattenCertificateSettings(apiObject *types.Certificate) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrType] = apiObject.Type

	if v := apiObject.CertificateVerificationDNSRecord; v != nil {
		tfMap["certificate_verification_dns_record"] = aws.ToString(v)
	}
	if v := apiObject.CustomCertificateArn; v != nil {
		tfMap["custom_certificate_arn"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenSubDomain(apiObject types.SubDomain) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.DnsRecord; v != nil {
		tfMap["dns_record"] = aws.ToString(v)
	}

	if v := apiObject.SubDomainSetting; v != nil {
		apiObject := v

		if v := apiObject.BranchName; v != nil {
			tfMap["branch_name"] = aws.ToString(v)
		}

		if v := apiObject.Prefix; v != nil {
			tfMap[names.AttrPrefix] = aws.ToString(v)
		}
	}

	if v := apiObject.Verified; v != nil {
		tfMap["verified"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenSubDomains(apiObjects []types.SubDomain) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenSubDomain(apiObject))
	}

	return tfList
}
