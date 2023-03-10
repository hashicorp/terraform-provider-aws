package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/schema"
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
		// ruleid: pointer-conversion-dereference-and-wrap
		aBool: aws.Bool(*aBool),
		// ruleid: pointer-conversion-dereference-and-wrap
		aFloat: aws.Float64(*aFloat),
		// ruleid: pointer-conversion-dereference-and-wrap
		anInt: aws.Int64(*anInt),
		// ruleid: pointer-conversion-dereference-and-wrap
		aString: aws.String(*aString),
		// ruleid: pointer-conversion-dereference-and-wrap
		aTime: aws.Time(*aTime),
	}
}

type valueStruct struct {
	aBool   bool
	aFloat  float64
	anInt   int64
	aString string
	aTime   time.Time
}

func wrapAndDereference() valueStruct {
	var (
		aBool   bool
		aFloat  float64
		anInt   int64
		aString string
		aTime   time.Time
	)

	return valueStruct{
		// ruleid: pointer-conversion-wrap-and-dereference
		aBool: *aws.Bool(aBool),
		// ruleid: pointer-conversion-wrap-and-dereference
		aFloat: *aws.Float64(aFloat),
		// ruleid: pointer-conversion-wrap-and-dereference
		anInt: *aws.Int64(anInt),
		// ruleid: pointer-conversion-wrap-and-dereference
		aString: *aws.String(aString),
		// ruleid: pointer-conversion-wrap-and-dereference
		aTime: *aws.Time(aTime),
	}
}

func setID(d *schema.ResourceData) {
	var value *string

	// ruleid: pointer-conversion-ResourceData-SetId
	d.SetId(*value)
}

func intConversion() int {
	var i *int64

	// ruleid: int64-pointer-dereference-and-convert
	return int(*i)
}

func assignment() string {
	var ptr *string

	// ruleid: pointer-conversion-on-assignment
	val := *ptr

	return val
}

func assignmentToRef(ref *string) {
	var ptr *string

	// ok: pointer-conversion-on-assignment
	*ref = *ptr
}
