package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotevents"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func generateActionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"clear_timer": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			"firehose": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"delivery_stream_name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"separator": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
			"iot_events": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			"iot_topic_publish": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"mqtt_topic": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			"lambda": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"function_arn": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			"reset_timer": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			"set_timer": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Required: true,
					},
					"seconds": {
						Type:         schema.TypeInt,
						Required:     true,
						ValidateFunc: validation.IntAtLeast(1),
					},
				},
			},
			"set_variable": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"name": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringLenBetween(1, 128),
					},
					"value": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			"sns": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"target_arn": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			"sqs": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"queue_url": {
						Type:     schema.TypeString,
						Required: true,
					},
					"use_base64": {
						Type:     schema.TypeBool,
						Optional: true,
					},
				},
			},
		},
	}
}

func generateEventSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateActionSchema(),
			},
			"condition": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func generateTransitionEventSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"action": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateActionSchema(),
			},
			"condition": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"next_state": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateStateSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"on_enter": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"event": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem:     generateEventSchema(),
					},
				},
			},
			"on_exit": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"event": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem:     generateEventSchema(),
					},
				},
			},
			"on_input": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: map[string]*schema.Schema{
					"event": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem:     generateEventSchema(),
					},
					"transition_event": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem:     generateTransitionEventSchema(),
					},
				},
			},
		},
	}
}

func resourceAwsIotDetectorModel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotDetectorCreate,
		Read:   resourceAwsIotDetectorRead,
		Update: resourceAwsIotDetectorUpdate,
		Delete: resourceAwsIotDetectorDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"definition": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"initial_state_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"states": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Required: true,
							Elem:     generateStateSchema(),
						},
					},
				}},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func parseAction(rawAction map[string]interface{}) *iotevents.ActionData {
	action := &iotevents.ActionData{}

	if v, ok := rawAction["clear_timer"].(*schema.Set); ok && len(v.List()) != 0 {

		rawClearTimer := v.List()[0].(map[string]interface{})
		action.ClearTimer = &iotevents.ClearTimerAction{
			TimerName: aws.String(rawClearTimer["name"].(string)),
		}
	}

	if v, ok := rawAction["firehose"].(*schema.Set); ok && len(v.List()) != 0 {

		rawFirehose := v.List()[0].(map[string]interface{})
		action.Firehose = &iotevents.FirehoseAction{
			DeliveryStreamName: aws.String(rawFirehose["delivery_stream_name"].(string)),
		}

		if fv, ok := rawFirehose["separator"].(string); ok && fv != "" {
			action.Firehose.Separator = aws.String(fv)
		}
	}

	if v, ok := rawAction["iot_events"].(*schema.Set); ok && len(v.List()) != 0 {

		rawIotEvents := v.List()[0].(map[string]interface{})
		action.IotEvents = &iotevents.Action{
			InputName: aws.String(rawIotEvents["name"].(string)),
		}
	}

	if v, ok := rawAction["iot_topic_publish"].(*schema.Set); ok && len(v.List()) != 0 {

		rawIotTopicPublish := v.List()[0].(map[string]interface{})
		action.IotTopicPublish = &iotevents.IotTopicPublishAction{
			MqttTopic: aws.String(rawIotTopicPublish["mqtt_topic"].(string)),
		}
	}

	if v, ok := rawAction["lambda"].(*schema.Set); ok && len(v.List()) != 0 {

		rawLambda := v.List()[0].(map[string]interface{})
		action.Lambda = &iotevents.LambdaAction{
			FunctionArn: aws.String(rawLambda["function_arn"].(string)),
		}
	}

	if v, ok := rawAction["reset_timer"].(*schema.Set); ok && len(v.List()) != 0 {

		rawResetTimer := v.List()[0].(map[string]interface{})
		action.ResetTimer = &iotevents.ResetTimerAction{
			TimerName: aws.String(rawResetTimer["name"].(string)),
		}
	}

	if v, ok := rawAction["set_timer"].(*schema.Set); ok && len(v.List()) != 0 {

		rawSetTimer := v.List()[0].(map[string]interface{})
		action.SetTimer = &iotevents.SetTimerAction{
			TimerName: aws.String(rawSetTimer["name"].(string)),
			Seconds:   aws.Int64(rawSetTimer["seconds"].(int64)),
		}
	}

	if v, ok := rawAction["set_variable"].(*schema.Set); ok && len(v.List()) != 0 {

		rawSetVariable := v.List()[0].(map[string]interface{})
		action.SetVariable = &iotevents.SetVariableAction{
			VariableName: aws.String(rawSetVariable["name"].(string)),
			Value:        aws.String(rawSetVariable["value"].(string)),
		}
	}

	if v, ok := rawAction["sns"].(*schema.Set); ok && len(v.List()) != 0 {

		rawSns := v.List()[0].(map[string]interface{})
		action.Sns = &iotevents.SNSTopicPublishAction{
			TargetArn: aws.String(rawSns["target_arn"].(string)),
		}
	}

	if v, ok := rawAction["sqs"].(*schema.Set); ok && len(v.List()) != 0 {

		rawSqs := v.List()[0].(map[string]interface{})
		action.Sqs = &iotevents.SqsAction{
			QueueUrl: aws.String(rawSqs["queue_url"].(string)),
		}

		if sqsv, ok := rawSqs["use_base64"].(bool); ok {
			action.Sqs.UseBase64 = aws.Bool(sqsv)
		}
	}

	return action
}

