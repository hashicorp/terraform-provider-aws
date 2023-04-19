package opsworks

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// This function sorts List A to look like a list found in the tf file.
func sortListBasedonTFFile(in []string, d *schema.ResourceData) ([]string, error) {
	listName := "layer_ids"
	if attributeCount, ok := d.Get(listName + ".#").(int); ok {
		for i := 0; i < attributeCount; i++ {
			currAttributeId := d.Get(listName + "." + strconv.Itoa(i))
			for j := 0; j < len(in); j++ {
				if currAttributeId == in[j] {
					in[i], in[j] = in[j], in[i]
				}
			}
		}
		return in, nil
	}
	return in, fmt.Errorf("Could not find list: %s", listName)
}
