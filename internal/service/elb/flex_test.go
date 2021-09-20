package elb

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
)

func TestExpandListeners(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"instance_port":     8000,
			"lb_port":           80,
			"instance_protocol": "http",
			"lb_protocol":       "http",
		},
		map[string]interface{}{
			"instance_port":      8000,
			"lb_port":            80,
			"instance_protocol":  "https",
			"lb_protocol":        "https",
			"ssl_certificate_id": "something",
		},
	}
	listeners, err := expandListeners(expanded)
	if err != nil {
		t.Fatalf("bad: %#v", err)
	}

	expected := &elb.Listener{
		InstancePort:     aws.Int64(int64(8000)),
		LoadBalancerPort: aws.Int64(int64(80)),
		InstanceProtocol: aws.String("http"),
		Protocol:         aws.String("http"),
	}

	if !reflect.DeepEqual(listeners[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			listeners[0],
			expected)
	}
}

// this test should produce an error from expandlisteners on an invalid
// combination
func TestExpandListeners_invalid(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"instance_port":      8000,
			"lb_port":            80,
			"instance_protocol":  "http",
			"lb_protocol":        "http",
			"ssl_certificate_id": "something",
		},
	}
	_, err := expandListeners(expanded)
	if err != nil {
		// Check the error we got
		if !strings.Contains(err.Error(), "ssl_certificate_id may be set only when protocol") {
			t.Fatalf("Got error in TestExpandListeners_invalid, but not what we expected: %s", err)
		}
	}

	if err == nil {
		t.Fatalf("Expected TestExpandListeners_invalid to fail, but passed")
	}
}

func TestFlattenHealthCheck(t *testing.T) {
	cases := []struct {
		Input  *elb.HealthCheck
		Output []map[string]interface{}
	}{
		{
			Input: &elb.HealthCheck{
				UnhealthyThreshold: aws.Int64(int64(10)),
				HealthyThreshold:   aws.Int64(int64(10)),
				Target:             aws.String("HTTP:80/"),
				Timeout:            aws.Int64(int64(30)),
				Interval:           aws.Int64(int64(30)),
			},
			Output: []map[string]interface{}{
				{
					"unhealthy_threshold": int64(10),
					"healthy_threshold":   int64(10),
					"target":              "HTTP:80/",
					"timeout":             int64(30),
					"interval":            int64(30),
				},
			},
		},
	}

	for _, tc := range cases {
		output := flattenHealthCheck(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}

func TestExpandInstanceString(t *testing.T) {

	expected := []*elb.Instance{
		{InstanceId: aws.String("test-one")},
		{InstanceId: aws.String("test-two")},
	}

	ids := []interface{}{
		"test-one",
		"test-two",
	}

	expanded := expandInstanceString(ids)

	if !reflect.DeepEqual(expanded, expected) {
		t.Fatalf("Expand Instance String output did not match.\nGot:\n%#v\n\nexpected:\n%#v", expanded, expected)
	}
}

func TestExpandPolicyAttributes(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"name":  "Protocol-TLSv1",
			"value": "false",
		},
		map[string]interface{}{
			"name":  "Protocol-TLSv1.1",
			"value": "false",
		},
		map[string]interface{}{
			"name":  "Protocol-TLSv1.2",
			"value": "true",
		},
	}
	attributes := expandPolicyAttributes(expanded)

	if len(attributes) != 3 {
		t.Fatalf("expected number of attributes to be 3, but got %d", len(attributes))
	}

	expected := &elb.PolicyAttribute{
		AttributeName:  aws.String("Protocol-TLSv1.2"),
		AttributeValue: aws.String("true"),
	}

	if !reflect.DeepEqual(attributes[2], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			attributes[2],
			expected)
	}
}

func TestExpandPolicyAttributes_empty(t *testing.T) {
	var expanded []interface{}

	attributes := expandPolicyAttributes(expanded)

	if len(attributes) != 0 {
		t.Fatalf("expected number of attributes to be 0, but got %d", len(attributes))
	}
}

func TestExpandPolicyAttributes_invalid(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"name":  "Protocol-TLSv1.2",
			"value": "true",
		},
	}
	attributes := expandPolicyAttributes(expanded)

	expected := &elb.PolicyAttribute{
		AttributeName:  aws.String("Protocol-TLSv1.2"),
		AttributeValue: aws.String("false"),
	}

	if reflect.DeepEqual(attributes[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			attributes[0],
			expected)
	}
}

func TestFlattenPolicyAttributes(t *testing.T) {
	cases := []struct {
		Input  []*elb.PolicyAttributeDescription
		Output []interface{}
	}{
		{
			Input: []*elb.PolicyAttributeDescription{
				{
					AttributeName:  aws.String("Protocol-TLSv1.2"),
					AttributeValue: aws.String("true"),
				},
			},
			Output: []interface{}{
				map[string]string{
					"name":  "Protocol-TLSv1.2",
					"value": "true",
				},
			},
		},
	}

	for _, tc := range cases {
		output := flattenPolicyAttributes(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}
