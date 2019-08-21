package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSagemakerNotebookInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerNotebookInstanceCreate,
		Read:   resourceAwsSagemakerNotebookInstanceRead,
		Update: resourceAwsSagemakerNotebookInstanceUpdate,
		Delete: resourceAwsSagemakerNotebookInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"accelerator_types": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"additional_code_repositories": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"default_code_repository": {
				Type:     schema.TypeString,
				Optional: true,
				// TODO min 1
			},

			"direct_internet_access": {
				// Enum value set: [Enabled, Disabled]
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateSagemakerName,
			},

			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
				// TODO min length 20
			},

			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"security_groups": {
				// TODO is type list in API
				Type:     schema.TypeSet,
				MinItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"lifecycle_config_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateSagemakerName,
			},

			"volume_size": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  5,
			},

			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  sagemaker.NotebookInstanceStatusInService,
			},

			// TODO enum
			"root_access": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSagemakerNotebookInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	createOpts := &sagemaker.CreateNotebookInstanceInput{
		SecurityGroupIds:     expandStringSet(d.Get("security_groups").(*schema.Set)),
		NotebookInstanceName: aws.String(name),
		RoleArn:              aws.String(d.Get("role_arn").(string)),
		InstanceType:         aws.String(d.Get("instance_type").(string)),
	}

	if v, ok := d.GetOk("accelerator_types"); ok {
		acceleratorTypes := expandStringList(v.([]interface{}))
		createOpts.AcceleratorTypes = acceleratorTypes
	}

	if v, ok := d.GetOk("additional_code_repositories"); ok {
		additionalCodeRepositories := expandStringList(v.([]interface{}))
		createOpts.AdditionalCodeRepositories = additionalCodeRepositories
	}

	if v, ok := d.GetOk("default_code_repository"); ok {
		createOpts.DefaultCodeRepository = aws.String(v.(string))
	}

	if v, ok := d.GetOk("direct_internet_access"); ok {
		createOpts.DirectInternetAccess = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		createOpts.KmsKeyId = aws.String(v.(string))
	}

	if l, ok := d.GetOk("lifecycle_config_name"); ok {
		createOpts.LifecycleConfigName = aws.String(l.(string))
	}

	if s, ok := d.GetOk("root_access"); ok {
		createOpts.RootAccess = aws.String(s.(string))
	}

	if s, ok := d.GetOk("subnet_id"); ok {
		createOpts.SubnetId = aws.String(s.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		tagsIn := v.(map[string]interface{})
		createOpts.Tags = tagsFromMapSagemaker(tagsIn)
	}

	if v, ok := d.GetOk("volume_size"); ok {
		createOpts.VolumeSizeInGB = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] SageMaker Notebook Instance create config: %#v", *createOpts)
	_, err := conn.CreateNotebookInstance(createOpts)
	if err != nil {
		return fmt.Errorf("error creating SageMaker Notebook Instance: %s", err)
	}

	d.SetId(name)
	log.Printf("[INFO] SageMaker Notebook Instance ID: %s", d.Id())

	describeInput := &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(name),
	}

	if err := conn.WaitUntilNotebookInstanceInService(describeInput); err != nil {
		return fmt.Errorf("error waiting for  SageMaker Notebook Instance (%s) to be in service: %s", name, err)
	}
	return resourceAwsSagemakerNotebookInstanceRead(d, meta)
}

func resourceAwsSagemakerNotebookInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	}
	notebookInstance, err := conn.DescribeNotebookInstance(describeNotebookInput)
	if err != nil {
		if isAWSErr(err, "ValidationException", "RecordNotFound") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Notebook Instance (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error finding SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("accelerator_types", notebookInstance.AcceleratorTypes); err != nil {
		return fmt.Errorf("error setting accelerator_types for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("additional_code_repositories", notebookInstance.AdditionalCodeRepositories); err != nil {
		return fmt.Errorf("error setting additional_code_repositories for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("default_code_repository", notebookInstance.DefaultCodeRepository); err != nil {
		return fmt.Errorf("error setting default_code_repository for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("direct_internet_access", notebookInstance.DirectInternetAccess); err != nil {
		return fmt.Errorf("error setting direct_internet_access for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("instance_type", notebookInstance.InstanceType); err != nil {
		return fmt.Errorf("error setting instance_type for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("name", notebookInstance.NotebookInstanceName); err != nil {
		return fmt.Errorf("error setting name for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("role_arn", notebookInstance.RoleArn); err != nil {
		return fmt.Errorf("error setting role_arn for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("root_access", notebookInstance.RootAccess); err != nil {
		return fmt.Errorf("error setting root_access for sSageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("security_groups", flattenStringList(notebookInstance.SecurityGroups)); err != nil {
		return fmt.Errorf("error setting security groups for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("subnet_id", notebookInstance.SubnetId); err != nil {
		return fmt.Errorf("error setting subnet_id for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("kms_key_id", notebookInstance.KmsKeyId); err != nil {
		return fmt.Errorf("error setting kms_key_id for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("lifecycle_config_name", notebookInstance.NotebookInstanceLifecycleConfigName); err != nil {
		return fmt.Errorf("error setting lifecycle_config_name for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("arn", notebookInstance.NotebookInstanceArn); err != nil {
		return fmt.Errorf("error setting arn for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("volume_size", notebookInstance.VolumeSizeInGB); err != nil {
		return fmt.Errorf("error setting volume_size for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("status", notebookInstance.NotebookInstanceStatus); err != nil {
		return fmt.Errorf("error setting status for SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	// FailureReason
	// NetworkInterfaceId
	// Url

	tagsOutput, err := conn.ListTags(&sagemaker.ListTagsInput{
		ResourceArn: notebookInstance.NotebookInstanceArn,
	})
	if err != nil {
		return fmt.Errorf("error listing tags for sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tagsToMapSagemaker(tagsOutput.Tags)); err != nil {
		return fmt.Errorf("error setting tags for notebook instance (%s): %s", d.Id(), err)
	}
	return nil
}

func resourceAwsSagemakerNotebookInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	d.Partial(true)

	if err := setSagemakerTags(conn, d); err != nil {
		return err
	}
	d.SetPartial("tags")

	hasChanged := false
	// Update
	updateOpts := &sagemaker.UpdateNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("accelerator_types") {
		//updateOpts.SetAcceleratorTypes(d.Get("accelerator_types").(string))
		hasChanged = true
		// TODO set DisassociateAcceleratorTypes to true
	}

	if d.HasChange("additional_code_repositories") {
		//updateOpts.SetAdditionalCodeRepositories(d.Get("additional_code_repositories").(string))
		hasChanged = true
	}

	if d.HasChange("default_code_repository") {
		updateOpts.SetDefaultCodeRepository(d.Get("default_code_repository").(string))
		hasChanged = true
	}

	if d.HasChange("role_arn") {
		updateOpts.SetRoleArn(d.Get("role_arn").(string))
		hasChanged = true
	}

	if d.HasChange("instance_type") {
		updateOpts.SetInstanceType(d.Get("instance_type").(string))
		hasChanged = true
	}

	if d.HasChange("lifecycle_config_name") {
		updateOpts.SetLifecycleConfigName(d.Get("lifecycle_config_name").(string))
		hasChanged = true
		// TODO set flag to true
	}

	if d.HasChange("volume_size") {
		updateOpts.SetVolumeSizeInGB(d.Get("volume_size").(int64))
		hasChanged = true
	}

	if d.HasChange("status") {
		_, status, _ := sagemakerNotebookInstanceStateRefreshFunc(conn, d.Id())()
		if status == sagemaker.NotebookInstanceStatusStopped ||
			status == sagemaker.NotebookInstanceStatusFailed {
			log.Printf("[INFO] Starting Sagemaker Notebook Instance %q", d.Id())
			_, err := conn.StartNotebookInstance(&sagemaker.StartNotebookInstanceInput{
				NotebookInstanceName: aws.String(d.Id()),
			})
			if err != nil {
				return fmt.Errorf("error starting Sagemaker Notebook Instance (%s): %s", d.Id(), err)
			}

			describeInput := &sagemaker.DescribeNotebookInstanceInput{
				NotebookInstanceName: aws.String(d.Id()),
			}

			if err := conn.WaitUntilNotebookInstanceInService(describeInput); err != nil {
				return fmt.Errorf("error waiting for SageMaker Notebook Instance (%s) to be in service: %s", d.Id(), err)
			}
		}
	}

	if hasChanged {
		// Stop notebook
		_, previousStatus, _ := sagemakerNotebookInstanceStateRefreshFunc(conn, d.Id())()
		if previousStatus != sagemaker.NotebookInstanceStatusStopped {
			if err := stopSagemakerNotebookInstance(conn, d.Id()); err != nil {
				return fmt.Errorf("error stopping sagemaker notebook instance prior to updating: %s", err)
			}
		}

		if _, err := conn.UpdateNotebookInstance(updateOpts); err != nil {
			return fmt.Errorf("error updating sagemaker notebook instance: %s", err)
		}

		stateConf := &resource.StateChangeConf{
			Pending: []string{
				sagemaker.NotebookInstanceStatusUpdating,
			},
			Target:  []string{sagemaker.NotebookInstanceStatusStopped},
			Refresh: sagemakerNotebookInstanceStateRefreshFunc(conn, d.Id()),
			Timeout: 10 * time.Minute,
		}
		_, err := stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to update: %s", d.Id(), err)
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
				Target:  []string{sagemaker.NotebookInstanceStatusInService, sagemaker.NotebookInstanceStatusPending},
				Refresh: sagemakerNotebookInstanceStateRefreshFunc(conn, d.Id()),
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
			if isResourceTimeoutError(err) {
				_, err = conn.StartNotebookInstance(startOpts)
				if err != nil {
					return fmt.Errorf("error starting sagemaker notebook instance (%s): %s", d.Id(), err)
				}

				_, err = stateConf.WaitForState()
			}
			if err != nil {
				return fmt.Errorf("Error waiting for sagemaker notebook instance to start: %s", err)
			}

			describeInput := &sagemaker.DescribeNotebookInstanceInput{
				NotebookInstanceName: aws.String(d.Id()),
			}

			if err := conn.WaitUntilNotebookInstanceInService(describeInput); err != nil {
				return fmt.Errorf("error waiting for SageMaker Notebook Instance (%s) to be in service: %s", d.Id(), err)
			}
		}
	}

	d.Partial(false)

	return resourceAwsSagemakerNotebookInstanceRead(d, meta)
}

func resourceAwsSagemakerNotebookInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	}
	notebook, err := conn.DescribeNotebookInstance(describeNotebookInput)
	if err != nil {
		if isAWSErr(err, "ValidationException", "RecordNotFound") {
			return nil
		}
		return fmt.Errorf("unable to find sagemaker notebook instance to delete (%s): %s", d.Id(), err)
	}
	if *notebook.NotebookInstanceStatus != sagemaker.NotebookInstanceStatusFailed && *notebook.NotebookInstanceStatus != sagemaker.NotebookInstanceStatusStopped {
		if err := stopSagemakerNotebookInstance(conn, d.Id()); err != nil {
			return err
		}
	}

	deleteOpts := &sagemaker.DeleteNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteNotebookInstance(deleteOpts); err != nil {
		return fmt.Errorf("error trying to delete sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	describeInput := &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	}

	if err := conn.WaitUntilNotebookInstanceDeleted(describeInput); err != nil {
		return fmt.Errorf("error waiting for SageMaker Notebook Instance (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func stopSagemakerNotebookInstance(conn *sagemaker.SageMaker, id string) error {
	describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(id),
	}
	notebook, err := conn.DescribeNotebookInstance(describeNotebookInput)
	if err != nil {
		if isAWSErr(err, "ValidationException", "RecordNotFound") {
			return nil
		}
		return fmt.Errorf("unable to find sagemaker notebook instance (%s): %s", id, err)
	}
	if *notebook.NotebookInstanceStatus == sagemaker.NotebookInstanceStatusStopped {
		return nil
	}

	stopOpts := &sagemaker.StopNotebookInstanceInput{
		NotebookInstanceName: aws.String(id),
	}

	if _, err := conn.StopNotebookInstance(stopOpts); err != nil {
		return fmt.Errorf("error stopping sagemaker notebook instance: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			sagemaker.NotebookInstanceStatusStopping,
		},
		Target:  []string{sagemaker.NotebookInstanceStatusStopped},
		Refresh: sagemakerNotebookInstanceStateRefreshFunc(conn, id),
		Timeout: 10 * time.Minute,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to stop: %s", id, err)
	}

	return nil
}

func sagemakerNotebookInstanceStateRefreshFunc(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(name),
		}
		notebook, err := conn.DescribeNotebookInstance(describeNotebookInput)
		if err != nil {
			if isAWSErr(err, "ValidationException", "RecordNotFound") {
				return "", "", nil
			}
			return nil, "", err
		}

		if notebook == nil {
			return nil, "", nil
		}

		return notebook, *notebook.NotebookInstanceStatus, nil
	}
}
