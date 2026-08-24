// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMailManager_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"IngressPoint": {
			acctest.CtBasic:                testAccMailManagerIngressPoint_basic,
			acctest.CtDisappears:           testAccMailManagerIngressPoint_disappears,
			"update":                       testAccMailManagerIngressPoint_update,
			"tlsPolicy":                    testAccMailManagerIngressPoint_tlsPolicy,
			"type":                         testAccMailManagerIngressPoint_type,
			"networkConfiguration_public":  testAccMailManagerIngressPoint_networkConfiguration_public,
			"networkConfiguration_private": testAccMailManagerIngressPoint_networkConfiguration_private,
			"ingressPointConfiguration_smtpPasswordWO": testAccMailManagerIngressPoint_ingressPointConfiguration_smtpPasswordWO,
			"ingressPointConfiguration_tlsAuth":        testAccMailManagerIngressPoint_ingressPointConfiguration_tlsAuth,
			"Identity":                                 testAccMailManagerIngressPoint_identitySerial,
			"Tags":                                     testAccMailManagerIngressPoint_tagsSerial,
			"List":                                     testAccMailManagerIngressPoint_listSerial,
		},
		"Relay": {
			"Authentication_secretARN": testAccMailManagerRelay_Authentication_secretARN,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
