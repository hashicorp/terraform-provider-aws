package aws

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsServiceCatalogLaunchRoleConstraint() *schema.Resource {
	var awsResourceIdPattern = regexp.MustCompile("^[a-zA-Z0-9_\\-]*")
	return &schema.Resource{
		Create: resourceAwsServiceCatalogLaunchRoleConstraintCreate,
		Read:   resourceAwsServiceCatalogLaunchRoleConstraintRead,
		Update: resourceAwsServiceCatalogLaunchRoleConstraintUpdate,
		Delete: resourceAwsServiceCatalogLaunchRoleConstraintDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			// one of local_role_name or role_arn but not both
			"local_role_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"role_arn"},
			},
			"role_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"local_role_name"},
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					awsResourceIdPattern,
					"invalid id format"),
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					awsResourceIdPattern,
					"invalid id format"),
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parameters": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsServiceCatalogLaunchRoleConstraintCreate(d *schema.ResourceData, meta interface{}) error {
	jsonDoc, errJson := resourceAwsServiceCatalogLaunchRoleConstraintJsonParameters(d)
	if errJson != nil {
		return errJson
	}
	errCreate := resourceAwsServiceCatalogConstraintCreateFromJson(d, meta, jsonDoc, "LAUNCH")
	if errCreate != nil {
		return errCreate
	}
	return resourceAwsServiceCatalogLaunchRoleConstraintRead(d, meta)
}

func resourceAwsServiceCatalogLaunchRoleConstraintJsonParameters(d *schema.ResourceData) (string, error) {
	if localRoleName, ok := d.GetOk("local_role_name"); ok && localRoleName != "" {
		type LocalRoleNameLaunchParameters struct {
			LocalRoleName string
		}
		var launchParameters LocalRoleNameLaunchParameters
		launchParameters.LocalRoleName = localRoleName.(string)
		marshal, err := json.Marshal(&launchParameters)
		return string(marshal), err
	} else if roleArn, ok := d.GetOk("role_arn"); ok && roleArn != "" {
		type LocalRoleNameLaunchParameters struct {
			RoleArn string
		}
		var launchParameters LocalRoleNameLaunchParameters
		launchParameters.RoleArn = roleArn.(string)
		marshal, err := json.Marshal(&launchParameters)
		return string(marshal), err
	}
	return "", fmt.Errorf("either local_role_name or role_arn must be specified")
}

func resourceAwsServiceCatalogLaunchRoleConstraintRead(d *schema.ResourceData, meta interface{}) error {
	constraint, err := resourceAwsServiceCatalogConstraintReadBase(d, meta)
	if err != nil {
		return err
	}
	if constraint == nil {
		return nil
	}
	var jsonDoc *string = constraint.ConstraintParameters
	var bytes []byte = []byte(*jsonDoc)
	type LaunchParameters struct {
		LocalRoleName string
		RoleArn       string
	}
	var launchParameters LaunchParameters
	err = json.Unmarshal(bytes, &launchParameters)
	if err != nil {
		return err
	}
	if launchParameters.LocalRoleName != "" {
		d.Set("local_role_name", launchParameters.LocalRoleName)
		d.Set("role_arn", nil)
	} else if launchParameters.RoleArn != "" {
		d.Set("role_arn", launchParameters.RoleArn)
		d.Set("local_role_name", nil)
	}
	return nil
}

func resourceAwsServiceCatalogLaunchRoleConstraintUpdate(d *schema.ResourceData, meta interface{}) error {
	input := servicecatalog.UpdateConstraintInput{}
	if d.HasChanges("launch_role_arn", "role_arn") {
		parameters, err := resourceAwsServiceCatalogLaunchRoleConstraintJsonParameters(d)
		if err != nil {
			return err
		}
		input.Parameters = aws.String(parameters)
	}
	err := resourceAwsServiceCatalogConstraintUpdateBase(d, meta, input)
	if err != nil {
		return err
	}
	return resourceAwsServiceCatalogLaunchRoleConstraintRead(d, meta)
}

func resourceAwsServiceCatalogLaunchRoleConstraintDelete(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsServiceCatalogConstraintDelete(d, meta)
}
