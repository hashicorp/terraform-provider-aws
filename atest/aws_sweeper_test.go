package atest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMain(m *testing.M) {
	sweeperAwsClients = make(map[string]interface{})
	resource.TestMain(m)
}
