// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestInstanceMigrateState(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         any
	}{
		"v0.3.6 and earlier": {
			StateVersion: 0,
			Attributes: map[string]string{
				// EBS
				"block_device.#": "2",
				"block_device.3851383343.delete_on_termination": acctest.CtTrue,
				"block_device.3851383343.device_name":           "/dev/sdx",
				"block_device.3851383343.encrypted":             acctest.CtFalse,
				"block_device.3851383343.snapshot_id":           "",
				"block_device.3851383343.virtual_name":          "",
				"block_device.3851383343.volume_size":           "5",
				"block_device.3851383343.volume_type":           "standard",
				// Ephemeral
				"block_device.3101711606.delete_on_termination": acctest.CtFalse,
				"block_device.3101711606.device_name":           "/dev/sdy",
				"block_device.3101711606.encrypted":             acctest.CtFalse,
				"block_device.3101711606.snapshot_id":           "",
				"block_device.3101711606.virtual_name":          "ephemeral0",
				"block_device.3101711606.volume_size":           "",
				"block_device.3101711606.volume_type":           "",
				// Root
				"block_device.56575650.delete_on_termination": acctest.CtTrue,
				"block_device.56575650.device_name":           "/dev/sda1",
				"block_device.56575650.encrypted":             acctest.CtFalse,
				"block_device.56575650.snapshot_id":           "",
				"block_device.56575650.volume_size":           "10",
				"block_device.56575650.volume_type":           "standard",
			},
			Expected: map[string]string{
				"ebs_block_device.#": "1",
				"ebs_block_device.3851383343.delete_on_termination":  acctest.CtTrue,
				"ebs_block_device.3851383343.device_name":            "/dev/sdx",
				"ebs_block_device.3851383343.encrypted":              acctest.CtFalse,
				"ebs_block_device.3851383343.snapshot_id":            "",
				"ebs_block_device.3851383343.volume_size":            "5",
				"ebs_block_device.3851383343.volume_type":            "standard",
				"ephemeral_block_device.#":                           "1",
				"ephemeral_block_device.2458403513.device_name":      "/dev/sdy",
				"ephemeral_block_device.2458403513.virtual_name":     "ephemeral0",
				"root_block_device.#":                                "1",
				"root_block_device.3018388612.delete_on_termination": acctest.CtTrue,
				"root_block_device.3018388612.device_name":           "/dev/sda1",
				"root_block_device.3018388612.snapshot_id":           "",
				"root_block_device.3018388612.volume_size":           "10",
				"root_block_device.3018388612.volume_type":           "standard",
			},
		},
		"v0.3.7": {
			StateVersion: 0,
			Attributes: map[string]string{
				// EBS
				"block_device.#": "2",
				"block_device.3851383343.delete_on_termination": acctest.CtTrue,
				"block_device.3851383343.device_name":           "/dev/sdx",
				"block_device.3851383343.encrypted":             acctest.CtFalse,
				"block_device.3851383343.snapshot_id":           "",
				"block_device.3851383343.virtual_name":          "",
				"block_device.3851383343.volume_size":           "5",
				"block_device.3851383343.volume_type":           "standard",
				"block_device.3851383343.iops":                  "",
				// Ephemeral
				"block_device.3101711606.delete_on_termination": acctest.CtFalse,
				"block_device.3101711606.device_name":           "/dev/sdy",
				"block_device.3101711606.encrypted":             acctest.CtFalse,
				"block_device.3101711606.snapshot_id":           "",
				"block_device.3101711606.virtual_name":          "ephemeral0",
				"block_device.3101711606.volume_size":           "",
				"block_device.3101711606.volume_type":           "",
				"block_device.3101711606.iops":                  "",
				// Root
				"root_block_device.#":                                "1",
				"root_block_device.3018388612.delete_on_termination": acctest.CtTrue,
				"root_block_device.3018388612.device_name":           "/dev/sda1",
				"root_block_device.3018388612.snapshot_id":           "",
				"root_block_device.3018388612.volume_size":           "10",
				"root_block_device.3018388612.volume_type":           "io1",
				"root_block_device.3018388612.iops":                  "1000",
			},
			Expected: map[string]string{
				"ebs_block_device.#": "1",
				"ebs_block_device.3851383343.delete_on_termination":  acctest.CtTrue,
				"ebs_block_device.3851383343.device_name":            "/dev/sdx",
				"ebs_block_device.3851383343.encrypted":              acctest.CtFalse,
				"ebs_block_device.3851383343.snapshot_id":            "",
				"ebs_block_device.3851383343.volume_size":            "5",
				"ebs_block_device.3851383343.volume_type":            "standard",
				"ephemeral_block_device.#":                           "1",
				"ephemeral_block_device.2458403513.device_name":      "/dev/sdy",
				"ephemeral_block_device.2458403513.virtual_name":     "ephemeral0",
				"root_block_device.#":                                "1",
				"root_block_device.3018388612.delete_on_termination": acctest.CtTrue,
				"root_block_device.3018388612.device_name":           "/dev/sda1",
				"root_block_device.3018388612.snapshot_id":           "",
				"root_block_device.3018388612.volume_size":           "10",
				"root_block_device.3018388612.volume_type":           "io1",
				"root_block_device.3018388612.iops":                  "1000",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "i-abc123",
			Attributes: tc.Attributes,
		}
		is, err := tfec2.InstanceMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
		}
	}
}

