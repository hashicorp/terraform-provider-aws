package grafana

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceWorkspace() *schema.Resource {
	return &schema.Resource{
		Create: resourceWorkspaceCreate,
		Read:   resourceWorkspaceRead,
		Update: resourceWorkspaceUpdate,
		Delete: resourceWorkspaceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"account_access_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.AccountAccessType_Values(), false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication_providers": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(managedgrafana.AuthenticationProviderTypes_Values(), false),
				},
			},
			"data_sources": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(managedgrafana.DataSourceType_Values(), false),
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grafana_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"notification_destinations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(managedgrafana.NotificationDestinationType_Values(), false),
				},
			},
			"organization_role_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"organizational_units": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"permission_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.PermissionType_Values(), false),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"saml_configuration_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stack_set_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &managedgrafana.CreateWorkspaceInput{
		AccountAccessType:       aws.String(d.Get("account_access_type").(string)),
		AuthenticationProviders: flex.ExpandStringList(d.Get("authentication_providers").([]interface{})),
		ClientToken:             aws.String(resource.UniqueId()),
		PermissionType:          aws.String(d.Get("permission_type").(string)),
		Tags:                    Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("organization_role_name"); ok {
		input.OrganizationRoleName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("stack_set_name"); ok {
		input.StackSetName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_sources"); ok {
		input.WorkspaceDataSources = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.WorkspaceDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.WorkspaceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_destinations"); ok {
		input.WorkspaceNotificationDestinations = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("organizational_units"); ok {
		input.WorkspaceOrganizationalUnits = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.WorkspaceRoleArn = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Grafana Workspace: %s", input)
	output, err := conn.CreateWorkspace(input)

	if err != nil {
		return fmt.Errorf("error creating Grafana Workspace: %w", err)
	}

	d.SetId(aws.StringValue(output.Workspace.Id))

	if _, err := waitWorkspaceCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Grafana Workspace (%s) create: %w", d.Id(), err)
	}

	return resourceWorkspaceRead(d, meta)
}

func resourceWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	workspace, err := FindWorkspaceByID(conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Grafana Workspace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Grafana Workspace (%s): %w", d.Id(), err)
	}

	d.Set("account_access_type", workspace.AccountAccessType)
	// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonmanagedgrafana.html#amazonmanagedgrafana-resources-for-iam-policies.
	workspaceARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   managedgrafana.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("/workspaces/%s", d.Id()),
	}.String()
	d.Set("arn", workspaceARN)
	d.Set("authentication_providers", workspace.Authentication.Providers)
	d.Set("data_sources", workspace.DataSources)
	d.Set("description", workspace.Description)
	d.Set("endpoint", workspace.Endpoint)
	d.Set("grafana_version", workspace.GrafanaVersion)
	d.Set("name", workspace.Name)
	d.Set("notification_destinations", workspace.NotificationDestinations)
	d.Set("organization_role_name", workspace.OrganizationRoleName)
	d.Set("organizational_units", workspace.OrganizationalUnits)
	d.Set("permission_type", workspace.PermissionType)
	d.Set("role_arn", workspace.WorkspaceRoleArn)
	d.Set("saml_configuration_status", workspace.Authentication.SamlConfigurationStatus)
	d.Set("stack_set_name", workspace.StackSetName)

	tags := KeyValueTags(workspace.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &managedgrafana.UpdateWorkspaceInput{
			WorkspaceId: aws.String(d.Id()),
		}

		if d.HasChange("account_access_type") {
			input.AccountAccessType = aws.String(d.Get("account_access_type").(string))
		}

		if d.HasChange("data_sources") {
			input.WorkspaceDataSources = flex.ExpandStringList(d.Get("data_sources").([]interface{}))
		}

		if d.HasChange("description") {
			input.WorkspaceDescription = aws.String(d.Get("description").(string))
		}

		if d.HasChange("name") {
			input.WorkspaceName = aws.String(d.Get("name").(string))
		}

		if d.HasChange("notification_destinations") {
			input.WorkspaceNotificationDestinations = flex.ExpandStringList(d.Get("notification_destinations").([]interface{}))
		}

		if d.HasChange("organization_role_name") {
			input.OrganizationRoleName = aws.String(d.Get("organization_role_name").(string))
		}

		if d.HasChange("organizational_units") {
			input.WorkspaceOrganizationalUnits = flex.ExpandStringList(d.Get("organizational_units").([]interface{}))
		}

		if d.HasChange("permission_type") {
			input.PermissionType = aws.String(d.Get("permission_type").(string))
		}

		if d.HasChange("role_arn") {
			input.WorkspaceRoleArn = aws.String(d.Get("role_arn").(string))
		}

		if d.HasChange("stack_set_name") {
			input.StackSetName = aws.String(d.Get("stack_set_name").(string))
		}

		_, err := conn.UpdateWorkspace(input)

		if err != nil {
			return fmt.Errorf("error updating Grafana Workspace (%s): %w", d.Id(), err)
		}

		if _, err := waitWorkspaceUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Grafana Workspace (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Grafana Workspace (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceWorkspaceRead(d, meta)
}

func resourceWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	log.Printf("[DEBUG] Deleting Grafana Workspace: %s", d.Id())
	_, err := conn.DeleteWorkspace(&managedgrafana.DeleteWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, managedgrafana.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Grafana Workspace (%s): %w", d.Id(), err)
	}

	if _, err := waitWorkspaceDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Grafana Workspace (%s) delete: %w", d.Id(), err)
	}

	return nil
}
