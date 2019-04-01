package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsSagemakerLifeCycleConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerLifeCycleConfigurationCreate,
		Read:   resourceAwsSagemakerLifeCycleConfigurationRead,
		Update: resourceAwsSagemakerLifeCycleConfigurationUpdate,
		Delete: resourceAwsSagemakerLifeCycleConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"on_create": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},

			"on_start": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
		},
	}
}

func resourceAwsSagemakerLifeCycleConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	createOpts := &sagemaker.CreateNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(name),
	}

	// on_create is technically a list of NotebookInstanceLifecycleHook elements, but the list has to be length 1
	// (same for on_start)
	if v, ok := d.GetOk("on_create"); ok {
		hook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		createOpts.SetOnCreate([]*sagemaker.NotebookInstanceLifecycleHook{hook})
	}

	if v, ok := d.GetOk("on_start"); ok {
		hook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		createOpts.SetOnStart([]*sagemaker.NotebookInstanceLifecycleHook{hook})
	}

	log.Printf("[DEBUG] SageMaker notebook instance lifecycle configuration create config: %#v", *createOpts)
	_, err := conn.CreateNotebookInstanceLifecycleConfig(createOpts)
	if err != nil {
		return fmt.Errorf("error creating SageMaker notebook instance lifecycle configuration: %s", err)
	}
	d.SetId(name)

	return resourceAwsSagemakerLifeCycleConfigurationRead(d, meta)
}

func resourceAwsSagemakerLifeCycleConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	request := &sagemaker.DescribeNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(d.Id()),
	}

	lifecycleConfig, err := conn.DescribeNotebookInstanceLifecycleConfig(request)
	if err != nil {
		if isAWSErr(err, "ValidationException", "") {
			log.Printf("[INFO] unable to find the SageMaker notebook instance lifecycle configuration (%s); therefore it is removed from the state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading SageMaker notebook instance lifecycle configuration %s: %s", d.Id(), err)
	}

	if err := d.Set("name", lifecycleConfig.NotebookInstanceLifecycleConfigName); err != nil {
		return fmt.Errorf("error setting name for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
	}

	if len(lifecycleConfig.OnCreate) > 0 {
		if err := d.Set("on_create", lifecycleConfig.OnCreate[0].Content); err != nil {
			return fmt.Errorf("error setting on_create for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
		}
	}

	if len(lifecycleConfig.OnStart) > 0 {
		if err := d.Set("on_start", lifecycleConfig.OnStart[0].Content); err != nil {
			return fmt.Errorf("error setting on_start for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
		}
	}

	if err := d.Set("arn", lifecycleConfig.NotebookInstanceLifecycleConfigArn); err != nil {
		return fmt.Errorf("error setting arn for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsSagemakerLifeCycleConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	updateOpts := &sagemaker.UpdateNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(d.Get("name").(string)),
	}

	onCreateHook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(d.Get("on_create").(string))}
	updateOpts.SetOnCreate([]*sagemaker.NotebookInstanceLifecycleHook{onCreateHook})

	onStartHook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(d.Get("on_start").(string))}
	updateOpts.SetOnStart([]*sagemaker.NotebookInstanceLifecycleHook{onStartHook})

	_, err := conn.UpdateNotebookInstanceLifecycleConfig(updateOpts)
	if err != nil {
		return fmt.Errorf("error updating SageMaker notebook instance lifecycle configuration: %s", err)
	}
	return resourceAwsSagemakerLifeCycleConfigurationRead(d, meta)
}

func resourceAwsSagemakerLifeCycleConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	deleteOpts := &sagemaker.DeleteNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker notebook instance lifecycle configuration: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteNotebookInstanceLifecycleConfig(deleteOpts)
		if err == nil {
			return nil
		}

		if isAWSErr(err, "ValidationException", "") {
			return nil
		}

		return resource.NonRetryableError(fmt.Errorf("error deleting SageMaker notebook instance lifecycle configuration: %s", err))
	})
}
