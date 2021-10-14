package aws

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
	tfeks "github.com/hashicorp/terraform-provider-aws/aws/internal/service/eks"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/eks/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/eks/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceIdentityProviderConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIdentityProviderConfigCreate,
		ReadWithoutTimeout:   resourceIdentityProviderConfigRead,
		UpdateWithoutTimeout: resourceIdentityProviderConfigUpdate,
		DeleteWithoutTimeout: resourceIdentityProviderConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: SetTagsDiff,

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
							ValidateDiagFunc: allDiagFunc(
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

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

// https://github.com/hashicorp/terraform-plugin-sdk/issues/780.
func allDiagFunc(validators ...schema.SchemaValidateDiagFunc) schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		for _, validator := range validators {
			diags = append(diags, validator(i, k)...)
		}
		return diags
	}
}

func resourceIdentityProviderConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	clusterName := d.Get("cluster_name").(string)
	configName, oidc := expandEksOidcIdentityProviderConfigRequest(d.Get("oidc").([]interface{})[0].(map[string]interface{}))
	id := tfeks.IdentityProviderConfigCreateResourceID(clusterName, configName)

	input := &eks.AssociateIdentityProviderConfigInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		ClusterName:        aws.String(clusterName),
		Oidc:               oidc,
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().EksTags()
	}

	_, err := conn.AssociateIdentityProviderConfig(input)

	if err != nil {
		return diag.Errorf("error associating EKS Identity Provider Config (%s): %s", id, err)
	}

	d.SetId(id)

	_, err = waiter.OidcIdentityProviderConfigCreated(ctx, conn, clusterName, configName, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return diag.Errorf("error waiting for EKS Identity Provider Config (%s) association: %s", d.Id(), err)
	}

	return resourceIdentityProviderConfigRead(ctx, d, meta)
}

func resourceIdentityProviderConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterName, configName, err := tfeks.IdentityProviderConfigParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	oidc, err := finder.OidcIdentityProviderConfigByClusterNameAndConfigName(ctx, conn, clusterName, configName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Identity Provider Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading EKS Identity Provider Config (%s): %s", d.Id(), err)
	}

	d.Set("arn", oidc.IdentityProviderConfigArn)
	d.Set("cluster_name", oidc.ClusterName)

	if err := d.Set("oidc", []interface{}{flattenEksOidcIdentityProviderConfig(oidc)}); err != nil {
		return diag.Errorf("error setting oidc: %s", err)
	}

	d.Set("status", oidc.Status)

	tags := keyvaluetags.EksKeyValueTags(oidc.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceIdentityProviderConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.EksUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

	return resourceIdentityProviderConfigRead(ctx, d, meta)
}

func resourceIdentityProviderConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSConn

	clusterName, configName, err := tfeks.IdentityProviderConfigParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Disassociating EKS Identity Provider Config: %s", d.Id())
	_, err = conn.DisassociateIdentityProviderConfigWithContext(ctx, &eks.DisassociateIdentityProviderConfigInput{
		ClusterName: aws.String(clusterName),
		IdentityProviderConfig: &eks.IdentityProviderConfig{
			Name: aws.String(configName),
			Type: aws.String(tfeks.IdentityProviderConfigTypeOidc),
		},
	})

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
		return nil
	}

	if tfawserr.ErrMessageContains(err, eks.ErrCodeInvalidRequestException, "Identity provider config is not associated with cluster") {
		return nil
	}

	if err != nil {
		return diag.Errorf("error disassociating EKS Identity Provider Config (%s): %s", d.Id(), err)
	}

	_, err = waiter.OidcIdentityProviderConfigDeleted(ctx, conn, clusterName, configName, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return diag.Errorf("error waiting for EKS Identity Provider Config (%s) disassociation: %s", d.Id(), err)
	}

	return nil
}

func expandEksOidcIdentityProviderConfigRequest(tfMap map[string]interface{}) (string, *eks.OidcIdentityProviderConfigRequest) {
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

func flattenEksOidcIdentityProviderConfig(apiObject *eks.OidcIdentityProviderConfig) map[string]interface{} {
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
