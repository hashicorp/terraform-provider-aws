package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iotevents"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIoTEventsDetectorModel_basic(t *testing.T) {
	dName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_basic(dName),
				Check: resource.ComposeTestCheckFunc(
					testAccDetectorModelBasic("aws_iotevents_detector_model.detector"),
					resource.TestCheckResourceAttr("aws_iotevents_detector_model.detector", "name", fmt.Sprintf("test_detector_%s", rName)),
					resource.TestCheckResourceAttr("aws_iotevents_detector_model.detector", "description", "Example detector model"),
				),
			},
		},
	})
}

func TestAccAWSIoTEventsDetectorModel_full(t *testing.T) {
	dName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventsDetectorModelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventsDetectorModel_full(dName),
				Check: resource.ComposeTestCheckFunc(
					testAccDetectorModelFull("aws_iotevents_detector_model.detector"),
				),
			},
		},
	})
}

func testAccDetectorModelBasic() {
	conn := testAccProvider.Meta().(*AWSClient).ioteventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotevents_detector_model" {
			continue
		}

		params := &iotevents.DescribeDetectorModelInput{
			DetectorModelName: aws.string(rs.Primary.ID),
		}

		out, err := conn.DescribeDetectorModel(params)

		if err != nil {
			return err
		}

		detectorDefinition := out.DetectorModel.DetectorModelDefinition
		testInitialStateName := "first_state"
		if detectorDefinition.InitialStateName != testInitialStateName {
			return fmt.Errorf("Initial state name %s not equals to %s", detectorDefinition.InitialStateName, testInitialStateName)
		}

		states := detectorDefinition.States
		testStatesLen := 1
		if len(states) != testStatesLen {
			return fmt.Errorf("States len %d not equals to %d", len(states), testStatesLen)
		}

		// Test first state
		firstState := states[0]
		testFirstStateName := "first_state"
		if firstState.StateName != testFirstStateName {
			return fmt.Errorf("State name %s not equals to %s", firstState.StateName, testFirstStateName)
		}
		if firstState.OnEnter == nil {
			return fmt.Errorf("State OnEnter is equal Nil")
		}
		if firstState.OnExit != nil {
			return fmt.Errorf("State onExit is not equal Nil")
		}
		if firstState.OnInput != nil {
			return fmt.Errorf("State OnInput is not equal Nil")
		}

		// Test OnEnter
		onEnter := firstState.OnEnter
		events := onEnter.Events
		testEventsLen := 1
		if len(events) != testEventsLen {
			return fmt.Errorf("Events len %d not equals to %d", len(events), testEventsLen)
		}

		event := events[0]
		testEventName := "test_event_name"
		testEventCondition := "convert(Decimal, $input.test_input.temperature) > 20"
		if event.EventName != testEventName {
			return fmt.Errorf("Event name %s not equals to %s", event.EventName, testEventName)
		}
		if event.Condition != testEventCondition {
			return fmt.Errorf("Event condition %s not equals to %s", event.Condition, testEventCondition)
		}

		actions := event.Actions
		testActionsLen := 1
		if len(actions) != testActionsLen {
			return fmt.Errorf("Actions len %d not equals to %d", len(actions), testActionsLen)
		}

		action := actions[0]
		if action.ClearTimer == nil {
			return fmt.Errorf("ClearTimer is equal Nil")
		}

		clearTimer := action.ClearTimer
		testClearTimerName := "test_timer_name"
		if clearTimer.TimerName != testClearTimerName {
			return fmt.Errorf("Clear Timer name %s not equals to %s", clearTimer.TimerName, testClearTimerName)
		}
	}

	return nil
}

