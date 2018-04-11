package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsIamServiceLinkedRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIamServiceLinkedRoleCreate,
		Read:   resourceAwsIamServiceLinkedRoleRead,
		Delete: resourceAwsIamServiceLinkedRoleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"aws_service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value := v.(string)
					if !strings.HasSuffix(value, ".amazonaws.com") {
						es = append(es, fmt.Errorf(
							"%q must be a service URL e.g. elasticbeanstalk.amazonaws.com", k))
					}
					return
				},
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"create_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"unique_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsIamServiceLinkedRoleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	serviceName := d.Get("aws_service_name").(string)

	params := &iam.CreateServiceLinkedRoleInput{
		AWSServiceName: aws.String(serviceName),
	}

	resp, err := conn.CreateServiceLinkedRole(params)

	if err != nil {
		return fmt.Errorf("Error creating service-linked role with name %s: %s", serviceName, err)
	}
	d.SetId(*resp.Role.Arn)

	return resourceAwsIamServiceLinkedRoleRead(d, meta)
}

func resourceAwsIamServiceLinkedRoleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	arnSplit := strings.Split(d.Id(), "/")
	roleName := arnSplit[len(arnSplit)-1]
	serviceName := arnSplit[len(arnSplit)-2]

	params := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}

	resp, err := conn.GetRole(params)

	if err != nil {
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			log.Printf("[WARN] IAM service linked role %s not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	role := resp.Role

	d.Set("name", role.RoleName)
	d.Set("path", role.Path)
	d.Set("arn", role.Arn)
	d.Set("create_date", role.CreateDate)
	d.Set("unique_id", role.RoleId)
	d.Set("description", role.Description)
	d.Set("aws_service_name", serviceName)

	return nil
}

func resourceAwsIamServiceLinkedRoleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iamconn

	arnSplit := strings.Split(d.Id(), "/")
	roleName := arnSplit[len(arnSplit)-1]

	params := &iam.DeleteServiceLinkedRoleInput{
		RoleName: aws.String(roleName),
	}

	resp, err := conn.DeleteServiceLinkedRole(params)

	if err != nil {
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting service-linked role %s: %s", d.Id(), err)
	}

	deletionTaskId := aws.StringValue(resp.DeletionTaskId)

	stateConf := &resource.StateChangeConf{
		Pending: []string{iam.DeletionTaskStatusTypeInProgress, iam.DeletionTaskStatusTypeNotStarted},
		Target:  []string{iam.DeletionTaskStatusTypeSucceeded},
		Refresh: deletionRefreshFunc(conn, deletionTaskId),
		Timeout: 5 * time.Minute,
		Delay:   10 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return nil
		}
		return fmt.Errorf("Error waiting for role (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func deletionRefreshFunc(conn *iam.IAM, deletionTaskId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		params := &iam.GetServiceLinkedRoleDeletionStatusInput{
			DeletionTaskId: aws.String(deletionTaskId),
		}

		resp, err := conn.GetServiceLinkedRoleDeletionStatus(params)
		if err != nil {
			return nil, "", err
		}

		return resp, aws.StringValue(resp.Status), nil
	}
}
