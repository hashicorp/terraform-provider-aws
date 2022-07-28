package cognitoidentity

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourcePool() *schema.Resource {
	return &schema.Resource{
		Create: resourcePoolCreate,
		Read:   resourcePoolRead,
		Update: resourcePoolUpdate,
		Delete: resourcePoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"identity_pool_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentityPoolName,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cognito_identity_providers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validIdentityProvidersClientID,
						},
						"provider_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validIdentityProvidersProviderName,
						},
						"server_side_token_check": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"developer_provider_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true, // Forcing a new resource since it cannot be edited afterwards
				ValidateFunc: validProviderDeveloperName,
			},

			"allow_unauthenticated_identities": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"allow_classic_flow": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"openid_connect_provider_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},

			"saml_provider_arns": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},

			"supported_login_providers": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validSupportedLoginProviders,
				},
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePoolCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIdentityConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	log.Print("[DEBUG] Creating Cognito Identity Pool")

	params := &cognitoidentity.CreateIdentityPoolInput{
		IdentityPoolName:               aws.String(d.Get("identity_pool_name").(string)),
		AllowUnauthenticatedIdentities: aws.Bool(d.Get("allow_unauthenticated_identities").(bool)),
		AllowClassicFlow:               aws.Bool(d.Get("allow_classic_flow").(bool)),
	}

	if v, ok := d.GetOk("developer_provider_name"); ok {
		params.DeveloperProviderName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("supported_login_providers"); ok {
		params.SupportedLoginProviders = expandSupportedLoginProviders(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("cognito_identity_providers"); ok {
		params.CognitoIdentityProviders = expandIdentityProviders(v.(*schema.Set))
	}

	if v, ok := d.GetOk("saml_provider_arns"); ok {
		params.SamlProviderARNs = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("openid_connect_provider_arns"); ok {
		params.OpenIdConnectProviderARNs = flex.ExpandStringSet(v.(*schema.Set))
	}

	if len(tags) > 0 {
		params.IdentityPoolTags = Tags(tags.IgnoreAWS())
	}

	entity, err := conn.CreateIdentityPool(params)
	if err != nil {
		return fmt.Errorf("Error creating Cognito Identity Pool: %s", err)
	}

	d.SetId(aws.StringValue(entity.IdentityPoolId))

	return resourcePoolRead(d, meta)
}

func resourcePoolRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIdentityConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading Cognito Identity Pool: %s", d.Id())

	ip, err := conn.DescribeIdentityPool(&cognitoidentity.DescribeIdentityPoolInput{
		IdentityPoolId: aws.String(d.Id()),
	})
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cognitoidentity.ErrCodeResourceNotFoundException) {
		names.LogNotFoundRemoveState(names.CognitoIdentity, names.ErrActionReading, ResPool, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CognitoIdentity, names.ErrActionReading, ResPool, d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "cognito-identity",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("identitypool/%s", d.Id()),
	}
	d.Set("arn", arn.String())
	d.Set("identity_pool_name", ip.IdentityPoolName)
	d.Set("allow_unauthenticated_identities", ip.AllowUnauthenticatedIdentities)
	d.Set("allow_classic_flow", ip.AllowClassicFlow)
	d.Set("developer_provider_name", ip.DeveloperProviderName)
	tags := KeyValueTags(ip.IdentityPoolTags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if err := d.Set("cognito_identity_providers", flattenIdentityProviders(ip.CognitoIdentityProviders)); err != nil {
		return fmt.Errorf("Error setting cognito_identity_providers error: %w", err)
	}

	if err := d.Set("openid_connect_provider_arns", flex.FlattenStringList(ip.OpenIdConnectProviderARNs)); err != nil {
		return fmt.Errorf("Error setting openid_connect_provider_arns error: %w", err)
	}

	if err := d.Set("saml_provider_arns", flex.FlattenStringList(ip.SamlProviderARNs)); err != nil {
		return fmt.Errorf("Error setting saml_provider_arns error: %w", err)
	}

	if err := d.Set("supported_login_providers", aws.StringValueMap(ip.SupportedLoginProviders)); err != nil {
		return fmt.Errorf("Error setting supported_login_providers error: %w", err)
	}

	return nil
}

func resourcePoolUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIdentityConn
	log.Print("[DEBUG] Updating Cognito Identity Pool")

	params := &cognitoidentity.IdentityPool{
		IdentityPoolId:                 aws.String(d.Id()),
		AllowUnauthenticatedIdentities: aws.Bool(d.Get("allow_unauthenticated_identities").(bool)),
		AllowClassicFlow:               aws.Bool(d.Get("allow_classic_flow").(bool)),
		IdentityPoolName:               aws.String(d.Get("identity_pool_name").(string)),
	}

	if d.HasChanges(
		"cognito_identity_providers",
		"supported_login_providers",
		"openid_connect_provider_arns",
		"saml_provider_arns",
	) {
		params.CognitoIdentityProviders = expandIdentityProviders(d.Get("cognito_identity_providers").(*schema.Set))
		params.SupportedLoginProviders = expandSupportedLoginProviders(d.Get("supported_login_providers").(map[string]interface{}))
		params.OpenIdConnectProviderARNs = flex.ExpandStringSet(d.Get("openid_connect_provider_arns").(*schema.Set))
		params.SamlProviderARNs = flex.ExpandStringList(d.Get("saml_provider_arns").([]interface{}))
	}

	_, err := conn.UpdateIdentityPool(params)
	if err != nil {
		return fmt.Errorf("Error updating Cognito Identity Pool: %s", err)
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Cognito Identity Pool (%s) tags: %s", arn, err)
		}
	}

	return resourcePoolRead(d, meta)
}

func resourcePoolDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CognitoIdentityConn
	log.Printf("[DEBUG] Deleting Cognito Identity Pool: %s", d.Id())

	_, err := conn.DeleteIdentityPool(&cognitoidentity.DeleteIdentityPoolInput{
		IdentityPoolId: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("Error deleting Cognito identity pool: %s", err)
	}
	return nil
}
