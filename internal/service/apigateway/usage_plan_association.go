package apigateway

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceUsagePlanAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceUsagePlanAssociationCreate,
		Read:   resourceUsagePlanAssociationRead,
		Delete: resourceUsagePlanAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"usage_plan_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"api_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},

			"stage": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceUsagePlanAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	operations := make([]*apigateway.PatchOperation, 0)

	api_stage := fmt.Sprintf("%s:%s", d.Get("api_id"), d.Get("stage"))
	operations = append(operations, &apigateway.PatchOperation{
		Op:    aws.String(apigateway.OpAdd),
		Path:  aws.String("/apiStages"),
		Value: aws.String(api_stage),
	})
	input := &apigateway.UpdateUsagePlanInput{
		UsagePlanId:     aws.String(fmt.Sprintf("%s", d.Get("usage_plan_id"))),
		PatchOperations: operations,
	}
	log.Printf("[DEBUG] UsagePlanAssociation creation input: %#v", input)

	var up *apigateway.UsagePlan
	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		up, err = conn.UpdateUsagePlan(input)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("Error creating UsagePlanAssociation: %s for plan %s", err, d.Get("usage_plan_id"))
	}

	log.Printf("[DEBUG] UsagePlan creation response: %s", up)

	d.SetId(fmt.Sprintf("%s/%s/%s", d.Get("usage_plan_id"), d.Get("api_id"), d.Get("stage")))

	return resourceUsagePlanAssociationRead(d, meta)
}

func resourceUsagePlanAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	spl := strings.Split(d.Id(), "/")
	input := &apigateway.GetUsagePlanInput{
		UsagePlanId: aws.String(spl[0]),
	}

	output, err := conn.GetUsagePlan(input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway UsagePlanAssociation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	for _, stage := range output.ApiStages {
		if aws.StringValue(stage.ApiId) == spl[1] && aws.StringValue(stage.Stage) == spl[2] {
			d.Set("usage_plan_id", spl[0])
			d.Set("api_id", aws.StringValue(stage.ApiId))
			d.Set("stage", aws.StringValue(stage.Stage))
			return nil
		}
	}

	return err
}

func resourceUsagePlanAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	input := &apigateway.GetUsagePlanInput{
		UsagePlanId: aws.String(fmt.Sprintf("%s", d.Get("usage_plan_id"))),
	}

	output, err := conn.GetUsagePlan(input)
	if tfresource.NotFound(err) {
		return nil
	}

	found := false
	for _, stage := range output.ApiStages {
		if aws.StringValue(stage.ApiId) == d.Get("api_id").(string) && aws.StringValue(stage.Stage) == d.Get("stage").(string) {
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	operations := make([]*apigateway.PatchOperation, 0)

	api_stage := fmt.Sprintf("%s:%s", d.Get("api_id"), d.Get("stage"))
	operations = append(operations, &apigateway.PatchOperation{
		Op:    aws.String(apigateway.OpRemove),
		Path:  aws.String("/apiStages"),
		Value: aws.String(api_stage),
	})

	req := &apigateway.UpdateUsagePlanInput{
		UsagePlanId:     aws.String(fmt.Sprintf("%s", d.Get("usage_plan_id"))),
		PatchOperations: operations,
	}

	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error
		_, err = conn.UpdateUsagePlan(req)

		if tfresource.NotFound(err) {
			return nil
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("Error deleting UsagePlan Stage: %s for plan %s", err, d.Get("usage_plan_id"))
	}

	return nil
}
