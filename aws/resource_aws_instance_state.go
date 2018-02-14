package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsInstanceState() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsInstanceStateCreate,
		Read:   resourceAwsInstanceStateRead,
		Update: resourceAwsInstanceStateUpdate,
		Delete: resourceAwsInstanceStateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"InstanceId": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
			},
			"State": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"Force": {
				Type:     schema.TypeBool,
				Required: false,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsInstanceStateCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsInstanceStateUpdate(d, meta)
}

func resourceAwsInstanceStateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	id := d.Get("InstanceId").(string)
	state, err := awsInStateReadInstance(conn, id)

	if err != nil {
		return err
	}

	d.Set("State", state)

	return nil
}

func resourceAwsInstanceStateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	state := d.Get("State").(string)
	id := d.Get("InstanceId").(string)

	currentState, err := awsInStateReadInstance(conn, id)

	if err != nil {
		return err
	}

	if currentState == state {
		return nil
	}

	switch state {
	case "stopped":
		return awsStateStopInstance(conn, id, d)
	case "running":
	case "started":
		return awsStateStartInstance(conn, id, d)
	default:
		return fmt.Errorf("State name %s is invalid.", state)
	}
	return nil
}

func resourceAwsInstanceStateDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

func awsStateStartInstance(conn *ec2.EC2, id string, d *schema.ResourceData) error {
	log.Printf("[INFO] changing instance state from %s to running", d.Get("State").(string))

	startOutput, err := conn.StartInstances(&ec2.StartInstancesInput{
		InstanceIds: aws.StringSlice([]string{id}),
	},
	)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Waiting for instance (%s) to become running", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending", "running", "shutting-down", "stopped", "stopping"},
		Target:     []string{"running"},
		Refresh:    InstanceStateRefreshFunc(conn, id, ""),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for instance (%s) to stopped: %s", id, err)
	}

	if len(startOutput.StartingInstances) > 0 {
		d.Set("State", startOutput.StartingInstances[0].CurrentState.Name)
	}

	return nil
}

func awsStateStopInstance(conn *ec2.EC2, id string, d *schema.ResourceData) error {
	log.Printf("[INFO] changing instance state from %s to stopped", d.Get("State").(string))

	stopOutput, err := conn.StopInstances(&ec2.StopInstancesInput{
		InstanceIds: aws.StringSlice([]string{id}),
	},
	)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Waiting for instance (%s) to become stopped", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending", "running", "shutting-down", "stopped", "stopping"},
		Target:     []string{"stopped"},
		Refresh:    InstanceStateRefreshFunc(conn, id, ""),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for instance (%s) to stopped: %s", id, err)
	}

	if len(stopOutput.StoppingInstances) > 0 {
		d.Set("State", stopOutput.StoppingInstances[0].CurrentState.Name)
	}

	return nil
}

func awsInStateReadInstance(conn *ec2.EC2, id string) (string, error) {
	descOutput, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{id}),
	},
	)

	if err != nil {
		return "", err
	}

	instance := descOutput.Reservations[0].Instances[0]
	log.Printf("[DEBUG] current state is %s", *instance.State.Name)

	return *instance.State.Name, nil
}