func TestInstanceMigrateState_empty(t *testing.T) {
	t.Parallel()

	var is *terraform.InstanceState
	var meta any

	// should handle nil
	is, err := tfec2.InstanceMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	_, err = tfec2.InstanceMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}

func TestInstanceStateUpgradeV1(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawState map[string]any
		expected map[string]any
	}{
		{
			name:     "empty rawState",
			rawState: nil,
			expected: map[string]any{},
		},
		{
			name: "no cpu_options, only cpu_core_count",
			rawState: map[string]any{
				"cpu_core_count": 4,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count": 4,
					},
				},
			},
		},
		{
			name: "no cpu_options, only cpu_threads_per_core",
			rawState: map[string]any{
				"cpu_threads_per_core": 2,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"threads_per_core": 2,
					},
				},
			},
		},
		{
			name: "no cpu_options, both cpu_core_count and cpu_threads_per_core",
			rawState: map[string]any{
				"cpu_core_count":       4,
				"cpu_threads_per_core": 2,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count":       4,
						"threads_per_core": 2,
					},
				},
			},
		},
		{
			name: "existing cpu_options with core_count and threads_per_core",
			rawState: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count":       8,
						"threads_per_core": 4,
					},
				},
				"cpu_core_count":       4,
				"cpu_threads_per_core": 2,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count":       8,
						"threads_per_core": 4,
					},
				},
			},
		},
		{
			name: "existing cpu_options with only core_count",
			rawState: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count": 8,
					},
				},
				"cpu_threads_per_core": 2,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count":       8,
						"threads_per_core": 2,
					},
				},
			},
		},
		{
			name: "existing cpu_options with only threads_per_core",
			rawState: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"threads_per_core": 4,
					},
				},
				"cpu_core_count": 8,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count":       8,
						"threads_per_core": 4,
					},
				},
			},
		},
		{
			name: "no cpu_options and no cpu_core_count or cpu_threads_per_core",
			rawState: map[string]any{
				"some_other_key": "some_value",
			},
			expected: map[string]any{
				"some_other_key": "some_value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tfec2.InstanceStateUpgradeV1(context.Background(), tt.rawState, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

// TestInstanceStateUpgradeV1_complexState simulates state as decoded from a JSON state file.
// This is to ensure nothing in a complex state prevents state upgrading.
func TestInstanceStateUpgradeV1_complexState(t *testing.T) {
	t.Parallel()

	stateJSON := `{
		"ami": "ami-XXXXXXXXXXXXXXXXX",
		"arn": "arn:aws:ec2:eu-central-1:XXXXXXXXXXXX:instance/i-XXXXXXXXXXXXXXXXX",
		"associate_public_ip_address": false,
		"availability_zone": "eu-central-1a",
		"capacity_reservation_specification": [{
			"capacity_reservation_preference": "open",
			"capacity_reservation_target": []
		}],
		"cpu_core_count": 2,
		"cpu_options": [{
			"amd_sev_snp": "",
			"core_count": 2,
			"threads_per_core": 1
		}],
		"cpu_threads_per_core": 1,
		"credit_specification": [],
		"disable_api_stop": false,
		"disable_api_termination": false,
		"ebs_block_device": [{
			"delete_on_termination": true,
			"device_name": "/dev/xvda",
			"encrypted": true,
			"iops": 3000,
			"kms_key_id": "arn:aws:kms:eu-central-1:XXXXXXXXXXXX:key/XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
			"snapshot_id": "",
			"tags": {},
			"tags_all": {},
			"throughput": 125,
			"volume_id": "vol-XXXXXXXXXXXXXXXXX",
			"volume_size": 100,
			"volume_type": "gp3"
		}],
		"ebs_optimized": false,
		"enable_primary_ipv6": null,
		"enclave_options": [{
			"enabled": false
		}],
		"ephemeral_block_device": [],
		"get_password_data": false,
		"hibernation": false,
		"host_id": "",
		"host_resource_group_arn": null,
		"iam_instance_profile": "xxx",
		"id": "i-XXXXXXXXXXXXXXXXX",
		"instance_initiated_shutdown_behavior": "stop",
		"instance_lifecycle": "",
		"instance_market_options": [],
		"instance_state": "running",
		"instance_type": "m6g.large",
		"ipv6_address_count": 0,
		"ipv6_addresses": [],
		"key_name": "xxx",
		"launch_template": [],
		"maintenance_options": [{
			"auto_recovery": "default"
		}],
		"metadata_options": [{
			"http_endpoint": "enabled",
			"http_protocol_ipv6": "disabled",
			"http_put_response_hop_limit": 1,
			"http_tokens": "required",
			"instance_metadata_tags": "disabled"
		}],
		"monitoring": false,
		"network_interface": [],
		"outpost_arn": "",
		"password_data": "",
		"placement_group": "",
		"placement_partition_number": 0,
		"primary_network_interface_id": "eni-XXXXXXXXXXXXXXXXX",
		"private_dns": "ip-X-X-X-X.eu-central-1.compute.internal",
		"private_dns_name_options": [{
			"enable_resource_name_dns_a_record": false,
			"enable_resource_name_dns_aaaa_record": false,
			"hostname_type": "ip-name"
		}],
		"private_ip": "X.X.X.X",
		"public_dns": "",
		"public_ip": "",
		"root_block_device": [{
			"delete_on_termination": true,
			"device_name": "/dev/xvda",
			"encrypted": true,
			"iops": 3000,
			"kms_key_id": "arn:aws:kms:eu-central-1:XXXXXXXXXXXX:key/XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
			"tags": {},
			"tags_all": {},
			"throughput": 125,
			"volume_id": "vol-XXXXXXXXXXXXXXXXX",
			"volume_size": 100,
			"volume_type": "gp3"
		}],
		"secondary_private_ips": [],
		"security_groups": [],
		"source_dest_check": true,
		"spot_instance_request_id": "",
		"subnet_id": "subnet-XXXXXXXXXXXXXXXXXXX",
		"tags": {},
		"tags_all": {},
		"tenancy": "default",
		"timeouts": null,
		"user_data": null,
		"user_data_base64": "xxx",
		"user_data_replace_on_change": false,
		"volume_tags": null,
		"vpc_security_group_ids": ["sg-XXXXXXXXXXXXXXXXX"]
	}` //lintignore:AWSAT003,AWSAT005

	var rawState map[string]any
	if err := json.Unmarshal([]byte(stateJSON), &rawState); err != nil {
		t.Fatalf("failed to unmarshal state JSON: %v", err)
	}

	result, err := tfec2.InstanceStateUpgradeV1(context.Background(), rawState, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := result["cpu_core_count"]; ok {
		t.Errorf("expected cpu_core_count to be removed, but it is still present")
	}
	if _, ok := result["cpu_threads_per_core"]; ok {
		t.Errorf("expected cpu_threads_per_core to be removed, but it is still present")
	}
	if result[names.AttrInstanceType] != "m6g.large" {
		t.Errorf("expected instance_type to be preserved, got: %v", result[names.AttrInstanceType])
	}
}
