package rds

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func compareActualEngineVersion(d *schema.ResourceData, oldVersion string, newVersion string) {
	newVersionSubstr := newVersion

	if len(newVersion) > len(oldVersion) {
		newVersionSubstr = string([]byte(newVersion)[0 : len(oldVersion)+1])
	}

	if oldVersion != newVersion && string(append([]byte(oldVersion), []byte(".")...)) != newVersionSubstr {
		d.Set("engine_version", newVersion)
		fmt.Printf("[READ/cluster] engine_version: %s\n", newVersion)
	}

	d.Set("engine_version_actual", newVersion)
	fmt.Printf("[READ/cluster] engine_version_actual: %s\n", newVersion)

}
