package aws

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const typeOidc string = "oidc"

func resourceAwsEksIdentityProviderConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsEksIdentityProviderConfigCreate,
		ReadContext:   resourceAwsEksIdentityProviderConfigRead,
		DeleteContext: resourceAwsEksIdentityProviderConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
			Delete: schema.DefaultTimeout(25 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
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
							ValidateDiagFunc: all(
								validation.MapKeyLenBetween(1, 63),
								validation.MapValueLenBetween(1, 253),
							),
							Elem: &schema.Schema{
								Type:     schema.TypeString,
								ForceNew: true,
							},
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
			"tags": tagsSchemaForceNew(),
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func all(validators ...schema.SchemaValidateDiagFunc) schema.SchemaValidateDiagFunc {
	return func(i interface{}, k cty.Path) diag.Diagnostics {
		var diags diag.Diagnostics
		for _, validator := range validators {
			diags = append(diags, validator(i, k)...)
		}
		return diags
	}
}

func resourceAwsEksIdentityProviderConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn
	clusterName := d.Get("cluster_name").(string)

	//TODO - Should I break away from the aws sdk api and move the name outside of the oidc map?
	configName, oidcRequest := expandEksOidcIdentityProviderConfigRequest(d.Get("oidc").([]interface{}))

	id := fmt.Sprintf("%s:%s", clusterName, configName)

	input := &eks.AssociateIdentityProviderConfigInput{
		ClientRequestToken: aws.String(resource.UniqueId()),
		ClusterName:        aws.String(clusterName),
		Oidc:               oidcRequest,
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().EksTags()
	}

	_, err := conn.AssociateIdentityProviderConfig(input)
	if err != nil {
		return diag.Errorf("error associating EKS Identity Provider Config (%s): %s", id, err)
	}

	d.SetId(id)

	stateConf := resource.StateChangeConf{
		Pending: []string{eks.ConfigStatusCreating},
		Target:  []string{eks.ConfigStatusActive},
		Timeout: d.Timeout(schema.TimeoutCreate),
		Refresh: refreshEksIdentityProviderConfigStatus(conn, clusterName, &eks.IdentityProviderConfig{
			Name: aws.String(configName),
			Type: aws.String(typeOidc),
		}),
	}

	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		return diag.Errorf("error waiting for EKS Identity Provider Config (%s) association: %s", d.Id(), err)
	}

	return resourceAwsEksIdentityProviderConfigRead(ctx, d, meta)
}

func resourceAwsEksIdentityProviderConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	clusterName, configName, err := resourceAwsEksIdentityProviderConfigParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &eks.DescribeIdentityProviderConfigInput{
		ClusterName: aws.String(clusterName),
		IdentityProviderConfig: &eks.IdentityProviderConfig{
			Name: aws.String(configName),
			Type: aws.String(typeOidc),
		},
	}

	output, err := conn.DescribeIdentityProviderConfigWithContext(ctx, input)

	if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] EKS Identity Provider Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading EKS Identity Provider Config (%s): %s", d.Id(), err)
	}

	config := output.IdentityProviderConfig

	if config == nil {
		log.Printf("[WARN] EKS Identity Provider Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if config.Oidc == nil {
		log.Printf("[WARN] EKS OIDC Identity Provider Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("cluster_name", clusterName)

	if err := d.Set("oidc", flattenEksOidcIdentityProviderConfig(config.Oidc)); err != nil {
		return diag.Errorf("error setting oidc: %s", err)
	}

	d.Set("status", config.Oidc.Status)

	if err := d.Set("tags", keyvaluetags.EksKeyValueTags(config.Oidc.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsEksIdentityProviderConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn

	clusterName, configName, err := resourceAwsEksIdentityProviderConfigParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	config := &eks.IdentityProviderConfig{
		Name: aws.String(configName),
		Type: aws.String(typeOidc),
	}

	log.Printf("[DEBUG] Disassociating EKS Identity Provider Config: %s", d.Id())
	input := &eks.DisassociateIdentityProviderConfigInput{
		ClusterName:            aws.String(clusterName),
		IdentityProviderConfig: config,
	}

	_, err = conn.DisassociateIdentityProviderConfigWithContext(ctx, input)

	// TODO - Is checking for the exception message too brittle?
	if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") ||
		isAWSErr(err, eks.ErrCodeInvalidRequestException, "Identity provider config is not associated with cluster") {
		return nil
	}

	if err != nil {
		return diag.Errorf("error disassociating EKS Identity Provider Config (%s): %s", d.Id(), err)
	}

	if err := waitForEksIdentityProviderConfigDisassociation(ctx, conn, clusterName, config, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("error waiting for EKS Identity Provider Config (%s) disassociation: %s", d.Id(), err)
	}

	return nil
}

func refreshEksIdentityProviderConfigStatus(conn *eks.EKS, clusterName string, config *eks.IdentityProviderConfig) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &eks.DescribeIdentityProviderConfigInput{
			ClusterName:            aws.String(clusterName),
			IdentityProviderConfig: config,
		}

		output, err := conn.DescribeIdentityProviderConfig(input)

		if err != nil {
			return "", "", err
		}

		identityProviderConfig := output.IdentityProviderConfig

		if identityProviderConfig == nil {
			return identityProviderConfig, "", fmt.Errorf("EKS Identity Provider Config (%s:%s) missing", clusterName, aws.StringValue(config.Name))
		}

		oidc := identityProviderConfig.Oidc

		if oidc == nil {
			return identityProviderConfig, "", fmt.Errorf("EKS OIDC Identity Provider Config (%s:%s) missing", clusterName, aws.StringValue(config.Name))
		}

		return identityProviderConfig, aws.StringValue(oidc.Status), nil
	}
}

func waitForEksIdentityProviderConfigDisassociation(ctx context.Context, conn *eks.EKS, clusterName string, config *eks.IdentityProviderConfig, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			eks.ConfigStatusActive,
			eks.ConfigStatusDeleting,
		},
		Target:  []string{""},
		Timeout: timeout,
		Refresh: refreshEksIdentityProviderConfigStatus(conn, clusterName, config),
	}
	_, err := stateConf.WaitForStateContext(ctx)

	if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	return err
}

func resourceAwsEksIdentityProviderConfigParseId(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected cluster-name:identity-provider-config-name", id)
	}

	return parts[0], parts[1], nil
}

func expandEksOidcIdentityProviderConfigRequest(l []interface{}) (string, *eks.OidcIdentityProviderConfigRequest) {
	if len(l) == 0 {
		return "", nil
	}

	m := l[0].(map[string]interface{})

	configName := m["identity_provider_config_name"].(string)
	oidcIdentityProviderConfigRequest := &eks.OidcIdentityProviderConfigRequest{
		ClientId:                   aws.String(m["client_id"].(string)),
		IdentityProviderConfigName: aws.String(configName),
		IssuerUrl:                  aws.String(m["issuer_url"].(string)),
	}

	if v, ok := m["groups_claim"].(string); ok && v != "" {
		oidcIdentityProviderConfigRequest.GroupsClaim = aws.String(v)
	}

	if v, ok := m["groups_prefix"].(string); ok && v != "" {
		oidcIdentityProviderConfigRequest.GroupsPrefix = aws.String(v)
	}

	if v, ok := m["required_claims"].(map[string]interface{}); ok && len(v) > 0 {
		oidcIdentityProviderConfigRequest.RequiredClaims = stringMapToPointers(v)
	}

	if v, ok := m["username_claim"].(string); ok && v != "" {
		oidcIdentityProviderConfigRequest.UsernameClaim = aws.String(v)
	}

	if v, ok := m["username_prefix"].(string); ok && v != "" {
		oidcIdentityProviderConfigRequest.UsernamePrefix = aws.String(v)
	}

	return configName, oidcIdentityProviderConfigRequest
}

func flattenEksOidcIdentityProviderConfig(config *eks.OidcIdentityProviderConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"client_id":                     aws.StringValue(config.ClientId),
		"groups_claim":                  aws.StringValue(config.GroupsClaim),
		"groups_prefix":                 aws.StringValue(config.GroupsPrefix),
		"identity_provider_config_name": aws.StringValue(config.IdentityProviderConfigName),
		"issuer_url":                    aws.StringValue(config.IssuerUrl),
		"required_claims":               aws.StringValueMap(config.RequiredClaims),
		"username_claim":                aws.StringValue(config.UsernameClaim),
		"username_prefix":               aws.StringValue(config.UsernamePrefix),
	}

	return []map[string]interface{}{m}
}
