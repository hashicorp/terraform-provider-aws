package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Computed:     false,
				ForceNew:     true,
				ValidateFunc: validateSagemakerName,
			},

			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
				Computed: false,
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

			"creation_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSagemakerNotebookInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	name := d.Get("name").(string)

	createOpts := &sagemaker.CreateNotebookInstanceInput{
		NotebookInstanceName: aws.String(name),
		RoleArn:              aws.String(d.Get("role_arn").(string)),
		InstanceType:         aws.String(d.Get("instance_type").(string)),
	}

	if s, ok := d.GetOk("subnet_id"); ok {
		createOpts.SubnetId = aws.String(s.(string))
	}

	if sgs, ok := d.GetOk("security_groups"); ok {
		sgList := sgs.(*schema.Set).List()
		var groups []*string
		for _, v := range sgList {
			groups = append(groups, aws.String(v.(string)))
		}
		createOpts.SecurityGroupIds = groups
	}

	if k, ok := d.GetOk("kms_key_id"); ok {
		createOpts.KmsKeyId = aws.String(k.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		tagsIn := v.(map[string]interface{})
		createOpts.Tags = tagsFromMapSagemaker(tagsIn)
	}

	log.Printf("[DEBUG] Sagemaker Notebook Instance create config: %#v", *createOpts)
	createResponse, err := conn.CreateNotebookInstance(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Sagemaker Notebook Instance: %s", err)
	}

	d.SetId(name)
	if err := d.Set("arn", createResponse.NotebookInstanceArn); err != nil {
		return err
	}
	log.Printf("[INFO] Sagemaker Notebook Instance ID: %s", d.Id())

	if err := waitSagemakerNotebookInstanceStatus(conn, d, "InService", "Failed"); err != nil {
		log.Printf("[ERR] Sagemaker Notebook Instance (%s) did not start", d.Id())
	}

	return resourceAwsSagemakerNotebookInstanceRead(d, meta)
}

func resourceAwsSagemakerNotebookInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	notebookInstanceRaw, _, err := SagemakerNotebookInstanceStateRefreshFunc(conn, d.Id())()
	if err != nil {
		return err
	}
	if notebookInstanceRaw == nil {
		d.SetId("")
		return nil
	}

	notebookInstance := notebookInstanceRaw.(*sagemaker.DescribeNotebookInstanceOutput)

	if err := d.Set("name", notebookInstance.NotebookInstanceName); err != nil {
		log.Printf("[ERR] Error setting Name: %s", err)
		return err
	}
	if err := d.Set("role_arn", notebookInstance.RoleArn); err != nil {
		log.Printf("[ERR] Error setting Role ARN: %s", err)
		return err
	}
	if err := d.Set("instance_type", notebookInstance.InstanceType); err != nil {
		log.Printf("[ERR] Error setting Instance Type: %s", err)
		return err
	}
	if err := d.Set("subnet_id", notebookInstance.SubnetId); err != nil {
		log.Printf("[ERR] Error setting Subnet ID: %s", err)
		return err
	}
	sgs := notebookInstance.SecurityGroups
	if sgs != nil && len(sgs) > 0 {
		if err := d.Set("security_groups", fromStringPSliceToStringSliceSagemaker(&sgs)); err != nil {
			log.Printf("[ERR] Error setting Security Groups: %s", err)
			return err
		}
	}
	if err := d.Set("kms_key_id", notebookInstance.KmsKeyId); err != nil {
		log.Printf("[ERR] Error setting KMS Key ID: %s", err)
		return err
	}
	if err := d.Set("creation_time", notebookInstance.CreationTime.Format(time.RFC3339)); err != nil {
		log.Printf("[ERR] Error setting Creation Time: %s", err)
		return err
	}
	if err := d.Set("last_modified_time", notebookInstance.LastModifiedTime.Format(time.RFC3339)); err != nil {
		log.Printf("[ERR] Error setting Last Modified Time: %s", err)
		return err
	}
	if err := d.Set("status", notebookInstance.NotebookInstanceStatus); err != nil {
		log.Printf("[ERR] Error setting Status: %s", err)
		return err
	}
	if err := d.Set("arn", notebookInstance.NotebookInstanceArn); err != nil {
		log.Printf("[ERR] Error setting ARN: %s", err)
		return err
	}

	tagsOutput, err := conn.ListTags(&sagemaker.ListTagsInput{
		ResourceArn: notebookInstance.NotebookInstanceArn,
	})

	d.Set("tags", tagsToMapSagemaker(tagsOutput.Tags))

	return nil
}

func resourceAwsSagemakerNotebookInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	d.Partial(true)

	if err := setSagemakerTags(conn, d); err != nil {
		return err
	} else {
		d.SetPartial("tags")
	}

	// Once a nootebook is created only the instance type and role can be modified
	if d.HasChange("subnet_id") || d.HasChange("security_groups") || d.HasChange("kms_key_id") {
		return fmt.Errorf("Error existing Notebook Instance cannot be updated")
	}

	hasChanged := false
	// Update
	updateOpts := &sagemaker.UpdateNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("role_arn") {
		updateOpts.RoleArn = aws.String(d.Get("role_arn").(string))
		hasChanged = true
	}

	if d.HasChange("instance_type") {
		updateOpts.InstanceType = aws.String(d.Get("instance_type").(string))
		hasChanged = true
	}

	if hasChanged {
		// Stop notebook
		_, previousStatus, _ := SagemakerNotebookInstanceStateRefreshFunc(conn, d.Id())()
		if err := stopSagemakerNotebookInstance(conn, d, true); err != nil {
			return fmt.Errorf("Error stopping Sagemaker Notebook Instance: %s", err)
		}

		if _, err := conn.UpdateNotebookInstance(updateOpts); err != nil {
			return fmt.Errorf("Error updating Sagemaker Notebook Instance: %s", err)
		}

		// Restart if needed
		if previousStatus == "InService" {
			startOpts := &sagemaker.StartNotebookInstanceInput{
				NotebookInstanceName: aws.String(d.Id()),
			}

			if _, err := conn.StartNotebookInstance(startOpts); err != nil {
				log.Printf("[DEBUG] Sagemaker Notebook Instance (%s) unable to restart", d.Id())
			} else if err := waitSagemakerNotebookInstanceStatus(conn, d, "InService", "Failed"); err != nil {
				log.Printf("[ERR] Sagemaker Notebook Instance (%s) did not start", d.Id())
			}
		}
	}

	d.Partial(false)

	return resourceAwsSagemakerNotebookInstanceRead(d, meta)
}

func resourceAwsSagemakerNotebookInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if err := stopSagemakerNotebookInstance(conn, d, true); err != nil {
		return err
	}

	deleteOpts := &sagemaker.DeleteNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteNotebookInstance(deleteOpts); err != nil {
		return err
	}

	return resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, status, _ := SagemakerNotebookInstanceStateRefreshFunc(conn, d.Id())()

		if status == "" {
			log.Printf("[DEBUG] Sagemaker Notebook Instance (%s) deleted", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("[DEBUG] Waiting for Sagemaker Notebook Instance (%s) to be deleted", d.Id()))
	})
}

func SagemakerNotebookInstanceStateRefreshFunc(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(name),
		}
		notebook, err := conn.DescribeNotebookInstance(describeNotebookInput)
		if err != nil {
			if sagemakerErr, ok := err.(awserr.Error); ok && sagemakerErr.Code() == "ResourceNotFound" {
				notebook = nil
			} else {
				log.Printf("Error on SagemakerNotebookInstanceStateRefreshFunc: %s", err)
				return nil, "", err
			}
		}

		if notebook == nil {
			return nil, "", nil
		}

		return notebook, *notebook.NotebookInstanceStatus, nil
	}
}

func stopSagemakerNotebookInstance(conn *sagemaker.SageMaker, d *schema.ResourceData, wait bool) error {
	stopOpts := &sagemaker.StopNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	}

	if _, err := conn.StopNotebookInstance(stopOpts); err != nil {
		return fmt.Errorf("Error stopping Sagemaker Notebook Instance: %s", err)
	}

	if !wait {
		return nil
	}

	return waitSagemakerNotebookInstanceStatus(conn, d, "Stopped")
}

func waitSagemakerNotebookInstanceStatus(conn *sagemaker.SageMaker, d *schema.ResourceData, desiredStatus ...string) error {
	return resource.Retry(10*time.Minute, func() *resource.RetryError {
		_, status, err := SagemakerNotebookInstanceStateRefreshFunc(conn, d.Id())()

		if err == nil {
			if status == "" {
				log.Printf("[DEBUG] Sagemaker Notebook Instance (%s) not found", d.Id())
				return nil
			}

			for _, s := range desiredStatus {
				if status == s {
					log.Printf("[DEBUG] Sagemaker Notebook Instance (%s) is %s", d.Id(), s)
					return nil
				}
			}
		}

		return resource.RetryableError(fmt.Errorf("[DEBUG] Waiting for Sagemaker Notebook Instance (%s) to be %s", d.Id(), desiredStatus))
	})
}

func fromStringPSliceToStringSliceSagemaker(sgs *[]*string) *[]string {
	result := make([]string, 0, len(*sgs))
	for _, sg := range *sgs {
		result = append(result, *sg)
	}
	return &result
}