func checkAccFirstStateFull(state *iotevents.State) error {
	testFirstStateName := "first_state"
	if state.StateName != testFirstStateName {
		return fmt.Errorf("State name %s not equals to %s", state.StateName, testFirstStateName)
	}
	if state.OnEnter == nil {
		return fmt.Errorf("State OnEnter is equal Nil")
	}
	if state.OnExit == nil {
		return fmt.Errorf("State onExit is equal Nil")
	}
	if state.OnInput == nil {
		return fmt.Errorf("State OnInput is equal Nil")
	}

	// Check OnEnter
	onEnter := state.OnEnter
	events := onEnter.Events
	testEventsLen := 1
	if len(events) != testEventsLen {
		return fmt.Errorf("Events len %d not equals to %d", len(events), testEventsLen)
	}

	event := events[0]
	testEventName := "test_event_name"
	testEventCondition := "convert(Decimal, $input.test_input.temperature) > 20"
	if event.EventName != testEventName {
		return fmt.Errorf("Event name %s not equals to %s", event.EventName, testEventName)
	}
	if event.Condition != testEventCondition {
		return fmt.Errorf("Event condition %s not equals to %s", event.Condition, testEventCondition)
	}

	actions := event.Actions
	testActionsLen := 1
	if len(actions) != testActionsLen {
		return fmt.Errorf("Actions len %d not equals to %d", len(actions), testActionsLen)
	}

	action := actions[0]
	if action.ClearTimer == nil {
		return fmt.Errorf("ClearTimer is equal Nil")
	}

	clearTimer := action.ClearTimer
	testClearTimerName := "test_timer_name"
	if clearTimer.TimerName != testClearTimerName {
		return fmt.Errorf("Clear Timer name %s not equals to %s", clearTimer.TimerName, testClearTimerName)
	}

	// Test OnExit
	onExit := state.OnExit
	events = onExit.Events
	testEventsLen = 1
	if len(events) != testEventsLen {
		return fmt.Errorf("Events len %d not equals to %d", len(events), testEventsLen)
	}

	event = events[0]
	testEventName = "test_event_name"
	testEventCondition = "convert(Decimal, $input.test_input.temperature) > 20"
	if event.EventName != testEventName {
		return fmt.Errorf("Event name %s not equals to %s", event.EventName, testEventName)
	}
	if event.Condition != testEventCondition {
		return fmt.Errorf("Event condition %s not equals to %s", event.Condition, testEventCondition)
	}

	actions = event.Actions
	testActionsLen = 1
	if len(actions) != testActionsLen {
		return fmt.Errorf("Actions len %d not equals to %d", len(actions), testActionsLen)
	}

	action = actions[0]
	if action.ClearTimer == nil {
		return fmt.Errorf("ClearTimer is equal Nil")
	}

	clearTimer = action.ClearTimer
	testClearTimerName = "test_timer_name"
	if clearTimer.TimerName != testClearTimerName {
		return fmt.Errorf("Clear Timer name %s not equals to %s", clearTimer.TimerName, testClearTimerName)
	}

	// Test OnInput
	onInput = state.OnInput
	events = onInput.Events
	testEventsLen = 1
	if len(events) != testEventsLen {
		return fmt.Errorf("Events len %d not equals to %d", len(events), testEventsLen)
	}

	event = events[0]
	testEventName = "test_event_name"
	testEventCondition = "convert(Decimal, $input.test_input.temperature) > 20"
	if event.EventName != testEventName {
		return fmt.Errorf("Event name %s not equals to %s", event.EventName, testEventName)
	}
	if event.Condition != testEventCondition {
		return fmt.Errorf("Event condition %s not equals to %s", event.Condition, testEventCondition)
	}

	actions = event.Actions
	testActionsLen = 1
	if len(actions) != testActionsLen {
		return fmt.Errorf("Actions len %d not equals to %d", len(actions), testActionsLen)
	}

	action = actions[0]
	if action.ClearTimer == nil {
		return fmt.Errorf("ClearTimer is equal Nil")
	}

	clearTimer = action.ClearTimer
	testClearTimerName = "test_timer_name"
	if clearTimer.TimerName != testClearTimerName {
		return fmt.Errorf("Clear Timer name %s not equals to %s", clearTimer.TimerName, testClearTimerName)
	}

	transitionEvents := onInput.TransitionEvents
	testTransitionEventsLen = 1
	if len(transitionEvents) != testTransitionEventsLen {
		return fmt.Errorf("Transition Events len %d not equals to %d", len(transitionEvents), testTransitionEventsLen)
	}

	transitionEvent := transitionEvents[0]
	testTransitionEventName := "test_transition_event_name"
	testTransitionEventCondition := "convert(Decimal, $input.test_input.temperature) > 20"
	testTransitionEventNextState := "second_state"
	if transitionEvent.EventName != testTransitionEventName {
		return fmt.Errorf("Event name %s not equals to %s", transitionEvent.EventName, testTransitionEventName)
	}
	if transitionEvent.Condition != testTransitionEventCondition {
		return fmt.Errorf("Event condition %s not equals to %s", transitionEvent.Condition, testTransitionEventCondition)
	}
	if transitionEvent.NextState != testTransitionEventNextState {
		return fmt.Errorf("Next state %s not equals to %s", transitionEvent.NextState, testTransitionEventNextState)
	}

	return nil
}