func parseEvent(rawEvent map[string]interface{}) *iotevents.Event {
	event := &iotevents.Event{
		EventName: aws.String(rawEvent["name"].(string)),
	}

	actions := make([]*iotevents.ActionData, 0)
	for i, rawAct := range rawEvent["action"].(*schema.Set).List() {
		actions = append(actions, parseAction(rawAct.(map[string]interface{})))
	}

	event.Actions = actions

	if v, ok := rawEvent["condition"].(string); ok && v != "" {
		event.Condition = aws.String(v)
	}

	return event
}

func parseTransitionEvent(rawTransitionEvent map[string]interface{}) *iotevents.TransitionEvent {
	event := &iotevents.TransitionEvent{
		EventName: aws.String(rawTransitionEvent["name"].(string)),
		NextState: aws.String(rawTransitionEvent["next_state"].(string)),
	}

	actions := make([]*iotevents.ActionData, 0)
	for i, rawAct := range rawTransitionEvent["action"].(*schema.Set).List() {
		actions = append(actions, parseAction(rawAct.(map[string]interface{})))
	}

	event.Actions = actions

	if v, ok := rawTransitionEvent["condition"].(string); ok && v != "" {
		event.Condition = aws.String(v)
	}

	return event
}

func parseEventsList(rawEvents []interface{}) []*iotevents.Event {
	events := make([]*iotevents.Event, 0)

	for i, rawEvent := range rawEvents {
		events = append(events, parseEvent(rawEvent.(map[string]interface{})))
	}

	return events
}

func parseOnEnter(rawOnEnter map[string]interface{}) *iotevents.OnEnterLifecycle {
	events := parseEventsList(rawOnEnter["event"].(*schema.Set).List())

	if len(events) != 0 {
		onEnter := &iotevents.OnEnterLifecycle{
			Events: events,
		}
		return onEnter
	}

	return nil
}

func parseOnExit(rawOnExit map[string]interface{}) *iotevents.OnExitLifecycle {
	events := parseEventsList(rawOnExit["event"].(*schema.Set).List())

	if len(events) != 0 {
		onExit := &iotevents.OnExitLifecycle{
			Events: events,
		}
		return onExit
	}
	return nil
}

