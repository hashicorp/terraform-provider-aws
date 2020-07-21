package aws

import (
	"encoding/json"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsServiceCatalogLaunchNotificationConstraint() *schema.Resource {
	var awsResourceIdPattern = regexp.MustCompile("^[a-zA-Z0-9_\\-]*")
	return &schema.Resource{
		Create: resourceAwsServiceCatalogLaunchNotificationConstraintCreate,
		Read:   resourceAwsServiceCatalogLaunchNotificationConstraintRead,
		Update: resourceAwsServiceCatalogLaunchNotificationConstraintUpdate,
		Delete: resourceAwsServiceCatalogLaunchNotificationConstraintDelete,
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
			"notification_arns": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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

func resourceAwsServiceCatalogLaunchNotificationConstraintCreate(d *schema.ResourceData, meta interface{}) error {
	jsonDoc, errJson := resourceAwsServiceCatalogLaunchNotificationConstraintJsonParameters(d)
	if errJson != nil {
		return errJson
	}
	errCreate := resourceAwsServiceCatalogConstraintCreateFromJson(d, meta, jsonDoc, "NOTIFICATION")
	if errCreate != nil {
		return errCreate
	}
	return resourceAwsServiceCatalogLaunchNotificationConstraintRead(d, meta)
}

type awsServiceCatalogLaunchNotificationConstraint struct {
	Description      string
	PortfolioId      string
	ProductId        string
	NotificationArns []string
}

func resourceAwsServiceCatalogLaunchNotificationConstraintJsonParameters(d *schema.ResourceData) (string, error) {
	constraint := awsServiceCatalogLaunchNotificationConstraint{}
	if description, ok := d.GetOk("description"); ok && description != "" {
		constraint.Description = description.(string)
	}
	constraint.PortfolioId = d.Get("portfolio_id").(string)
	constraint.ProductId = d.Get("product_id").(string)
	if err := resourceAwsServiceCatalogLaunchNotificationConstraintParseArns(d, &constraint); err != nil {
		return "", err
	}
	marshal, err := json.Marshal(constraint)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

func resourceAwsServiceCatalogLaunchNotificationConstraintParseArns(d *schema.ResourceData, constraint *awsServiceCatalogLaunchNotificationConstraint) error {
	constraint.NotificationArns = []string{}
	if arns, ok := d.GetOk("notification_arns"); ok {
		for _, arn := range arns.([]interface{}) {
			constraint.NotificationArns = append(constraint.NotificationArns, arn.(string))
		}
	}
	return nil
}

func resourceAwsServiceCatalogLaunchNotificationConstraintRead(d *schema.ResourceData, meta interface{}) error {
	constraint, err := resourceAwsServiceCatalogConstraintReadBase(d, meta)
	if err != nil {
		return err
	}
	if constraint == nil {
		return nil
	}
	arns, err := flattenAwsServiceCatalogLaunchNotificationConstraintArnsJson(constraint.ConstraintParameters)
	if err != nil {
		return err
	}
	if err := d.Set("notification_arns", arns); err != nil {
		return err
	}
	return nil
}

func flattenAwsServiceCatalogLaunchNotificationConstraintArnsJson(constraintParameters *string) ([]interface{}, error) {
	// ConstraintParameters is returned from AWS as a JSON string
	var parameters awsServiceCatalogLaunchNotificationConstraint
	err := json.Unmarshal([]byte(*constraintParameters), &parameters)
	if err != nil {
		return nil, err
	}
	return flattenAwsServiceCatalogLaunchNotificationConstraintArns(parameters.NotificationArns), nil
}

func flattenAwsServiceCatalogLaunchNotificationConstraintArns(arns []string) []interface{} {
	stringList := make([]*string, 0)
	for _, arn := range arns {
		var copyOfArn = arn
		stringList = append(stringList, &copyOfArn)
	}
	return flattenStringList(stringList)
}

func resourceAwsServiceCatalogLaunchNotificationConstraintUpdate(d *schema.ResourceData, meta interface{}) error {
	input := servicecatalog.UpdateConstraintInput{}
	if d.HasChanges("notification_arns") {
		parameters, err := resourceAwsServiceCatalogLaunchNotificationConstraintJsonParameters(d)
		if err != nil {
			return err
		}
		input.Parameters = aws.String(parameters)
	}
	err := resourceAwsServiceCatalogConstraintUpdateBase(d, meta, input)
	if err != nil {
		return err
	}
	return resourceAwsServiceCatalogLaunchNotificationConstraintRead(d, meta)
}

func resourceAwsServiceCatalogLaunchNotificationConstraintDelete(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsServiceCatalogConstraintDelete(d, meta)
}
