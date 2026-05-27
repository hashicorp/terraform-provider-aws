// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package glue

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_glue_connection", name="Connection")
// @Tags(identifierAttribute="arn")
func dataSourceConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectionRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"athena_properties": {
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
			},
			"authentication_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authentication_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"basic_authentication_credentials": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrPassword: {
										Type:      schema.TypeString,
										Computed:  true,
										Sensitive: true,
									},
									names.AttrUsername: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"custom_authentication_credentials": {
							Type:      schema.TypeMap,
							Computed:  true,
							Sensitive: true,
							Elem:      &schema.Schema{Type: schema.TypeString},
						},
						names.AttrKMSKeyARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"oauth2_properties": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authorization_code_properties": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"authorization_code": {
													Type:      schema.TypeString,
													Computed:  true,
													Sensitive: true,
												},
												"redirect_uri": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"oauth2_client_application": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"aws_managed_client_application_reference": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"user_managed_client_application_client_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"oauth2_credentials": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"access_token": {
													Type:      schema.TypeString,
													Computed:  true,
													Sensitive: true,
												},
												"jwt_token": {
													Type:      schema.TypeString,
													Computed:  true,
													Sensitive: true,
												},
												"refresh_token": {
													Type:      schema.TypeString,
													Computed:  true,
													Sensitive: true,
												},
												"user_managed_client_application_client_secret": {
													Type:      schema.TypeString,
													Computed:  true,
													Sensitive: true,
												},
											},
										},
									},
									"oauth2_grant_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"token_url": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"token_url_parameters_map": {
										Type:      schema.TypeMap,
										Computed:  true,
										Sensitive: true,
										Elem:      &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"secret_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrCatalogID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_properties": {
				Type:      schema.TypeMap,
				Computed:  true,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
			},
			"connection_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"match_criteria": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"physical_connection_requirements": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"security_group_id_list": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.GlueClient(ctx)

	id := d.Get(names.AttrID).(string)
	catalogID, connectionName, err := connectionParseResourceID(id)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "decoding Glue Connection %s: %s", id, err)
	}

	connection, err := findConnectionByTwoPartKey(ctx, conn, connectionName, catalogID)
	if err != nil {
		if retry.NotFound(err) {
			return sdkdiag.AppendErrorf(diags, "Glue Connection (%s) not found", id)
		}
		return sdkdiag.AppendErrorf(diags, "reading Glue Connection (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set(names.AttrARN, connectionARN(ctx, c, connectionName))
	d.Set("athena_properties", connection.AthenaProperties)
	if err := d.Set("authentication_configuration", flattenAuthenticationConfiguration(connection.AuthenticationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting authentication_configuration: %s", err)
	}
	d.Set(names.AttrCatalogID, catalogID)
	d.Set("connection_properties", connection.ConnectionProperties)
	d.Set("connection_type", connection.ConnectionType)
	d.Set(names.AttrDescription, connection.Description)
	d.Set("match_criteria", connection.MatchCriteria)
	d.Set(names.AttrName, connection.Name)
	if err := d.Set("physical_connection_requirements", flattenPhysicalConnectionRequirements(connection.PhysicalConnectionRequirements)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting physical_connection_requirements: %s", err)
	}

	return diags
}
