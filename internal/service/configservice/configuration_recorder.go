package configservice

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceConfigurationRecorder() *schema.Resource {
	return &schema.Resource{
		Create: resourceConfigurationRecorderPut,
		Read:   resourceConfigurationRecorderRead,
		Update: resourceConfigurationRecorderPut,
		Delete: resourceConfigurationRecorderDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "default",
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"recording_group": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"all_supported": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_global_resource_types": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"resource_types": {
							Type:     schema.TypeSet,
							Set:      schema.HashString,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceConfigurationRecorderPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	name := d.Get("name").(string)
	recorder := configservice.ConfigurationRecorder{
		Name:    aws.String(name),
		RoleARN: aws.String(d.Get("role_arn").(string)),
	}

	if g, ok := d.GetOk("recording_group"); ok {
		recorder.RecordingGroup = expandRecordingGroup(g.([]interface{}))
	}

	input := configservice.PutConfigurationRecorderInput{
		ConfigurationRecorder: &recorder,
	}
	_, err := conn.PutConfigurationRecorder(&input)
	if err != nil {
		return fmt.Errorf("Creating Configuration Recorder failed: %s", err)
	}

	d.SetId(name)

	return resourceConfigurationRecorderRead(d, meta)
}

func resourceConfigurationRecorderRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn

	input := configservice.DescribeConfigurationRecordersInput{
		ConfigurationRecorderNames: []*string{aws.String(d.Id())},
	}
	out, err := conn.DescribeConfigurationRecorders(&input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigurationRecorderException) {
		names.LogNotFoundRemoveState(names.ConfigService, names.ErrActionReading, "Configuration Recorder", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.ConfigService, names.ErrActionReading, "Configuration Recorder", d.Id(), err)
	}

	numberOfRecorders := len(out.ConfigurationRecorders)
	if !d.IsNewResource() && numberOfRecorders < 1 {
		names.LogNotFoundRemoveState(names.ConfigService, names.ErrActionReading, "Configuration Recorder", d.Id())
		d.SetId("")
		return nil
	}

	if d.IsNewResource() && numberOfRecorders < 1 {
		return names.Error(names.ConfigService, names.ErrActionReading, "Configuration Recorder", d.Id(), errors.New("none found"))
	}

	if numberOfRecorders > 1 {
		return fmt.Errorf("Expected exactly 1 Configuration Recorder, received %d: %#v",
			numberOfRecorders, out.ConfigurationRecorders)
	}

	recorder := out.ConfigurationRecorders[0]

	d.Set("name", recorder.Name)
	d.Set("role_arn", recorder.RoleARN)

	if recorder.RecordingGroup != nil {
		flattened := flattenRecordingGroup(recorder.RecordingGroup)
		err = d.Set("recording_group", flattened)
		if err != nil {
			return fmt.Errorf("Failed to set recording_group: %s", err)
		}
	}

	return nil
}

func resourceConfigurationRecorderDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ConfigServiceConn
	input := configservice.DeleteConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String(d.Id()),
	}
	_, err := conn.DeleteConfigurationRecorder(&input)
	if err != nil {
		if !tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigurationRecorderException) {
			return fmt.Errorf("Deleting Configuration Recorder failed: %s", err)
		}
	}
	return nil
}