func checkAccSecondStateFull(state *iotevents.State) {
	testStateName := "second_state"
	if state.StateName != testStateName {
		return fmt.Errorf("State name %s not equals to %s", state.StateName, testStateName)
	}
	if state.OnEnter == nil {
		return fmt.Errorf("State OnEnter is equal Nil")
	}
	if state.OnExit != nil {
		return fmt.Errorf("State onExit is not equal Nil")
	}
	if state.OnInput != nil {
		return fmt.Errorf("State OnInput is not equal Nil")
	}

	// Check OnEnter
	onEnter := state.OnEnter
	events := onEnter.Events
	testEventsLen := 1
	if len(events) != testEventsLen {
		return fmt.Errorf("Events len %d not equals to %d", len(events), testEventsLen)
	}

	event := events[0]
	testEventName := "test_event_name"
	testEventCondition := "convert(Decimal, $input.test_input.temperature) > 20"
	if event.EventName != testEventName {
		return fmt.Errorf("Event name %s not equals to %s", event.EventName, testEventName)
	}
	if event.Condition != testEventCondition {
		return fmt.Errorf("Event condition %s not equals to %s", event.Condition, testEventCondition)
	}

	actions := event.Actions
	testActionsLen := 1
	if len(actions) != testActionsLen {
		return fmt.Errorf("Actions len %d not equals to %d", len(actions), testActionsLen)
	}

	action := actions[0]
	if action.ClearTimer == nil {
		return fmt.Errorf("ClearTimer is equal Nil")
	}

	clearTimer := action.ClearTimer
	testClearTimerName := "test_timer_name"
	if clearTimer.TimerName != testClearTimerName {
		return fmt.Errorf("Clear Timer name %s not equals to %s", clearTimer.TimerName, testClearTimerName)
	}

	return nil

}

func testAccDetectorModelFull() {
	conn := testAccProvider.Meta().(*AWSClient).ioteventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotevents_detector_model" {
			continue
		}

		params := &iotevents.DescribeDetectorModelInput{
			DetectorModelName: aws.string(rs.Primary.ID),
		}

		out, err := conn.DescribeDetectorModel(params)

		if err != nil {
			return err
		}

		detectorDefinition := out.DetectorModel.DetectorModelDefinition
		testInitialStateName := "first_state"
		if detectorDefinition.InitialStateName != testInitialStateName {
			return fmt.Errorf("Initial state name %s not equals to %s", detectorDefinition.InitialStateName, testInitialStateName)
		}

		states := detectorDefinition.States
		testStatesLen := 2
		if len(states) != testStatesLen {
			return fmt.Errorf("States len %d not equals to %d", len(states), testStatesLen)
		}

		// Test first State
		firstState := states[0]
		err = checkAccFirstStateFull(firstState)
		if err != nil {
			return err
		}

		// Test second State
		secondState := states[1]
		err = checkAccSecondStateFull(secondState)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkAWSResponseError(responseErr error, checkErr error, responseErrNilValueMessage string) error {
	// responseErr: error part of response to AWS API.
	// checkErr: expected error
	// responseErrNilValueMessage: text of error that should be returned if responseErr equals nil.
	// return: nil if responseErr equals to checkErr
	//		   error if responseErr is nil or does not equal to checkErr

	if responseErr != nil {
		if isAWSErr(err, checkErr, "") {
			return nil
		} else {
			return responseErr
		}
	} else {
		return fmt.Errorf(responseErrNilValueMessage)
	}
}

