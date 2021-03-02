package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strings"
	"time"
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
			// TODO - Do I set ForceNew here to true as it doesn't look like you can update a identity provider config
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
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
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

	//if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
	//	input.Tags = keyvaluetags.New(v).IgnoreAws().EksTags()
	//}

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
	//ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

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

	// TODO - Do I need the nil check here?
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

	//if err := d.Set("tags", keyvaluetags.EksKeyValueTags(fargateProfile.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
	//	return fmt.Errorf("error setting tags: %s", err)
	//}

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

	if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	// TODO - if we timeout on the association whilst acc testing this will fail with:
	// {
	//   RespMetadata: {
	//     StatusCode: 400,
	// 	   RequestID: "abf3f851-04e7-4d66-867e-f871897efbb3"
	//   },
	//   ClusterName: "tf-acc-test-387064423719985885",
	// 	 Message_: "Identity provider config is not associated with cluster tf-acc-test-387064423719985885"
	// }
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

		// TODO: not sure about the use of StringValue here.
		if identityProviderConfig == nil {
			return identityProviderConfig, "", fmt.Errorf("EKS Identity Provider Config (%s:%s) missing", clusterName, aws.StringValue(config.Name))
		}

		oidc := identityProviderConfig.Oidc

		// TODO: Should I be checking for nil on OIDC here too?
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

	return configName, oidcIdentityProviderConfigRequest
}

func flattenEksOidcIdentityProviderConfig(config *eks.OidcIdentityProviderConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"client_id":                     aws.StringValue(config.ClientId),
		"identity_provider_config_name": aws.StringValue(config.IdentityProviderConfigName),
		"issuer_url":                    aws.StringValue(config.IssuerUrl),
	}

	return []map[string]interface{}{m}
}
