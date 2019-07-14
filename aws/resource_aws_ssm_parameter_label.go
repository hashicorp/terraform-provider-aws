package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSsmParameterLabel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmParameterLabelCreate,
		Read:   resourceAwsSsmParameterLabelRead,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"labels": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateSsmParameterLabel,
				},
			},
			"ssm_parameter_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ssm_parameter_version": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceAwsSsmParameterLabelCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssmconn
	req := &ssm.LabelParameterVersionInput{
		Labels:           expandStringList(d.Get("labels").([]interface{})),
		Name:             aws.String(d.Get("ssm_parameter_name").(string)),
		ParameterVersion: aws.Int64(d.Get("ssm_parameter_version").(int64)),
	}

	log.Printf("[DEBUG] Labeling Ssm Parameter version: %#v", req)
	output, err := conn.LabelParameterVersion(req)
	if err != nil {
		return fmt.Errorf("error labeling Ssm Parameter version: %s", err)
	}

	d.SetId(fmt.Sprintf("%s|%s", d.Get("ssm_parameter_name").(string), d.Get("ssm_parameter_version").(string)))

	if len(output.InvalidLabels) > 0 {
		log.Printf("[WARN] Some lavels were invalid: %v", output.InvalidLabels)
	}

	return nil

}

func resourceAwsSsmParameterLabelRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] Reading SSM Parameter label: %s", d.Id())

	result, err := ssmconn.GetParameterHistory(&ssm.GetParameterHistoryInput{
		Name:           aws.String(d.Get("ssm_parameter_name").(string)),
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("error getting SSM parameter history: %s", err)
	}

	if result == nil || len(result.Parameters) == 0 || result.Parameters[0] == nil {
		log.Printf("[WARN] SSM Parameter %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	labels := make([]string, 0)
	parameterVersionFound := false
	for _, parameter := range result.Parameters {
		if *parameter.Version == d.Get("ssm_parameter_version").(int64) {
			parameterVersionFound = true
			for _, label := range parameter.Labels {
				labels = append(labels, *label)
			}
		}
	}

	if !parameterVersionFound {
		log.Printf("[WARN] SSM Parameter %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("labels", labels)

	return nil
}