func testAccCheckAWSIoTEventsDetectorModelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ioteventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotevents_detector_model" {
			continue
		}

		params := &iotevents.DescribeDetectorModelInput{
			DetectorModelName: aws.string(rs.Primary.ID),
		}

		_, err := conn.DescribeDetectorModel(params)

		errNilMessage := fmt.Sprintf("Expected IoTEvents Detector Model to be destroyed, %s found", rs.Primary.ID)
		return checkAWSResponseError(err, iotevents.ErrCodeResourceNotFoundException, errNilMessage)
	}

	return nil
}

const testAccAWSIoTEventsDetectorModelBasicConfig = `
resource "aws_iam_role" "iotevents_role" {
    name = "test_role_%[1]s"
    assume_role_policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement":[{
        "Effect": "Allow",
        "Principal": {
            "Service": "iotevents.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
    }]
}
EOF
}

resource "aws_iam_policy" "policy" {
    name = "test_policy_%[1]s"
    path = "/"
    description = "My test policy"
    policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Action": "*",
        "Resource": "*"
    }]
}
EOF
}

resource "aws_iam_policy_attachment" "attach_policy" {
    name = "test_policy_attachment_%[1]s"
    roles = ["${aws_iam_role.iotevents_role.name}"]
    policy_arn = "${aws_iam_policy.policy.arn}"
EOF
}

resource "aws_iotevents_input" "test_input" {
	name = "test_input"
  
	definition {
	  attribute {
		json_path = "temperature"
	  }
	}
}

`

func testAccAWSIoTEventsDetectorModel_basic(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  description = "Example detector model"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
        event {
          name      = "test_event_name"
          condition = "convert(Decimal, $input.test_input.temperature) > 20"

          action {
            clear_timer {
              name = "test_timer_name"
            }
          }
        }
      }
    }
  }

}
`, dName)
}

func testAccAWSIoTEventsDetectorModel_full(dName string) string {
	return fmt.Sprintf(testAccAWSIoTEventsDetectorModelBasicConfig+`
resource "aws_iotevents_detector_model" "detector" {
  name = "test_detector_%[1]s"
  role_arn = "${aws_iam_role.iotevents_role.arn}"

  definition {
    initial_state_name = "first_state"

    state {
      name = "first_state"
      on_enter {
		event {
			name      = "test_event_name"
			condition = "convert(Decimal, $input.test_input.temperature) > 20"
  
			action {
			  clear_timer {
				name = "test_timer_name"
			  }
			}
		  }
	  }
      on_exit {
		event {
			name      = "test_event_name"
			condition = "convert(Decimal, $input.test_input.temperature) > 20"
  
			action {
			  clear_timer {
				name = "test_timer_name"
			  }
			}
		  }	
      }
      on_input {
		event {
			name      = "test_event_name"
			condition = "convert(Decimal, $input.test_input.temperature) > 20"
  
			action {
			  clear_timer {
				name = "test_timer_name"
			  }
			}
		  }
        transition_event {
          name      = "test_transition_event_name"
          condition = "convert(Decimal, $input.test_input.temperature) > 20"
		  next_state = "second_state"

          action {
            clear_timer {
              name = "test_timer_name"
            }
          }
		}
      }
	}
	
    state {
		name = "second_state"
		on_enter {
		  event {
			  name      = "test_event_name"
			  condition = "convert(Decimal, $input.test_input.temperature) > 20"
	
			  action {
				clear_timer {
				  name = "test_timer_name"
				}
			  }
			}
		}
	  }

  }
}
`, dName)
}
