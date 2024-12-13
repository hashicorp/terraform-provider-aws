package main

import (
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
)

func test1() {
	// ruleid: map_block_key-meaningful-names
	attributes := map[string]schema.Attribute{
		"map_block_key": schema.StringAttribute{
			Required: true,
		},
	}
}

func test2() {
	// ok: map_block_key-meaningful-names
	attributes := map[string]schema.Attribute{
		"day_of_week": schema.StringAttribute{
			Required: true,
		},
	}
}
