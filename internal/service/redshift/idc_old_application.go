// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_idc_application", name="IDC Application")
// @Tags(identifierAttribute="arn")
func resourceIdcApplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIdcApplicationCreate,
		ReadWithoutTimeout:   resourceIdcApplicationRead,
		UpdateWithoutTimeout: resourceIdcApplicationUpdate,
		DeleteWithoutTimeout: resourceIdcApplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(75 * time.Minute),
			Update: schema.DefaultTimeout(75 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorized_token_issuer_list": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authorized_audiences_list": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"trusted_token_issuer_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"iam_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"idc_display_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 127),
					validation.StringMatch(regexache.MustCompile(`[\w+=,.@-]+`), "must match [\\w+=,.@-]"),
				),
			},
			"idc_instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"identity_namespace": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 127),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9_+.#@$-]+$`), "must match ^[a-zA-Z0-9_+.#@$-]+$"),
				),
			},
			"redshift_idc_application_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`[a-z][a-z0-9]*(-[a-z0-9]+)*`), "must match [a-z][a-z0-9]*(-[a-z0-9]+)"),
				),
			},
			"service_integrations": {
				Optional: true,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lake_formation": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"lake_formation_query": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
				Set: serviceIntegrationsHash,
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceIdcApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	input := &redshift.CreateRedshiftIdcApplicationInput{
		RedshiftIdcApplicationName: aws.String(d.Get("redshift_idc_application_name").(string)),
		IamRoleArn:                 aws.String(d.Get("iam_role_arn").(string)),
		IdcInstanceArn:             aws.String(d.Get("idc_instance_arn").(string)),
		IdcDisplayName:             aws.String(d.Get("idc_display_name").(string)),
	}

	if v, ok := d.GetOk("authorized_token_issuer_list"); ok {
		input.AuthorizedTokenIssuerList = expandAuthorizedTokenIssuerList(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("service_integrations"); ok {
		input.ServiceIntegrations = expandServiceIntegrations(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("identity_namespace"); ok {
		input.IdentityNamespace = aws.String(v.(string))
	}

	log.Printf("[DEBUG] creating Redshift IDC Application: %s", input)
	output, err := conn.CreateRedshiftIdcApplicationWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift IDC Application (%s): %s", "", err)
	}
	d.SetId(aws.StringValue(output.RedshiftIdcApplication.RedshiftIdcApplicationArn))

	return append(diags, resourceIdcApplicationRead(ctx, d, meta)...)
}

func resourceIdcApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	rsIdc, err := findIDCApplicationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift IDC Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift IDC Application (%s): %s", d.Id(), err)
	}
	d.Set("authorized_token_issuer_list", flattenAuthorizedTokenIssuerList(rsIdc.AuthorizedTokenIssuerList))
	d.Set("iam_role_arn", rsIdc.IamRoleArn)
	d.Set("idc_display_name", rsIdc.IdcDisplayName)
	d.Set("idc_instance_arn", rsIdc.IdcInstanceArn)
	d.Set("identity_namespace", rsIdc.IdentityNamespace)
	d.Set("redshift_idc_application_name", rsIdc.RedshiftIdcApplicationName)
	d.Set("service_integrations", flatternServiceIntegrations(rsIdc.ServiceIntegrations))

	d.Set(names.AttrARN, rsIdc.RedshiftIdcApplicationArn)

	return diags
}

func resourceIdcApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	input := &redshift.ModifyRedshiftIdcApplicationInput{
		RedshiftIdcApplicationArn: aws.String(d.Id()),
	}

	if d.HasChange("authorized_token_issuer_list") {
		input.AuthorizedTokenIssuerList = expandAuthorizedTokenIssuerList(d.Get("authorized_token_issuer_list").(*schema.Set).List())
	}

	if d.HasChange("iam_role_arn") {
		input.IamRoleArn = aws.String(d.Get("iam_role_arn").(string))
	}

	if d.HasChange("idc_display_name") {
		input.IdcDisplayName = aws.String(d.Get("idc_display_name").(string))
	}

	if d.HasChange("identity_namespace") {
		input.IdentityNamespace = aws.String(d.Get("identity_namespace").(string))
	}

	if d.HasChange("service_integrations") {
		input.ServiceIntegrations = expandServiceIntegrations(d.Get("service_integrations").(*schema.Set).List())
	}

	_, err := conn.ModifyRedshiftIdcApplicationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Redshift IDC Application (%s): %s", d.Id(), err)
	}

	return append(diags, resourceIdcApplicationRead(ctx, d, meta)...)
}

func resourceIdcApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	log.Printf("[DEBUG] deleting Redshift IDC Application: %s", d.Id())
	_, err := conn.DeleteRedshiftIdcApplicationWithContext(ctx, &redshift.DeleteRedshiftIdcApplicationInput{
		RedshiftIdcApplicationArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeRedshiftIdcApplicationNotExistsFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift IDC Application (%s): %s", d.Id(), err)
	}

	return diags
}

func expandAuthorizedTokenIssuerList(vAuthorizedTokenIssuerList []interface{}) []*redshift.AuthorizedTokenIssuer {
	if len(vAuthorizedTokenIssuerList) == 0 || vAuthorizedTokenIssuerList[0] == nil {
		return nil
	}

	authorizedTokenIssuerList := []*redshift.AuthorizedTokenIssuer{}

	for _, v := range vAuthorizedTokenIssuerList {
		authorizedTokenIssuer := &redshift.AuthorizedTokenIssuer{}
		m := v.(map[string]interface{})
		if v, ok := m["authorized_audiences_list"].([]interface{}); ok && len(v) > 0 && v[0] != "" {
			authorizedTokenIssuer.AuthorizedAudiencesList = expandAuthorizedAudiences(v)
		}
		if vTrustedTokenIssuerArn, ok := m["trusted_token_issuer_arn"].(string); ok && vTrustedTokenIssuerArn != "" {
			authorizedTokenIssuer.TrustedTokenIssuerArn = aws.String(vTrustedTokenIssuerArn)
		}
		authorizedTokenIssuerList = append(authorizedTokenIssuerList, authorizedTokenIssuer)
	}
	return authorizedTokenIssuerList
}

func expandAuthorizedAudiences(v []interface{}) []*string {
	var authorizedAudiencesList []*string
	for _, v := range v {
		authorizedAudiencesList = append(authorizedAudiencesList, aws.String(v.(string)))
	}
	return authorizedAudiencesList
}

func flattenAuthorizedTokenIssuerList(v []*redshift.AuthorizedTokenIssuer) *schema.Set {
	s := &schema.Set{F: authorizedTokenIssuerListHash}
	if len(v) == 0 {
		return nil
	}

	for _, v := range v {
		var authorizedToeknIsuser interface{}
		authorizedAudiences := flatternAuthorizedAudiences(v.AuthorizedAudiencesList)
		authorizedToeknIsuser = map[string]interface{}{
			"authorized_audiences_list": authorizedAudiences,
			"trusted_token_issuer_arn":  aws.StringValue(v.TrustedTokenIssuerArn),
		}

		s.Add(authorizedToeknIsuser)
	}

	return s
}

func expandServiceIntegrations(v []interface{}) []*redshift.ServiceIntegrationsUnion {
	if len(v) == 0 || v[0] == nil {
		return nil
	}

	serviceIntegrations := []*redshift.ServiceIntegrationsUnion{}

	for _, v := range v {
		serviceIntegration := &redshift.ServiceIntegrationsUnion{}
		m := v.(map[string]interface{})
		if v, ok := m["lake_formation"].(*schema.Set); ok && v.Len() > 0 {
			serviceIntegration.LakeFormation = expandLakeFormation(v.List())
		}
		serviceIntegrations = append(serviceIntegrations, serviceIntegration)
	}
	return serviceIntegrations
}

func expandLakeFormation(v []interface{}) []*redshift.LakeFormationScopeUnion {
	if len(v) == 0 || v[0] == nil {
		return nil
	}

	lakeFormation := []*redshift.LakeFormationScopeUnion{}
	for _, v := range v {
		lakeFormationScopeUnion := expandLakeFormationScopeUnion(v.(map[string]interface{}))
		lakeFormation = append(lakeFormation, lakeFormationScopeUnion)
	}
	return lakeFormation
}

