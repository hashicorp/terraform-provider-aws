package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

type pointerStruct struct {
	aBool   *bool
	aFloat  *float64
	anInt   *int64
	aString *string
	aTime   *time.Time
}

func dereferenceAndWrap() pointerStruct {
	var (
		aBool   *bool
		aFloat  *float64
		anInt   *int64
		aString *string
		aTime   *time.Time
	)

	return pointerStruct{
		// ruleid: immediate-dereference-and-wrap
		aBool: aws.Bool(*aBool),
		// ruleid: immediate-dereference-and-wrap
		aFloat: aws.Float64(*aFloat),
		// ruleid: immediate-dereference-and-wrap
		anInt: aws.Int64(*anInt),
		// ruleid: immediate-dereference-and-wrap
		aString: aws.String(*aString),
		// ruleid: immediate-dereference-and-wrap
		aTime: aws.Time(*aTime),
	}
}
