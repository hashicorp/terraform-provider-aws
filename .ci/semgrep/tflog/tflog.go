package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func noAssignment() {
	ctx := context.Background()

	// ruleid: setfield-without-assign
	tflog.SetField(ctx, "field", "value")
}

func assigned() {
	ctx := context.Background()

	// ok: setfield-without-assign
	ctx = tflog.SetField(ctx, "field", "value")
}

func returnedContext() context.Context {
	ctx := context.Background()

	// ok: setfield-without-assign
	return tflog.SetField(ctx, "field", "value")
}

func declareAndAssign_SameName() {
	ctx := context.Background()

	for i := 0; i < 1; i++ {
		// ok: setfield-without-assign
		ctx := tflog.SetField(ctx, "field", "value")
	}
}

func declareAndAssign_Rename() {
	outerCtx := context.Background()

	for i := 0; i < 1; i++ {
		// ok: setfield-without-assign
		innerCtx := tflog.SetField(outerCtx, "field", "value")
	}
}
