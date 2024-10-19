// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_connection", name="Connection")
// @Tags(identifierAttribute="arn")
func ResourceConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectionCreate,
		ReadWithoutTimeout:   resourceConnectionRead,
		UpdateWithoutTimeout: resourceConnectionUpdate,
		DeleteWithoutTimeout: resourceConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authentication_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AuthenticationType](),
						},
						"oauth2_properties": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authorization_code_properties": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"authorization_code": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"redirect_uri": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"oauth2_client_application": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"aws_managed_client_application_reference": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"user_managed_client_application_client_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"oauth2_grant_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.OAuth2GrantType](),
									},
									"token_url": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsURLWithHTTPS,
									},
									"token_url_parameters_map": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"secret_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Computed: true,
			},
			"connection_properties": {
				Type:             schema.TypeMap,
				Optional:         true,
				Sensitive:        true,
				ValidateDiagFunc: verify.MapKeysAre(validation.ToDiagFunc(validation.StringInSlice(connectionPropertyKey_Values(), false))),
				Elem:             &schema.Schema{Type: schema.TypeString},
			},
			"connection_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      awstypes.ConnectionTypeJdbc,
				ValidateFunc: validation.StringInSlice(connectionType_Values(), false),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"match_criteria": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"physical_connection_requirements": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"security_group_id_list": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetID: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	var catalogID string
	if v, ok := d.GetOkExists(names.AttrCatalogID); ok {
		catalogID = v.(string)
	} else {
		catalogID = meta.(*conns.AWSClient).AccountID
	}
	name := d.Get(names.AttrName).(string)

	input := &glue.CreateConnectionInput{
		CatalogId:       aws.String(catalogID),
		ConnectionInput: expandConnectionInput(d),
		Tags:            getTagsIn(ctx),
	}

	log.Printf("[DEBUG] Creating Glue Connection: %+v", input)
	_, err := conn.CreateConnection(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Connection (%s): %s", name, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", catalogID, name))

	return append(diags, resourceConnectionRead(ctx, d, meta)...)
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, connectionName, err := DecodeConnectionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Connection (%s): %s", d.Id(), err)
	}

	connection, err := FindConnectionByName(ctx, conn, connectionName, catalogID)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glue Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Connection (%s): %s", d.Id(), err)
	}

	connectionArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("connection/%s", connectionName),
	}.String()
	d.Set(names.AttrARN, connectionArn)

	d.Set(names.AttrCatalogID, catalogID)
	if err := d.Set("connection_properties", connection.ConnectionProperties); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting connection_properties: %s", err)
	}
	d.Set("connection_type", connection.ConnectionType)
	d.Set(names.AttrDescription, connection.Description)
	if err := d.Set("match_criteria", flex.FlattenStringValueList(connection.MatchCriteria)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting match_criteria: %s", err)
	}
	d.Set(names.AttrName, connection.Name)
	if err := d.Set("physical_connection_requirements", flattenPhysicalConnectionRequirements(connection.PhysicalConnectionRequirements)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting physical_connection_requirements: %s", err)
	}

	if err := d.Set("authentication_configuration", flattenAuthenticationConfiguration(connection.AuthenticationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting authentication_configuration: %s", err)
	}

	return diags
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		catalogID, connectionName, err := DecodeConnectionID(d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Connection (%s): %s", d.Id(), err)
		}

		input := &glue.UpdateConnectionInput{
			CatalogId:       aws.String(catalogID),
			ConnectionInput: expandConnectionInput(d),
			Name:            aws.String(connectionName),
		}

		log.Printf("[DEBUG] Updating Glue Connection: %+v", input)
		_, err = conn.UpdateConnection(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Connection (%s): %s", d.Id(), err)
		}
	}

	return diags
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	catalogID, connectionName, err := DecodeConnectionID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Connection (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Glue Connection: %s", d.Id())
	err = DeleteConnection(ctx, conn, catalogID, connectionName)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Connection (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeConnectionID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format CATALOG-ID:NAME, provided: %s", id)
	}
	return idParts[0], idParts[1], nil
}

func DeleteConnection(ctx context.Context, conn *glue.Client, catalogID, connectionName string) error {
	input := &glue.DeleteConnectionInput{
		CatalogId:      aws.String(catalogID),
		ConnectionName: aws.String(connectionName),
	}

	_, err := conn.DeleteConnection(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.EntityNotFoundException](err) {
			return nil
		}
		return err
	}

	return nil
}

