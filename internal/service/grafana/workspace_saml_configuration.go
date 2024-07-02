// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_grafana_workspace_saml_configuration", name="Grafana Workspace SAML Configuration")
func resourceWorkspaceSAMLConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkspaceSAMLConfigurationUpsert,
		ReadWithoutTimeout:   resourceWorkspaceSAMLConfigurationRead,
		UpdateWithoutTimeout: resourceWorkspaceSAMLConfigurationUpsert,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"admin_role_values": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"allowed_organizations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"editor_role_values": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"email_assertion": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"groups_assertion": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"idp_metadata_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"idp_metadata_xml": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"login_assertion": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"login_validity_duration": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"name_assertion": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"org_assertion": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role_assertion": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceWorkspaceSAMLConfigurationUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	workspaceID := d.Get("workspace_id").(string)
	workspace, err := findWorkspaceByID(ctx, conn, workspaceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Grafana Workspace (%s): %s", workspaceID, err)
	}

	authenticationProviders := workspace.Authentication.Providers
	roleValues := &awstypes.RoleValues{
		Editor: flex.ExpandStringValueList(d.Get("editor_role_values").([]interface{})),
	}

	if v, ok := d.GetOk("admin_role_values"); ok {
		roleValues.Admin = flex.ExpandStringValueList(v.([]interface{}))
	}

	samlConfiguration := &awstypes.SamlConfiguration{
		RoleValues: roleValues,
	}

	if v, ok := d.GetOk("allowed_organizations"); ok {
		samlConfiguration.AllowedOrganizations = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk("login_validity_duration"); ok {
		samlConfiguration.LoginValidityDuration = int32(v.(int))
	}

	var assertionAttributes *awstypes.AssertionAttributes

	if v, ok := d.GetOk("email_assertion"); ok {
		assertionAttributes = &awstypes.AssertionAttributes{
			Email: aws.String(v.(string)),
		}
	}

	if v, ok := d.GetOk("groups_assertion"); ok {
		if assertionAttributes == nil {
			assertionAttributes = &awstypes.AssertionAttributes{}
		}

		assertionAttributes.Groups = aws.String(v.(string))
	}

	if v, ok := d.GetOk("login_assertion"); ok {
		if assertionAttributes == nil {
			assertionAttributes = &awstypes.AssertionAttributes{}
		}

		assertionAttributes.Login = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name_assertion"); ok {
		if assertionAttributes == nil {
			assertionAttributes = &awstypes.AssertionAttributes{}
		}

		assertionAttributes.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("org_assertion"); ok {
		if assertionAttributes == nil {
			assertionAttributes = &awstypes.AssertionAttributes{}
		}

		assertionAttributes.Org = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_assertion"); ok {
		if assertionAttributes == nil {
			assertionAttributes = &awstypes.AssertionAttributes{}
		}

		assertionAttributes.Role = aws.String(v.(string))
	}

	if assertionAttributes != nil {
		samlConfiguration.AssertionAttributes = assertionAttributes
	}

	var idpMetadata awstypes.IdpMetadata

	if v, ok := d.GetOk("idp_metadata_url"); ok {
		idpMetadata = &awstypes.IdpMetadataMemberUrl{
			Value: v.(string),
		}
	}

	if v, ok := d.GetOk("idp_metadata_xml"); ok {
		idpMetadata = &awstypes.IdpMetadataMemberXml{
			Value: v.(string),
		}
	}

	samlConfiguration.IdpMetadata = idpMetadata

	input := &grafana.UpdateWorkspaceAuthenticationInput{
		AuthenticationProviders: authenticationProviders,
		SamlConfiguration:       samlConfiguration,
		WorkspaceId:             aws.String(workspaceID),
	}

	_, err = conn.UpdateWorkspaceAuthentication(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Grafana Workspace SAML Configuration (%s): %s", workspaceID, err)
	}

	if d.IsNewResource() {
		d.SetId(workspaceID)
	}

	if _, err := waitWorkspaceSAMLConfigurationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Grafana Workspace SAML Configuration (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceWorkspaceSAMLConfigurationRead(ctx, d, meta)...)
}

func resourceWorkspaceSAMLConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	saml, err := findSAMLConfigurationByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Grafana Workspace SAML Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Grafana Workspace SAML Configuration (%s): %s", d.Id(), err)
	}

	d.Set("admin_role_values", saml.Configuration.RoleValues.Admin)
	d.Set("allowed_organizations", saml.Configuration.AllowedOrganizations)
	d.Set("editor_role_values", saml.Configuration.RoleValues.Editor)
	d.Set("email_assertion", saml.Configuration.AssertionAttributes.Email)
	d.Set("groups_assertion", saml.Configuration.AssertionAttributes.Groups)
	d.Set("idp_metadata_url", "")
	d.Set("idp_metadata_xml", "")
	switch v := saml.Configuration.IdpMetadata.(type) {
	case *awstypes.IdpMetadataMemberUrl:
		d.Set("idp_metadata_url", v.Value)
	case *awstypes.IdpMetadataMemberXml:
		d.Set("idp_metadata_xml", v.Value)
	}
	d.Set("login_assertion", saml.Configuration.AssertionAttributes.Login)
	d.Set("login_validity_duration", saml.Configuration.LoginValidityDuration)
	d.Set("name_assertion", saml.Configuration.AssertionAttributes.Name)
	d.Set("org_assertion", saml.Configuration.AssertionAttributes.Org)
	d.Set("role_assertion", saml.Configuration.AssertionAttributes.Role)
	d.Set(names.AttrStatus, saml.Status)

	return diags
}

func findSAMLConfigurationByID(ctx context.Context, conn *grafana.Client, id string) (*awstypes.SamlAuthentication, error) {
	input := &grafana.DescribeWorkspaceAuthenticationInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeWorkspaceAuthentication(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Authentication == nil || output.Authentication.Saml == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.Authentication.Saml.Status; status == awstypes.SamlConfigurationStatusNotConfigured {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output.Authentication.Saml, nil
}

func statusWorkspaceSAMLConfiguration(ctx context.Context, conn *grafana.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findSAMLConfigurationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitWorkspaceSAMLConfigurationCreated(ctx context.Context, conn *grafana.Client, id string, timeout time.Duration) (*awstypes.SamlAuthentication, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.SamlConfigurationStatusConfigured),
		Refresh: statusWorkspaceSAMLConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.SamlAuthentication); ok {
		return output, err
	}

	return nil, err
}
