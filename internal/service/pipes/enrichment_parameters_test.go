package pipes

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/go-cmp/cmp"
)

func Test_expandEnrichmentParameters(t *testing.T) {
	tests := map[string]struct {
		config   map[string]interface{}
		expected *types.PipeEnrichmentParameters
	}{
		"input_template config": {
			config: map[string]interface{}{
				"input_template": "some template",
			},
			expected: &types.PipeEnrichmentParameters{
				InputTemplate: aws.String("some template"),
			},
		},
		"http_parameters config": {
			config: map[string]interface{}{
				"http_parameters": []interface{}{
					map[string]interface{}{
						"path_parameters": []interface{}{"a", "b"},
						"header": []interface{}{
							map[string]interface{}{
								"key":   "key1",
								"value": "value1",
							},
							map[string]interface{}{
								"key":   "key2",
								"value": "value2",
							},
						},
						"query_string": []interface{}{
							map[string]interface{}{
								"key":   "key3",
								"value": "value3",
							},
							map[string]interface{}{
								"key":   "key4",
								"value": "value4",
							},
						},
					},
				},
			},
			expected: &types.PipeEnrichmentParameters{
				HttpParameters: &types.PipeEnrichmentHttpParameters{
					PathParameterValues: []string{"a", "b"},
					HeaderParameters: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					QueryStringParameters: map[string]string{
						"key3": "value3",
						"key4": "value4",
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := expandEnrichmentParameters([]interface{}{tt.config})

			if diff := cmp.Diff(got, tt.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func Test_flattenEnrichmentParameters(t *testing.T) {
	tests := map[string]struct {
		config   *types.PipeEnrichmentParameters
		expected []map[string]interface{}
	}{
		"input_template config": {
			config: &types.PipeEnrichmentParameters{
				InputTemplate: aws.String("some template"),
			},
			expected: []map[string]interface{}{
				{
					"input_template": "some template",
				},
			},
		},
		"http_parameters config": {
			config: &types.PipeEnrichmentParameters{
				HttpParameters: &types.PipeEnrichmentHttpParameters{
					PathParameterValues: []string{"a", "b"},
					HeaderParameters: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					QueryStringParameters: map[string]string{
						"key3": "value3",
						"key4": "value4",
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"http_parameters": []map[string]interface{}{
						{
							"path_parameters": []interface{}{"a", "b"},
							"header": []map[string]interface{}{
								{
									"key":   "key1",
									"value": "value1",
								},
								{
									"key":   "key2",
									"value": "value2",
								},
							},
							"query_string": []map[string]interface{}{
								{
									"key":   "key3",
									"value": "value3",
								},
								{
									"key":   "key4",
									"value": "value4",
								},
							},
						},
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := flattenEnrichmentParameters(tt.config)

			if diff := cmp.Diff(got, tt.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
