// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccTimestreamInfluxDB_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"DB Instance": {
			acctest.CtBasic:                testAccTimestreamInfluxDBDBInstance_basic,
			"deploymentTypeMultiAZStandby": testAccTimestreamInfluxDBDBInstance_deploymentTypeMultiAzStandby,
			acctest.CtDisappears:           testAccTimestreamInfluxDBDBInstance_disappears,
			"logDeliveryConfiguration":     testAccTimestreamInfluxDBDBInstance_logDeliveryConfiguration,
			"publiclyAccessible":           testAccTimestreamInfluxDBDBInstance_publiclyAccessible,
			"tags":                         testAccTimestreamInfluxDBDBInstance_tagsSerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
