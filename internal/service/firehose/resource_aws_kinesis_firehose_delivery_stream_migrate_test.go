package firehose_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAWSKinesisFirehoseMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0.6.16 and earlier": {
			StateVersion: 0,
			Attributes: map[string]string{
				// EBS
				"role_arn":            "arn:aws:iam::somenumber:role/tf_acctest_4271506651559170635", //lintignore:AWSAT005
				"s3_bucket_arn":       "arn:aws:s3:::tf-test-bucket",                                 //lintignore:AWSAT005
				"s3_buffer_interval":  "400",
				"s3_buffer_size":      "10",
				"s3_data_compression": "GZIP",
			},
			Expected: map[string]string{
				"s3_configuration.#":                    "1",
				"s3_configuration.0.bucket_arn":         "arn:aws:s3:::tf-test-bucket", //lintignore:AWSAT005
				"s3_configuration.0.buffer_interval":    "400",
				"s3_configuration.0.buffer_size":        "10",
				"s3_configuration.0.compression_format": "GZIP",
				"s3_configuration.0.role_arn":           "arn:aws:iam::somenumber:role/tf_acctest_4271506651559170635", //lintignore:AWSAT005
			},
		},
		"v0.6.16 and earlier, sparse": {
			StateVersion: 0,
			Attributes: map[string]string{
				// EBS
				"role_arn":      "arn:aws:iam::somenumber:role/tf_acctest_4271506651559170635", //lintignore:AWSAT005
				"s3_bucket_arn": "arn:aws:s3:::tf-test-bucket",                                 //lintignore:AWSAT005
			},
			Expected: map[string]string{
				"s3_configuration.#":            "1",
				"s3_configuration.0.bucket_arn": "arn:aws:s3:::tf-test-bucket",                                 //lintignore:AWSAT005
				"s3_configuration.0.role_arn":   "arn:aws:iam::somenumber:role/tf_acctest_4271506651559170635", //lintignore:AWSAT005
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "i-abc123",
			Attributes: tc.Attributes,
		}
		is, err := resourceAwsKinesisFirehoseMigrateState(
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

func TestAWSKinesisFirehoseMigrateState_empty(t *testing.T) {
	var is *terraform.InstanceState
	var meta interface{}

	// should handle nil
	is, err := resourceAwsKinesisFirehoseMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	_, err = resourceAwsInstanceMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}

func migrateAwsInstanceStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)

	// Delete old count
	delete(is.Attributes, "block_device.#")

	oldBds, err := readV0BlockDevices(is)
	if err != nil {
		return is, err
	}
	// seed count fields for new types
	is.Attributes["ebs_block_device.#"] = "0"
	is.Attributes["ephemeral_block_device.#"] = "0"
	// depending on if state was v0.3.7 or an earlier version, it might have
	// root_block_device defined already
	if _, ok := is.Attributes["root_block_device.#"]; !ok {
		is.Attributes["root_block_device.#"] = "0"
	}
	for _, oldBd := range oldBds {
		writeV1BlockDevice(is, oldBd)
	}
	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}

func resourceAwsInstanceMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found AWS Instance State v0; migrating to v1")
		return migrateAwsInstanceStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}
