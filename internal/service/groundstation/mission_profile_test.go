package groundstation_test

import (
	"testing"
)

func TestResourceMissionProfile(t *testing.T) {
	r := resourceMissionProfile()
	if r.Schema["contact_post_pass_duration_seconds"].Required {
		t.Errorf("Expected contact_post_pass_duration_seconds to not be required")
	}
	if r.Schema["contact_pre_pass_duration_seconds"].Required {
		t.Errorf("Expected contact_pre_pass_duration_seconds to not be required")
	}
	if r.Schema["dataflow_edges"].Required {
		t.Errorf("Expected dataflow_edges to not be required")
	}
	if r.Schema["minimum_viable_contact_duration_seconds"].Required {
		t.Errorf("Expected minimum_viable_contact_duration_seconds to not be required")
	}
	if r.Schema["name"].Required {
		t.Errorf("Expected name to not be required")
	}
}

func TestResourceMissionProfileCreate(t *testing.T) {
	r := resourceMissionProfile()
	d := r.TestResourceData()
	d.SetId("test")
	d.Set("name", "test")
	d.Set("contact_post_pass_duration_seconds", 1)
	d.Set("contact_pre_pass_duration_seconds", 1)
	d.Set("minimum_viable_contact_duration_seconds", 1)
	d.Set("dataflow_edges", []interface{}{"test"})
	err := r.Create(d, nil)
	if err != nil {
		t.Fatalf("Expected create to succeed: %s", err)
	}
	if d.Id() != "test" {
		t.Errorf("Expected ID to not be empty")
	}
}

func TestResourceMissionProfileRead(t *testing.T) {
	r := resourceMissionProfile()
	d := r.TestResourceData()
	d.SetId("test")
	d.Set("name", "test")
	d.Set("contact_post_pass_duration_seconds", 1)
	d.Set("contact_pre_pass_duration_seconds", 1)
	d.Set("minimum_viable_contact_duration_seconds", 1)
	d.Set("dataflow_edges", []interface{}{"test"})
	err := r.Read(d, nil)
	if err != nil {
		t.Fatalf("Expected read to succeed: %s", err)
	}
	if d.Id() != "test" {
		t.Errorf("Expected ID to not be empty")
	}
}

func TestResourceMissionProfileUpdate(t *testing.T) {
	r := resourceMissionProfile()
	d := r.TestResourceData()
	d.SetId("test")
	d.Set("name", "test")
	d.Set("contact_post_pass_duration_seconds", 1)
	d.Set("contact_pre_pass_duration_seconds", 1)
	d.Set("minimum_viable_contact_duration_seconds", 1)
	d.Set("dataflow_edges", []interface{}{"test"})
	err := r.Update(d, nil)
	if err != nil {
		t.Fatalf("Expected update to succeed: %s", err)
	}
	if d.Id() != "test" {
		t.Errorf("Expected ID to not be empty")
	}
}

func TestResourceMissionProfileDelete(t *testing.T) {
	r := resourceMissionProfile()
	d := r.TestResourceData()
	d.SetId("test")
	d.Set("name", "test")
	d.Set("contact_post_pass_duration_seconds", 1)
	d.Set("contact_pre_pass_duration_seconds", 1)
	d.Set("minimum_viable_contact_duration_seconds", 1)
	d.Set("dataflow_edges", []interface{}{"test"})
	err := r.Delete(d, nil)
	if err != nil {
		t.Fatalf("Expected delete to succeed: %s", err)
	}
}