func parseOnInput(rawOnInput map[string]interface{}) *iotevents.OnInputLifecycle {
	var onInput *iotevents.OnInputLifecycle

	events := parseEventsList(rawOnInput["event"].(*schema.Set).List())
	if len(events) != 0 {
		onInput := &iotevents.OnInputLifecycle{
			Events: events,
		}
	}

	transitionEvents := make([]*iotevents.TransitionEvent, 0)
	for i, rawEvent := range rawOnInput["transition_event"].(*schema.Set).List() {
		transitionEvents = append(transitionEvents, parseTransitionEvent(rawEvent.(map[string]interface{})))
	}
	if len(transitionEvents) != 0 {
		if onInput == nil {
			onInput = &iotevents.OnInputLifecycle{
				TransitionEvents: transitionEvents,
			}
		} else {
			onInput.TransitionEvents = transitionEvents
		}
	}

	if onInput != nil {
		return onInput
	}

	return nil
}

// TODO: separate this function on different function that parse on... structures:
// parseOnEnter, parseOnExit, parseOnInput
func parseState(rawState map[string]interface{}) *iotevents.State {
	state := &iotevents.State{
		StateName: aws.String(rawState["name"].(string)),
	}

	if v, ok := rawState["on_enter"].(*schema.Set); ok && len(v.List()) != 0 {
		rawOnEnter := v.List()[0].(map[string]interface{})
		state.OnEnter = parseOnEnter(rawOnEnter)
	}

	if v, ok := rawState["on_exit"].(*schema.Set); ok && len(v.List()) != 0 {
		rawOnExit := v.List()[0].(map[string]interface{})
		state.OnExit = parseOnExit(rawOnExit)
	}

	if v, ok := rawState["on_input"].(*schema.Set); ok && len(v.List()) != 0 {
		rawOnInput := v.List()[0].(map[string]interface{})
		state.OnInput = parseOnInput(rawOnInput)
	}

	return state
}

// func parseDetectorModelDefinition(rawDetectorModelDefinition map[string]interface{}) *iotevents.DetectorModelDefinition {

// }

func resourceAwsIotDetectorCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	detectorName := d.Get("name").(string)
	detectorDefinition := d.Get("definition").(map[string]interface{})

	// How to convert list of structures to appropriate format usign aws. package
	detectorDefinitionParams := &iotevents.DetectorModelDefinition{
		InitialStateName: aws.String(detectorDefinition["initial_state_name"].(string)),
		States:           expandStringList(detectorDefinition["states"].([]interface{})),
	}

	roleArn := d.Get("role_arn").(string)

	params := &iotevents.CreateDetectorModelInput{
		DetectorModelName:       aws.String(detectorName),
		DetectorModelDefinition: detectorDefinitionParams,
		RoleArn:                 aws.String(roleArn),
	}

	if v, ok := d.GetOk("description"); ok {
		params.DetectorModelDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("key"); ok {
		params.Key = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating IoT Model Detector: %s", params)
	_, err := conn.CreateDetectorModel(params)

	if err != nil {
		return err
	}

	d.SetId(detectorName)

	return resourceAwsIotDetectorRead(d, meta)
}

func resourceAwsIotDetectorRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	params := &iotevents.DescribeDetectorModelInput{
		DetectorModelName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Events Detector Model: %s", params)
	out, err := conn.DescribeDetectorModel(params)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsIotDetectorUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	detectorName := d.Get("name").(string)
	detectorDefinition := d.Get("definition").(map[string]interface{})

	detectorDefinitionParams := &iotevents.DetectorModelDefinition{
		InitialStateName: aws.String(detectorDefinition["initial_state_name"].(string)),
		States:           aws.String(detectorDefinition["initial_state_name"].([]string)),
	}
	roleArn := d.Get("role_arn").(string)

	params := &iotevents.UpdateDetectorModelInput{
		DetectorModelName:       aws.String(detectorName),
		DetectorModelDefinition: detectorDefinitionParams,
		RoleArn:                 aws.String(roleArn),
	}

	if v, ok := d.GetOk("description"); ok {
		params.DetectorModelDescription = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating IoT Events Detector Model: %s", params)
	_, err := conn.UpdateDetectorModel(params)

	if err != nil {
		return err
	}

	return resourceAwsIotDetectorRead(d, meta)
}

func resourceAwsIotDetectorDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	params := &iotevents.DeleteDetectorModelInput{
		DetectorModelName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting IoT Events Detector Model: %s", params)
	_, err := conn.DeleteDetectorModel(params)

	return err
}