func expandLakeFormationScopeUnion(v map[string]interface{}) *redshift.LakeFormationScopeUnion {
	lakeFormationScopeUnion := &redshift.LakeFormationScopeUnion{}
	if v, ok := v["lake_formation_query"].(map[string]interface{}); ok {
		lakeFormationScopeUnion.LakeFormationQuery = expandLakeFormationQuery(v)
	}
	return lakeFormationScopeUnion
}

func expandLakeFormationQuery(v map[string]interface{}) *redshift.LakeFormationQuery {
	lakeFormationQuery := &redshift.LakeFormationQuery{}
	if v, ok := v["authorization"].(string); ok && v != "" {
		lakeFormationQuery.Authorization = aws.String(v)
	}
	return lakeFormationQuery
}

func flatternAuthorizedAudiences(v []*string) []interface{} {
	var authorizedAudiencesList []interface{}
	for _, v := range v {
		authorizedAudiencesList = append(authorizedAudiencesList, v)
	}
	return authorizedAudiencesList
}

func flatternServiceIntegrations(v []*redshift.ServiceIntegrationsUnion) *schema.Set {
	serviceIntegrations := []interface{}{}
	if len(v) == 0 {
		return nil
	}

	for _, v := range v {
		serviceIntegrationsUnion := flatternServiceIntegrationsUnion(v)
		serviceIntegrations = append(serviceIntegrations, serviceIntegrationsUnion)
	}
	return schema.NewSet(serviceIntegrationsHash, serviceIntegrations)
}

func flatternServiceIntegrationsUnion(v *redshift.ServiceIntegrationsUnion) map[string]interface{} {
	mServiceIntegrationsUnion := make(map[string]interface{})
	if lakeFormation := v.LakeFormation; lakeFormation != nil {
		mServiceIntegrationsUnion["lake_formation"] = flatternLakeFormation(v.LakeFormation)
	}
	return mServiceIntegrationsUnion
}

func flatternLakeFormation(v []*redshift.LakeFormationScopeUnion) []interface{} {
	lakeFormation := []interface{}{}
	if len(v) == 0 {
		return nil
	}
	for _, v := range v {
		lakeFormationScopeUnion := flatternLakeFormationScopeUnion(v)
		lakeFormation = append(lakeFormation, lakeFormationScopeUnion)
	}

	return lakeFormation
}

func flatternLakeFormationScopeUnion(v *redshift.LakeFormationScopeUnion) map[string]interface{} {
	mLakeFormationScopeUnion := make(map[string]interface{})
	if lakeFormationQuery := v.LakeFormationQuery; lakeFormationQuery != nil {
		mLakeFormationScopeUnion["lake_formation_query"] = flatternLakeFormationQuery(v.LakeFormationQuery)
	}
	return mLakeFormationScopeUnion
}

func flatternLakeFormationQuery(v *redshift.LakeFormationQuery) map[string]interface{} {
	lakeFormationQuery := make(map[string]interface{})
	lakeFormationQuery["authorization"] = aws.StringValue(v.Authorization)

	return lakeFormationQuery
}

func authorizedTokenIssuerListHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["authorized_audiences_list"].([]interface{})))
	buf.WriteString(fmt.Sprintf("%s-", m["trusted_token_issuer_arn"].(string)))

	return create.StringHashcode(buf.String())
}

func serviceIntegrationsHash(vServiceIntegrations interface{}) int {
	var buf bytes.Buffer

	if vLakeformation, ok := vServiceIntegrations.(map[string]interface{})["lake_formation"].([]map[string]interface{}); ok && len(vLakeformation) > 0 && vLakeformation[0] != nil {
		for _, v := range vLakeformation {
			if vLakeFormationQuery, ok := v["lake_formation_query"].(map[string]interface{}); ok && len(vLakeFormationQuery) > 0 {
				if v, ok := vLakeFormationQuery["authorization"].(string); ok && v != "" {
					buf.WriteString(fmt.Sprintf("%s-", v))
				}
			}
		}
	}

	return create.StringHashcode(buf.String())
}
