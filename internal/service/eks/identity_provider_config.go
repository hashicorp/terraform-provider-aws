// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_eks_identity_provider_config", name="Identity Provider Config")
// @Tags(identifierAttribute="arn")
func ResourceIdentityProviderConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIdentityProviderConfigCreate,
		ReadWithoutTimeout:   resourceIdentityProviderConfigRead,
		UpdateWithoutTimeout: resourceIdentityProviderConfigUpdate,
		DeleteWithoutTimeout: resourceIdentityProviderConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"oidc": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"groups_claim": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"groups_prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"identity_provider_config_name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"issuer_url": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
						},
						"required_claims": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							ValidateDiagFunc: verify.ValidAllDiag(
								validation.MapKeyLenBetween(1, 63),
								validation.MapValueLenBetween(1, 253),
							),
							Elem: &schema.Schema{Type: schema.TypeString},
						},
						"username_claim": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"username_prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.NoZeroValues,
						},
					},
				},
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceIdentityProviderConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn(ctx)

	clusterName := d.Get("cluster_name").(string)
	configName, oidc := expandOIDCIdentityProviderConfigRequest(d.Get("oidc").([]interface{})[0].(map[string]interface{}))
	idpID := IdentityProviderConfigCreateResourceID(clusterName, configName)
	input := &eks.AssociateIdentityProviderConfigInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		ClusterName:        aws.String(clusterName),
		Oidc:               oidc,
		Tags:               getTagsIn(ctx),
	}

	_, err := conn.AssociateIdentityProviderConfigWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("associating EKS Identity Provider Config (%s): %s", idpID, err)
	}

	d.SetId(idpID)

	_, err = waitOIDCIdentityProviderConfigCreated(ctx, conn, clusterName, configName, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.Errorf("waiting for EKS Identity Provider Config (%s) association: %s", d.Id(), err)
	}

	return resourceIdentityProviderConfigRead(ctx, d, meta)
}

func resourceIdentityProviderConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn(ctx)

	clusterName, configName, err := IdentityProviderConfigParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	oidc, err := FindOIDCIdentityProviderConfigByClusterNameAndConfigName(ctx, conn, clusterName, configName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Identity Provider Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading EKS Identity Provider Config (%s): %s", d.Id(), err)
	}

	d.Set("arn", oidc.IdentityProviderConfigArn)
	d.Set("cluster_name", oidc.ClusterName)

	if err := d.Set("oidc", []interface{}{flattenOIDCIdentityProviderConfig(oidc)}); err != nil {
		return diag.Errorf("setting oidc: %s", err)
	}

	d.Set("status", oidc.Status)

	setTagsOut(ctx, oidc.Tags)

	return nil
}

func resourceIdentityProviderConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceIdentityProviderConfigRead(ctx, d, meta)
}

func resourceIdentityProviderConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn(ctx)

	clusterName, configName, err := IdentityProviderConfigParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Disassociating EKS Identity Provider Config: %s", d.Id())
	_, err = conn.DisassociateIdentityProviderConfigWithContext(ctx, &eks.DisassociateIdentityProviderConfigInput{
		ClusterName: aws.String(clusterName),
		IdentityProviderConfig: &eks.IdentityProviderConfig{
			Name: aws.String(configName),
			Type: aws.String(IdentityProviderConfigTypeOIDC),
		},
	})

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
		return nil
	}

	if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidRequestException, "Identity provider config is not associated with cluster") {
		return nil
	}

	if err != nil {
		return diag.Errorf("disassociating EKS Identity Provider Config (%s): %s", d.Id(), err)
	}

	_, err = waitOIDCIdentityProviderConfigDeleted(ctx, conn, clusterName, configName, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return diag.Errorf("waiting for EKS Identity Provider Config (%s) disassociation: %s", d.Id(), err)
	}

	return nil
}

func expandOIDCIdentityProviderConfigRequest(tfMap map[string]interface{}) (string, *eks.OidcIdentityProviderConfigRequest) {
	if tfMap == nil {
		return "", nil
	}

	apiObject := &eks.OidcIdentityProviderConfigRequest{}

	if v, ok := tfMap["client_id"].(string); ok && v != "" {
		apiObject.ClientId = aws.String(v)
	}

	if v, ok := tfMap["groups_claim"].(string); ok && v != "" {
		apiObject.GroupsClaim = aws.String(v)
	}

	if v, ok := tfMap["groups_prefix"].(string); ok && v != "" {
		apiObject.GroupsPrefix = aws.String(v)
	}

	var identityProviderConfigName string
	if v, ok := tfMap["identity_provider_config_name"].(string); ok && v != "" {
		identityProviderConfigName = v
		apiObject.IdentityProviderConfigName = aws.String(v)
	}

	if v, ok := tfMap["issuer_url"].(string); ok && v != "" {
		apiObject.IssuerUrl = aws.String(v)
	}

	if v, ok := tfMap["required_claims"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.RequiredClaims = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["username_claim"].(string); ok && v != "" {
		apiObject.UsernameClaim = aws.String(v)
	}

	if v, ok := tfMap["username_prefix"].(string); ok && v != "" {
		apiObject.UsernamePrefix = aws.String(v)
	}

	return identityProviderConfigName, apiObject
}

func flattenOIDCIdentityProviderConfig(apiObject *eks.OidcIdentityProviderConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ClientId; v != nil {
		tfMap["client_id"] = aws.StringValue(v)
	}

	if v := apiObject.GroupsClaim; v != nil {
		tfMap["groups_claim"] = aws.StringValue(v)
	}

	if v := apiObject.GroupsPrefix; v != nil {
		tfMap["groups_prefix"] = aws.StringValue(v)
	}

	if v := apiObject.IdentityProviderConfigName; v != nil {
		tfMap["identity_provider_config_name"] = aws.StringValue(v)
	}

	if v := apiObject.IssuerUrl; v != nil {
		tfMap["issuer_url"] = aws.StringValue(v)
	}

	if v := apiObject.RequiredClaims; v != nil {
		tfMap["required_claims"] = aws.StringValueMap(v)
	}

	if v := apiObject.UsernameClaim; v != nil {
		tfMap["username_claim"] = aws.StringValue(v)
	}

	if v := apiObject.UsernamePrefix; v != nil {
		tfMap["username_prefix"] = aws.StringValue(v)
	}

	return tfMap
}
