package workspaces

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
		Schema: map[string]*schema.Schema{
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"computer_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_volume_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_volume_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"volume_encryption_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"workspace_properties": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_type_name": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      workspaces.ComputeValue,
							ValidateFunc: validation.StringInSlice(workspaces.Compute_Values(), false),
						},
						"root_volume_size_gib": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  80,
							ValidateFunc: validation.Any(
								validation.IntInSlice([]int{80}),
								validation.IntBetween(175, 2000),
							),
						},
						"running_mode": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  workspaces.RunningModeAlwaysOn,
							ValidateFunc: validation.StringInSlice([]string{
								workspaces.RunningModeAlwaysOn,
								workspaces.RunningModeAutoStop,
							}, false),
						},
						"running_mode_auto_stop_timeout_in_minutes": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								val := v.(int)
								if val%60 != 0 {
									errors = append(errors, fmt.Errorf(
										"%q should be configured in 60-minute intervals, got: %d", k, val))
								}
								return
							},
						},
						"user_volume_size_gib": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  10,
							ValidateFunc: validation.Any(
								validation.IntInSlice([]int{10, 50}),
								validation.IntBetween(100, 2000),
							),
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(WorkspaceAvailableTimeout),
			Update: schema.DefaultTimeout(WorkspaceUpdatingTimeout),
			Delete: schema.DefaultTimeout(WorkspaceTerminatedTimeout),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceWorkspaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkSpacesConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &workspaces.WorkspaceRequest{
		BundleId:                    aws.String(d.Get("bundle_id").(string)),
		DirectoryId:                 aws.String(d.Get("directory_id").(string)),
		UserName:                    aws.String(d.Get("user_name").(string)),
		RootVolumeEncryptionEnabled: aws.Bool(d.Get("root_volume_encryption_enabled").(bool)),
		UserVolumeEncryptionEnabled: aws.Bool(d.Get("user_volume_encryption_enabled").(bool)),
		Tags:                        Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("volume_encryption_key"); ok {
		input.VolumeEncryptionKey = aws.String(v.(string))
	}

	input.WorkspaceProperties = ExpandWorkspaceProperties(d.Get("workspace_properties").([]interface{}))

	resp, err := conn.CreateWorkspaces(&workspaces.CreateWorkspacesInput{
		Workspaces: []*workspaces.WorkspaceRequest{input},
	})
	if err != nil {
		return fmt.Errorf("creating WorkSpaces Workspace: %w", err)
	}

	wsFail := resp.FailedRequests
	if len(wsFail) > 0 {
		return fmt.Errorf("creating WorkSpaces Workspace: %s: %s", aws.StringValue(wsFail[0].ErrorCode), aws.StringValue(wsFail[0].ErrorMessage))
	}

	workspaceID := aws.StringValue(resp.PendingRequests[0].WorkspaceId)
	d.SetId(workspaceID)

	_, err = WaitWorkspaceAvailable(conn, workspaceID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("creating WorkSpaces Workspace: waiting for completion: %s", err)
	}

	return resourceWorkspaceRead(d, meta)
}

func resourceWorkspaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkSpacesConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	rawOutput, state, err := StatusWorkspaceState(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("reading WorkSpaces Workspace (%s): %s", d.Id(), err)
	}
	if state == workspaces.WorkspaceStateTerminated {
		log.Printf("[WARN] WorkSpaces Workspace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	workspace := rawOutput.(*workspaces.Workspace)
	d.Set("bundle_id", workspace.BundleId)
	d.Set("directory_id", workspace.DirectoryId)
	d.Set("ip_address", workspace.IpAddress)
	d.Set("computer_name", workspace.ComputerName)
	d.Set("state", workspace.State)
	d.Set("root_volume_encryption_enabled", workspace.RootVolumeEncryptionEnabled)
	d.Set("user_name", workspace.UserName)
	d.Set("user_volume_encryption_enabled", workspace.UserVolumeEncryptionEnabled)
	d.Set("volume_encryption_key", workspace.VolumeEncryptionKey)
	if err := d.Set("workspace_properties", FlattenWorkspaceProperties(workspace.WorkspaceProperties)); err != nil {
		return fmt.Errorf("setting workspace properties: %s", err)
	}

	tags, err := ListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("listing tags: %s", err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceWorkspaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkSpacesConn()

	// IMPORTANT: Only one workspace property could be changed in a time.
	// I've create AWS Support feature request to allow multiple properties modification in a time.
	// https://docs.aws.amazon.com/workspaces/latest/adminguide/modify-workspaces.html

	if d.HasChange("workspace_properties.0.compute_type_name") {
		if err := workspacePropertyUpdate("compute_type_name", conn, d); err != nil {
			return fmt.Errorf("updating WorkSpaces Workspace (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("workspace_properties.0.root_volume_size_gib") {
		if err := workspacePropertyUpdate("root_volume_size_gib", conn, d); err != nil {
			return fmt.Errorf("updating WorkSpaces Workspace (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("workspace_properties.0.running_mode") {
		if err := workspacePropertyUpdate("running_mode", conn, d); err != nil {
			return fmt.Errorf("updating WorkSpaces Workspace (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("workspace_properties.0.running_mode_auto_stop_timeout_in_minutes") {
		if err := workspacePropertyUpdate("running_mode_auto_stop_timeout_in_minutes", conn, d); err != nil {
			return fmt.Errorf("updating WorkSpaces Workspace (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("workspace_properties.0.user_volume_size_gib") {
		if err := workspacePropertyUpdate("user_volume_size_gib", conn, d); err != nil {
			return fmt.Errorf("updating WorkSpaces Workspace (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating tags: %s", err)
		}
	}

	return resourceWorkspaceRead(d, meta)
}

func resourceWorkspaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkSpacesConn()

	return WorkspaceDelete(conn, d.Id(), d.Timeout(schema.TimeoutDelete))
}

func WorkspaceDelete(conn *workspaces.WorkSpaces, id string, timeout time.Duration) error {
	resp, err := conn.TerminateWorkspaces(&workspaces.TerminateWorkspacesInput{
		TerminateWorkspaceRequests: []*workspaces.TerminateRequest{
			{
				WorkspaceId: aws.String(id),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("deleting WorkSpaces Workspace (%s): %w", id, err)
	}

	wsFail := resp.FailedRequests
	if len(wsFail) > 0 {
		return fmt.Errorf("deleting WorkSpaces Workspace (%s): %s: %s", id, aws.StringValue(wsFail[0].ErrorCode), aws.StringValue(wsFail[0].ErrorMessage))
	}

	_, err = WaitWorkspaceTerminated(conn, id, timeout)
	if err != nil {
		return fmt.Errorf("deleting WorkSpaces Workspace (%s): waiting for completion: %w", id, err)
	}

	return nil
}

func workspacePropertyUpdate(p string, conn *workspaces.WorkSpaces, d *schema.ResourceData) error {
	id := d.Id()

	var wsp *workspaces.WorkspaceProperties

	switch p {
	case "compute_type_name":
		wsp = &workspaces.WorkspaceProperties{
			ComputeTypeName: aws.String(d.Get("workspace_properties.0.compute_type_name").(string)),
		}
	case "root_volume_size_gib":
		wsp = &workspaces.WorkspaceProperties{
			RootVolumeSizeGib: aws.Int64(int64(d.Get("workspace_properties.0.root_volume_size_gib").(int))),
		}
	case "running_mode":
		wsp = &workspaces.WorkspaceProperties{
			RunningMode: aws.String(d.Get("workspace_properties.0.running_mode").(string)),
		}
	case "running_mode_auto_stop_timeout_in_minutes":
		if d.Get("workspace_properties.0.running_mode") != workspaces.RunningModeAutoStop {
			log.Printf("[DEBUG] Property running_mode_auto_stop_timeout_in_minutes makes sense only for AUTO_STOP running mode")
			return nil
		}

		wsp = &workspaces.WorkspaceProperties{
			RunningModeAutoStopTimeoutInMinutes: aws.Int64(int64(d.Get("workspace_properties.0.running_mode_auto_stop_timeout_in_minutes").(int))),
		}
	case "user_volume_size_gib":
		wsp = &workspaces.WorkspaceProperties{
			UserVolumeSizeGib: aws.Int64(int64(d.Get("workspace_properties.0.user_volume_size_gib").(int))),
		}
	}

	_, err := conn.ModifyWorkspaceProperties(&workspaces.ModifyWorkspacePropertiesInput{
		WorkspaceId:         aws.String(id),
		WorkspaceProperties: wsp,
	})
	if err != nil {
		return fmt.Errorf("modifying property %q: %w", p, err)
	}

	_, err = WaitWorkspaceUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		return fmt.Errorf("modifying property %q: waiting for completion: %w", p, err)
	}

	return nil
}

func ExpandWorkspaceProperties(properties []interface{}) *workspaces.WorkspaceProperties {
	log.Printf("[DEBUG] Expand Workspace properties: %+v ", properties)

	if len(properties) == 0 || properties[0] == nil {
		return nil
	}

	p := properties[0].(map[string]interface{})

	workspaceProperties := &workspaces.WorkspaceProperties{
		ComputeTypeName:   aws.String(p["compute_type_name"].(string)),
		RootVolumeSizeGib: aws.Int64(int64(p["root_volume_size_gib"].(int))),
		RunningMode:       aws.String(p["running_mode"].(string)),
		UserVolumeSizeGib: aws.Int64(int64(p["user_volume_size_gib"].(int))),
	}

	if p["running_mode"].(string) == workspaces.RunningModeAutoStop {
		workspaceProperties.RunningModeAutoStopTimeoutInMinutes = aws.Int64(int64(p["running_mode_auto_stop_timeout_in_minutes"].(int)))
	}

	return workspaceProperties
}

func FlattenWorkspaceProperties(properties *workspaces.WorkspaceProperties) []map[string]interface{} {
	log.Printf("[DEBUG] Flatten workspace properties: %+v ", properties)

	if properties == nil {
		return []map[string]interface{}{}
	}

	return []map[string]interface{}{
		{
			"compute_type_name":                         aws.StringValue(properties.ComputeTypeName),
			"root_volume_size_gib":                      int(aws.Int64Value(properties.RootVolumeSizeGib)),
			"running_mode":                              aws.StringValue(properties.RunningMode),
			"running_mode_auto_stop_timeout_in_minutes": int(aws.Int64Value(properties.RunningModeAutoStopTimeoutInMinutes)),
			"user_volume_size_gib":                      int(aws.Int64Value(properties.UserVolumeSizeGib)),
		},
	}
}
