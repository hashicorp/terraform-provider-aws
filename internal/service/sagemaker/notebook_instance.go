package sagemaker

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceNotebookInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceNotebookInstanceCreate,
		Read:   resourceNotebookInstanceRead,
		Update: resourceNotebookInstanceUpdate,
		Delete: resourceNotebookInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("volume_size", func(_ context.Context, old, new, meta interface{}) bool {
				return new.(int) < old.(int)
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"additional_code_repositories": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 3,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_code_repository": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"direct_internet_access": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      sagemaker.DirectInternetAccessEnabled,
				ValidateFunc: validation.StringInSlice(sagemaker.DirectInternetAccess_Values(), false),
			},
			"instance_metadata_service_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"minimum_instance_metadata_service_version": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"1", "2"}, false),
						},
					},
				},
			},
			"instance_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.InstanceType_Values(), false),
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"lifecycle_config_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(notebook-al1-v1|notebook-al2-v1|notebook-al2-v2)$`), ""),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"root_access": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      sagemaker.RootAccessEnabled,
				ValidateFunc: validation.StringInSlice(sagemaker.RootAccess_Values(), false),
			},
			"security_groups": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  5,
			},
		},
	}
}

func resourceNotebookInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	createOpts := &sagemaker.CreateNotebookInstanceInput{
		SecurityGroupIds:                     flex.ExpandStringSet(d.Get("security_groups").(*schema.Set)),
		NotebookInstanceName:                 aws.String(name),
		RoleArn:                              aws.String(d.Get("role_arn").(string)),
		InstanceType:                         aws.String(d.Get("instance_type").(string)),
		InstanceMetadataServiceConfiguration: expandNotebookInstanceMetadataServiceConfiguration(d.Get("instance_metadata_service_configuration").([]interface{})),
	}

	if v, ok := d.GetOk("root_access"); ok {
		createOpts.RootAccess = aws.String(v.(string))
	}

	if v, ok := d.GetOk("platform_identifier"); ok {
		createOpts.PlatformIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("direct_internet_access"); ok {
		createOpts.DirectInternetAccess = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_code_repository"); ok {
		createOpts.DefaultCodeRepository = aws.String(v.(string))
	}

	if s, ok := d.GetOk("subnet_id"); ok {
		createOpts.SubnetId = aws.String(s.(string))
	}

	if v, ok := d.GetOk("volume_size"); ok {
		createOpts.VolumeSizeInGB = aws.Int64(int64(v.(int)))
	}

	if k, ok := d.GetOk("kms_key_id"); ok {
		createOpts.KmsKeyId = aws.String(k.(string))
	}

	if l, ok := d.GetOk("lifecycle_config_name"); ok {
		createOpts.LifecycleConfigName = aws.String(l.(string))
	}

	if len(tags) > 0 {
		createOpts.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("additional_code_repositories"); ok && v.(*schema.Set).Len() > 0 {
		createOpts.AdditionalCodeRepositories = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] sagemaker notebook instance create config: %#v", *createOpts)
	_, err := conn.CreateNotebookInstance(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating SageMaker Notebook Instance: %s", err)
	}

	d.SetId(name)
	log.Printf("[INFO] sagemaker notebook instance ID: %s", d.Id())

	if _, err := WaitNotebookInstanceInService(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to create: %w", d.Id(), err)
	}

	return resourceNotebookInstanceRead(d, meta)
}

func resourceNotebookInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	notebookInstance, err := FindNotebookInstanceByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Notebook Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SageMaker Notebook Instance (%s): %w", d.Id(), err)
	}

	d.Set("name", notebookInstance.NotebookInstanceName)
	d.Set("role_arn", notebookInstance.RoleArn)
	d.Set("instance_type", notebookInstance.InstanceType)
	d.Set("platform_identifier", notebookInstance.PlatformIdentifier)
	d.Set("subnet_id", notebookInstance.SubnetId)
	d.Set("kms_key_id", notebookInstance.KmsKeyId)
	d.Set("volume_size", notebookInstance.VolumeSizeInGB)
	d.Set("lifecycle_config_name", notebookInstance.NotebookInstanceLifecycleConfigName)
	d.Set("arn", notebookInstance.NotebookInstanceArn)
	d.Set("root_access", notebookInstance.RootAccess)
	d.Set("direct_internet_access", notebookInstance.DirectInternetAccess)
	d.Set("default_code_repository", notebookInstance.DefaultCodeRepository)
	d.Set("url", notebookInstance.Url)
	d.Set("network_interface_id", notebookInstance.NetworkInterfaceId)
	d.Set("additional_code_repositories", flex.FlattenStringSet(notebookInstance.AdditionalCodeRepositories))
	d.Set("security_groups", flex.FlattenStringList(notebookInstance.SecurityGroups))

	if err := d.Set("instance_metadata_service_configuration", flattenNotebookInstanceMetadataServiceConfiguration(notebookInstance.InstanceMetadataServiceConfiguration)); err != nil {
		return fmt.Errorf("error setting instance_metadata_service_configuration for sagemaker notebook instance (%s): %w", d.Id(), err)
	}

	tags, err := ListTags(conn, aws.StringValue(notebookInstance.NotebookInstanceArn))

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Notebook Instance (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceNotebookInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker Notebook Instance (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChangesExcept("tags", "tags_all") {
		updateOpts := &sagemaker.UpdateNotebookInstanceInput{
			NotebookInstanceName: aws.String(d.Get("name").(string)),
		}

		if d.HasChange("role_arn") {
			updateOpts.RoleArn = aws.String(d.Get("role_arn").(string))
		}

		if d.HasChange("instance_type") {
			updateOpts.InstanceType = aws.String(d.Get("instance_type").(string))
		}

		if d.HasChange("volume_size") {
			updateOpts.VolumeSizeInGB = aws.Int64(int64(d.Get("volume_size").(int)))
		}

		if d.HasChange("lifecycle_config_name") {
			if v, ok := d.GetOk("lifecycle_config_name"); ok {
				updateOpts.LifecycleConfigName = aws.String(v.(string))
			} else {
				updateOpts.DisassociateLifecycleConfig = aws.Bool(true)
			}
		}

		if d.HasChange("default_code_repository") {
			if v, ok := d.GetOk("default_code_repository"); ok {
				updateOpts.DefaultCodeRepository = aws.String(v.(string))
			} else {
				updateOpts.DisassociateDefaultCodeRepository = aws.Bool(true)
			}
		}

		if d.HasChange("instance_metadata_service_configuration") {
			updateOpts.InstanceMetadataServiceConfiguration = expandNotebookInstanceMetadataServiceConfiguration(d.Get("instance_metadata_service_configuration").([]interface{}))
		}

		if d.HasChange("root_access") {
			updateOpts.RootAccess = aws.String(d.Get("root_access").(string))
		}

		if d.HasChange("additional_code_repositories") {
			if v, ok := d.GetOk("additional_code_repositories"); ok {
				updateOpts.AdditionalCodeRepositories = flex.ExpandStringSet(v.(*schema.Set))
			} else {
				updateOpts.DisassociateAdditionalCodeRepositories = aws.Bool(true)
			}
		}

		// Stop notebook
		_, previousStatus, _ := StatusNotebookInstance(conn, d.Id())()
		if previousStatus != sagemaker.NotebookInstanceStatusStopped {
			if err := StopNotebookInstance(conn, d.Id()); err != nil {
				return fmt.Errorf("error stopping sagemaker notebook instance prior to updating: %s", err)
			}
		}

		if _, err := conn.UpdateNotebookInstance(updateOpts); err != nil {
			return fmt.Errorf("error updating sagemaker notebook instance: %s", err)
		}

		if _, err := WaitNotebookInstanceStopped(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to stop: %w", d.Id(), err)
		}

		// Restart if needed
		if previousStatus == sagemaker.NotebookInstanceStatusInService {
			startOpts := &sagemaker.StartNotebookInstanceInput{
				NotebookInstanceName: aws.String(d.Id()),
			}
			stateConf := &resource.StateChangeConf{
				Pending: []string{
					sagemaker.NotebookInstanceStatusStopped,
				},
				Target: []string{
					sagemaker.NotebookInstanceStatusInService,
					sagemaker.NotebookInstanceStatusPending,
				},
				Refresh: StatusNotebookInstance(conn, d.Id()),
				Timeout: 30 * time.Second,
			}
			// StartNotebookInstance sometimes doesn't take so we'll check for a state change and if
			// it doesn't change we'll send another request
			err := resource.Retry(5*time.Minute, func() *resource.RetryError {
				_, err := conn.StartNotebookInstance(startOpts)
				if err != nil {
					return resource.NonRetryableError(fmt.Errorf("error starting sagemaker notebook instance (%s): %s", d.Id(), err))
				}

				_, err = stateConf.WaitForState()
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error waiting for sagemaker notebook instance (%s) to start: %s", d.Id(), err))
				}

				return nil
			})
			if tfresource.TimedOut(err) {
				_, err = conn.StartNotebookInstance(startOpts)
				if err != nil {
					return fmt.Errorf("error starting sagemaker notebook instance (%s): %s", d.Id(), err)
				}

				_, err = stateConf.WaitForState()
			}
			if err != nil {
				return fmt.Errorf("Error waiting for sagemaker notebook instance to start: %s", err)
			}

			if _, err := WaitNotebookInstanceInService(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to to start after update: %w", d.Id(), err)
			}
		}
	}

	return resourceNotebookInstanceRead(d, meta)
}

func resourceNotebookInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SageMakerConn

	notebook, err := FindNotebookInstanceByName(conn, d.Id())

	if err != nil {
		if tfresource.NotFound(err) {
			return nil
		}
		return fmt.Errorf("error reading SageMaker Notebook Instance (%s): %w", d.Id(), err)
	}

	if aws.StringValue(notebook.NotebookInstanceStatus) != sagemaker.NotebookInstanceStatusFailed &&
		aws.StringValue(notebook.NotebookInstanceStatus) != sagemaker.NotebookInstanceStatusStopped {
		if err := StopNotebookInstance(conn, d.Id()); err != nil {
			return err
		}
	}

	deleteOpts := &sagemaker.DeleteNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteNotebookInstance(deleteOpts); err != nil {
		return fmt.Errorf("error trying to delete sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	if _, err := WaitNotebookInstanceDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func StopNotebookInstance(conn *sagemaker.SageMaker, id string) error {
	notebook, err := FindNotebookInstanceByName(conn, id)

	if err != nil {
		if tfresource.NotFound(err) {
			return nil
		}
		return fmt.Errorf("error reading SageMaker Notebook Instance (%s): %w", id, err)
	}

	if aws.StringValue(notebook.NotebookInstanceStatus) == sagemaker.NotebookInstanceStatusStopped {
		return nil
	}

	stopOpts := &sagemaker.StopNotebookInstanceInput{
		NotebookInstanceName: aws.String(id),
	}

	if _, err := conn.StopNotebookInstance(stopOpts); err != nil {
		return fmt.Errorf("Error stopping sagemaker notebook instance: %s", err)
	}

	if _, err := WaitNotebookInstanceStopped(conn, id); err != nil {
		return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to stop: %w", id, err)
	}

	return nil
}

func expandNotebookInstanceMetadataServiceConfiguration(l []interface{}) *sagemaker.InstanceMetadataServiceConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.InstanceMetadataServiceConfiguration{
		MinimumInstanceMetadataServiceVersion: aws.String(m["minimum_instance_metadata_service_version"].(string)),
	}

	return config
}

func flattenNotebookInstanceMetadataServiceConfiguration(config *sagemaker.InstanceMetadataServiceConfiguration) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"minimum_instance_metadata_service_version": aws.StringValue(config.MinimumInstanceMetadataServiceVersion),
	}

	return []map[string]interface{}{m}
}
