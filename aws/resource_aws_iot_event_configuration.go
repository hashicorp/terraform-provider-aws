package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	// Serves as Id for terraform
	Name = "name"
	// Name and Enable map values for event
	ConfigurationsMap = "configurations_map"
	// Describe Event Ouput
	DescribeEventOutput = "describe_event_output"
)

/*
	This structure defines the data schema and CRUD operations for the resource.
	Terraform itself handles which function to call and with what data.
	Based on the schema and current state of the resource from terraform.state.
*/
func resourceAwsIoTEventConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceEventConfigurationCreate,
		Read:   resourceEventConfigurationRead,
		Update: resourceEventConfigurationUpdate,
		Delete: resourceEventConfigurationDelete,
		Schema: map[string]*schema.Schema{
			Name: {
				Type:     schema.TypeString,
				Required: true,
			},
			DescribeEventOutput: {
				Type:     schema.TypeMap,
				Optional: true,
			},
			ConfigurationsMap: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}
}

func resourceEventConfigurationCreate(d *schema.ResourceData, m interface{}) error {
	conn := m.(*iot.IoT)

	eventConfigurationsMap := make(map[string]*iot.Configuration)
	configurationsMapActions := d.Get(ConfigurationsMap).(*schema.Set).List()
	var eventConfigurationInput *iot.UpdateEventConfigurationsInput

	log.Printf("[DEBUG] Create Event Configuration: %v", ConfigurationsMap)

	for _, a := range configurationsMapActions {
		raw := a.(map[string]interface{})
		eventConfigurationsMap[raw["attribute_name"].(string)] = &iot.Configuration{Enabled: aws.Bool(raw["enabled"].(bool))}
		eventConfigurationInput = &iot.UpdateEventConfigurationsInput{
			EventConfigurations: eventConfigurationsMap,
		}
	}

	_, err := conn.UpdateEventConfigurations(eventConfigurationInput)

	if err != nil {
		log.Printf("[DEBUG] error in event configuration: %v", err)
		return err
	}

	d.SetId(d.Get(Name).(string))
	log.Printf("Event Configuration Created")
	return resourceEventConfigurationRead(d, m)
}

// Fetches all the event configurations, the output contains all the available configurations.
func resourceEventConfigurationRead(d *schema.ResourceData, m interface{}) error {
	conn := m.(*iot.IoT)

	out, err := conn.DescribeEventConfigurations(&iot.DescribeEventConfigurationsInput{})

	if err != nil {
		log.Printf("[DEBUG] error in describe event configuration: %v", err)
		return err
	}

	d.Set(DescribeEventOutput, out.EventConfigurations)
	return nil
}

// Updates the event configuration, when there is any update in the current state.
func resourceEventConfigurationUpdate(d *schema.ResourceData, m interface{}) error {
	conn := m.(*iot.IoT)

	eventConfigurationsMap := make(map[string]*iot.Configuration)
	configurationsMapActions := d.Get(ConfigurationsMap).(*schema.Set).List()
	var eventConfigurationInput *iot.UpdateEventConfigurationsInput

	log.Printf("[DEBUG] Update Event Configuration: %v", ConfigurationsMap)

	for _, a := range configurationsMapActions {
		raw := a.(map[string]interface{})
		eventConfigurationsMap[raw["attribute_name"].(string)] = &iot.Configuration{Enabled: aws.Bool(raw["enabled"].(bool))}
		eventConfigurationInput = &iot.UpdateEventConfigurationsInput{
			EventConfigurations: eventConfigurationsMap,
		}
	}

	_, err := conn.UpdateEventConfigurations(eventConfigurationInput)

	if err != nil {
		log.Printf("[DEBUG] error in update event configuration: %v", err)
		return err
	}

	log.Printf("Event Configuration Updated")
	return resourceEventConfigurationRead(d, m)
}

// Any non-error return value terraform assumes the resource was deleted successfully.
func resourceEventConfigurationDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
