package lambda

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceAlias() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliasCreate,
		Read:   resourceAliasRead,
		Update: resourceAliasUpdate,
		Delete: resourceAliasDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAliasImport,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"function_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Using function name or ARN should not be shown as a diff.
					// Try to convert the old and new values from ARN to function name
					oldFunctionName, oldFunctionNameErr := GetFunctionNameFromARN(old)
					newFunctionName, newFunctionNameErr := GetFunctionNameFromARN(new)
					return (oldFunctionName == new && oldFunctionNameErr == nil) || (newFunctionName == old && newFunctionNameErr == nil)
				},
			},
			"function_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invoke_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"routing_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_version_weights": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeFloat},
						},
					},
				},
			},
		},
	}
}

// resourceAliasCreate maps to:
// CreateAlias in the API / SDK
func resourceAliasCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	functionName := d.Get("function_name").(string)
	aliasName := d.Get("name").(string)

	log.Printf("[DEBUG] Creating Lambda alias: alias %s for function %s", aliasName, functionName)

	params := &lambda.CreateAliasInput{
		Description:     aws.String(d.Get("description").(string)),
		FunctionName:    aws.String(functionName),
		FunctionVersion: aws.String(d.Get("function_version").(string)),
		Name:            aws.String(aliasName),
		RoutingConfig:   expandAliasRoutingConfiguration(d.Get("routing_config").([]interface{})),
	}

	aliasConfiguration, err := conn.CreateAlias(params)
	if err != nil {
		return fmt.Errorf("Error creating Lambda alias: %s", err)
	}

	d.SetId(aws.StringValue(aliasConfiguration.AliasArn))

	return resourceAliasRead(d, meta)
}

// resourceAliasRead maps to:
// GetAlias in the API / SDK
func resourceAliasRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	log.Printf("[DEBUG] Fetching Lambda alias: %s:%s", d.Get("function_name"), d.Get("name"))

	params := &lambda.GetAliasInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
		Name:         aws.String(d.Get("name").(string)),
	}

	aliasConfiguration, err := conn.GetAlias(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" && strings.Contains(awsErr.Message(), "Cannot find alias arn") {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("description", aliasConfiguration.Description)
	d.Set("function_version", aliasConfiguration.FunctionVersion)
	d.Set("name", aliasConfiguration.Name)
	d.Set("arn", aliasConfiguration.AliasArn)
	d.SetId(aws.StringValue(aliasConfiguration.AliasArn))

	invokeArn := functionInvokeARN(*aliasConfiguration.AliasArn, meta)
	d.Set("invoke_arn", invokeArn)

	if err := d.Set("routing_config", flattenAliasRoutingConfiguration(aliasConfiguration.RoutingConfig)); err != nil {
		return fmt.Errorf("error setting routing_config: %s", err)
	}

	return nil
}

// resourceAliasDelete maps to:
// DeleteAlias in the API / SDK
func resourceAliasDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	log.Printf("[INFO] Deleting Lambda alias: %s:%s", d.Get("function_name"), d.Get("name"))

	params := &lambda.DeleteAliasInput{
		FunctionName: aws.String(d.Get("function_name").(string)),
		Name:         aws.String(d.Get("name").(string)),
	}

	_, err := conn.DeleteAlias(params)
	if err != nil {
		return fmt.Errorf("Error deleting Lambda alias: %s", err)
	}

	return nil
}

// resourceAliasUpdate maps to:
// UpdateAlias in the API / SDK
func resourceAliasUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LambdaConn

	log.Printf("[DEBUG] Updating Lambda alias: %s:%s", d.Get("function_name"), d.Get("name"))

	params := &lambda.UpdateAliasInput{
		Description:     aws.String(d.Get("description").(string)),
		FunctionName:    aws.String(d.Get("function_name").(string)),
		FunctionVersion: aws.String(d.Get("function_version").(string)),
		Name:            aws.String(d.Get("name").(string)),
		RoutingConfig:   expandAliasRoutingConfiguration(d.Get("routing_config").([]interface{})),
	}

	_, err := conn.UpdateAlias(params)
	if err != nil {
		return fmt.Errorf("Error updating Lambda alias: %s", err)
	}

	return nil
}

func expandAliasRoutingConfiguration(l []interface{}) *lambda.AliasRoutingConfiguration {
	aliasRoutingConfiguration := &lambda.AliasRoutingConfiguration{}

	if len(l) == 0 || l[0] == nil {
		return aliasRoutingConfiguration
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["additional_version_weights"]; ok {
		aliasRoutingConfiguration.AdditionalVersionWeights = expandFloat64Map(v.(map[string]interface{}))
	}

	return aliasRoutingConfiguration
}

func resourceAliasImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("Unexpected format of ID (%q), expected FUNCTION_NAME/ALIAS", d.Id())
	}

	functionName := idParts[0]
	alias := idParts[1]

	d.Set("function_name", functionName)
	d.Set("name", alias)
	return []*schema.ResourceData{d}, nil
}
