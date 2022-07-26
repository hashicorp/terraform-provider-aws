package grafana

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceWorkspaceSAMLConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkspaceSAMLConfigurationUpsert,
		Read:   resourceWorkspaceSAMLConfigurationRead,
		Update: resourceWorkspaceSAMLConfigurationUpsert,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"status": {
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

func resourceWorkspaceSAMLConfigurationUpsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	d.SetId(d.Get("workspace_id").(string))
	workspace, err := FindWorkspaceByID(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading Grafana Workspace (%s): %w", d.Id(), err)
	}

	authenticationProviders := workspace.Authentication.Providers
	roleValues := &managedgrafana.RoleValues{
		Editor: flex.ExpandStringList(d.Get("editor_role_values").([]interface{})),
	}

	if v, ok := d.GetOk("admin_role_values"); ok {
		roleValues.Admin = flex.ExpandStringList(v.([]interface{}))
	}

	samlConfiguration := &managedgrafana.SamlConfiguration{
		RoleValues: roleValues,
	}

	if v, ok := d.GetOk("allowed_organizations"); ok {
		samlConfiguration.AllowedOrganizations = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("login_validity_duration"); ok {
		samlConfiguration.LoginValidityDuration = aws.Int64(int64(v.(int)))
	}

	var assertionAttributes *managedgrafana.AssertionAttributes

	if v, ok := d.GetOk("email_assertion"); ok {
		assertionAttributes = &managedgrafana.AssertionAttributes{
			Email: aws.String(v.(string)),
		}
	}

	if v, ok := d.GetOk("groups_assertion"); ok {
		if assertionAttributes == nil {
			assertionAttributes = &managedgrafana.AssertionAttributes{}
		}

		assertionAttributes.Groups = aws.String(v.(string))
	}

	if v, ok := d.GetOk("login_assertion"); ok {
		if assertionAttributes == nil {
			assertionAttributes = &managedgrafana.AssertionAttributes{}
		}

		assertionAttributes.Login = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name_assertion"); ok {
		if assertionAttributes == nil {
			assertionAttributes = &managedgrafana.AssertionAttributes{}
		}

		assertionAttributes.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("org_assertion"); ok {
		if assertionAttributes == nil {
			assertionAttributes = &managedgrafana.AssertionAttributes{}
		}

		assertionAttributes.Org = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_assertion"); ok {
		if assertionAttributes == nil {
			assertionAttributes = &managedgrafana.AssertionAttributes{}
		}

		assertionAttributes.Role = aws.String(v.(string))
	}

	if assertionAttributes != nil {
		samlConfiguration.AssertionAttributes = assertionAttributes
	}

	var idpMetadata *managedgrafana.IdpMetadata

	if v, ok := d.GetOk("idp_metadata_url"); ok {
		idpMetadata = &managedgrafana.IdpMetadata{
			Url: aws.String(v.(string)),
		}
	}

	if v, ok := d.GetOk("idp_metadata_xml"); ok {
		idpMetadata = &managedgrafana.IdpMetadata{
			Xml: aws.String(v.(string)),
		}
	}

	samlConfiguration.IdpMetadata = idpMetadata

	input := &managedgrafana.UpdateWorkspaceAuthenticationInput{
		AuthenticationProviders: authenticationProviders,
		SamlConfiguration:       samlConfiguration,
		WorkspaceId:             aws.String(d.Id()),
	}

	_, err = conn.UpdateWorkspaceAuthentication(input)

	if err != nil {
		return fmt.Errorf("error creating Grafana Saml Configuration: %w", err)
	}

	if _, err := waitWorkspaceSAMLConfigurationCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Grafana Workspace Saml Configuration (%s) create: %w", d.Id(), err)
	}

	return resourceWorkspaceSAMLConfigurationRead(d, meta)
}

func resourceWorkspaceSAMLConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	saml, err := FindSamlConfigurationByID(conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Grafana Workspace Saml Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Grafana Workspace Saml Configuration (%s): %w", d.Id(), err)
	}

	if saml.Configuration.RoleValues.Admin != nil {
		d.Set("admin_role_values", saml.Configuration.RoleValues.Admin)
	}

	if saml.Configuration.AllowedOrganizations != nil {
		d.Set("allowed_organizations", saml.Configuration.AllowedOrganizations)
	}

	d.Set("editor_role_values", saml.Configuration.RoleValues.Editor)

	if saml.Configuration.AssertionAttributes.Email != nil {
		d.Set("email_assertion", saml.Configuration.AssertionAttributes.Email)
	}

	if saml.Configuration.AssertionAttributes.Groups != nil {
		d.Set("groups_assertion", saml.Configuration.AssertionAttributes.Groups)
	}

	if saml.Configuration.IdpMetadata.Url != nil {
		d.Set("idp_metadata_url", saml.Configuration.IdpMetadata.Url)
	}

	if saml.Configuration.IdpMetadata.Xml != nil {
		d.Set("idp_metadata_xml", saml.Configuration.IdpMetadata.Xml)
	}

	if saml.Configuration.AssertionAttributes.Login != nil {
		d.Set("login_assertion", saml.Configuration.AssertionAttributes.Login)
	}

	if saml.Configuration.LoginValidityDuration != nil {
		d.Set("login_validity_duration", saml.Configuration.LoginValidityDuration)
	}

	if saml.Configuration.AssertionAttributes.Name != nil {
		d.Set("name_assertion", saml.Configuration.AssertionAttributes.Name)
	}

	if saml.Configuration.AssertionAttributes.Org != nil {
		d.Set("org_assertion", saml.Configuration.AssertionAttributes.Org)
	}

	if saml.Configuration.AssertionAttributes.Role != nil {
		d.Set("role_assertion", saml.Configuration.AssertionAttributes.Role)
	}

	d.Set("status", saml.Status)

	return nil
}