func expandConnectionInput(d *schema.ResourceData) *awstypes.ConnectionInput {
	connectionProperties := make(map[string]string)
	if val, ok := d.GetOkExists("connection_properties"); ok {
		for k, v := range val.(map[string]interface{}) {
			connectionProperties[k] = v.(string)
		}
	}

	connectionInput := &awstypes.ConnectionInput{
		ConnectionProperties: connectionProperties,
		ConnectionType:       awstypes.ConnectionType(d.Get("connection_type").(string)),
		Name:                 aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("authentication_configuration"); ok && v.([]interface{})[0] != nil {
		authenticationConfigurationMap := v.([]interface{})[0].(map[string]interface{})
		connectionInput.AuthenticationConfiguration = expandAuthenticationConfiguration(authenticationConfigurationMap)
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		connectionInput.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("match_criteria"); ok {
		connectionInput.MatchCriteria = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk("physical_connection_requirements"); ok && v.([]interface{})[0] != nil {
		physicalConnectionRequirementsMap := v.([]interface{})[0].(map[string]interface{})
		connectionInput.PhysicalConnectionRequirements = expandPhysicalConnectionRequirements(physicalConnectionRequirementsMap)
	}

	return connectionInput
}

func expandPhysicalConnectionRequirements(m map[string]interface{}) *awstypes.PhysicalConnectionRequirements {
	physicalConnectionRequirements := &awstypes.PhysicalConnectionRequirements{}

	if v, ok := m[names.AttrAvailabilityZone]; ok {
		physicalConnectionRequirements.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := m["security_group_id_list"]; ok {
		physicalConnectionRequirements.SecurityGroupIdList = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := m[names.AttrSubnetID]; ok {
		physicalConnectionRequirements.SubnetId = aws.String(v.(string))
	}

	return physicalConnectionRequirements
}

func flattenPhysicalConnectionRequirements(physicalConnectionRequirements *awstypes.PhysicalConnectionRequirements) []map[string]interface{} {
	if physicalConnectionRequirements == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrAvailabilityZone: aws.ToString(physicalConnectionRequirements.AvailabilityZone),
		"security_group_id_list":   flex.FlattenStringValueSet(physicalConnectionRequirements.SecurityGroupIdList),
		names.AttrSubnetID:         aws.ToString(physicalConnectionRequirements.SubnetId),
	}

	return []map[string]interface{}{m}
}

func expandAuthenticationConfiguration(m map[string]interface{}) *awstypes.AuthenticationConfigurationInput {
	conf := &awstypes.AuthenticationConfigurationInput{}

	if v, ok := m["authentication_type"]; ok {
		conf.AuthenticationType = awstypes.AuthenticationType(v.(string))
	}

	if v, ok := m["code_editor_app_settings"].([]interface{}); ok && len(v) > 0 {
		conf.OAuth2Properties = expandOAuth2Properties(v)
	}

	if v, ok := m["secret_arn"]; ok {
		conf.SecretArn = aws.String(v.(string))
	}

	return conf
}

func expandOAuth2Properties(l []interface{}) *awstypes.OAuth2PropertiesInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.OAuth2PropertiesInput{}

	if v, ok := m["authorization_code_properties"].([]interface{}); ok && len(v) > 0 {
		config.AuthorizationCodeProperties = expandAuthorizationCodeProperties(v)
	}

	if v, ok := m["oauth2_client_application"].([]interface{}); ok && len(v) > 0 {
		config.OAuth2ClientApplication = expandOAuth2ClientApplication(v)
	}

	if v, ok := m["oauth2_grant_type"]; ok {
		config.OAuth2GrantType = awstypes.OAuth2GrantType(v.(string))
	}

	if v, ok := m["token_url"]; ok {
		config.TokenUrl = aws.String(v.(string))
	}

	if v, ok := m["token_url_parameters_map"]; ok {
		config.TokenUrlParametersMap = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	return config
}

func expandOAuth2ClientApplication(l []interface{}) *awstypes.OAuth2ClientApplication {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.OAuth2ClientApplication{}

	if v, ok := m["aws_managed_client_application_reference"]; ok {
		config.AWSManagedClientApplicationReference = aws.String(v.(string))
	}

	if v, ok := m["user_managed_client_application_client_id"]; ok {
		config.UserManagedClientApplicationClientId = aws.String(v.(string))
	}

	return config
}

func expandAuthorizationCodeProperties(l []interface{}) *awstypes.AuthorizationCodeProperties {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &awstypes.AuthorizationCodeProperties{}

	if v, ok := m["authorization_code"]; ok {
		config.AuthorizationCode = aws.String(v.(string))
	}

	if v, ok := m["redirect_uri"]; ok {
		config.RedirectUri = aws.String(v.(string))
	}

	return config
}

func flattenAuthenticationConfiguration(conf *awstypes.AuthenticationConfiguration) []map[string]interface{} {
	if conf == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"authentication_type": conf.AuthenticationType,
		"oauth2_properties":   flattenOAuth2Properties(conf.OAuth2Properties),
		"secret_arn":          aws.ToString(conf.SecretArn),
	}

	return []map[string]interface{}{m}
}

func flattenOAuth2Properties(config *awstypes.OAuth2Properties) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.OAuth2ClientApplication != nil {
		m["oauth2_client_application"] = flattenOAuth2ClientApplication(config.OAuth2ClientApplication)
	}

	if config.OAuth2GrantType != "" {
		m["oauth2_grant_type"] = config.OAuth2GrantType
	}

	if config.TokenUrl != nil {
		m["token_url"] = aws.ToString(config.TokenUrl)
	}

	if config.TokenUrlParametersMap != nil {
		m["token_url_parameters_map"] = flex.FlattenStringValueMap(config.TokenUrlParametersMap)
	}

	return []map[string]interface{}{m}
}

func flattenOAuth2ClientApplication(config *awstypes.OAuth2ClientApplication) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if config.AWSManagedClientApplicationReference != nil {
		m["aws_managed_client_application_reference"] = aws.ToString(config.AWSManagedClientApplicationReference)
	}

	if config.UserManagedClientApplicationClientId != nil {
		m["user_managed_client_application_client_id"] = aws.ToString(config.UserManagedClientApplicationClientId)
	}

	return []map[string]interface{}{m}
}

func connectionPropertyKey_Values() []string {
	return tfslices.AppendUnique(enum.Values[awstypes.ConnectionPropertyKey](), "SparkProperties")
}

func connectionType_Values() []string {
	return tfslices.AppendUnique(enum.Values[awstypes.ConnectionType](), "AZURECOSMOS", "AZURESQL", "BIGQUERY", "OPENSEARCH", "SNOWFLAKE")
}
