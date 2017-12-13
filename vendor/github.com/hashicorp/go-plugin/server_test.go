package plugin

import (
	"io/ioutil"
	"net"
	"os"
	"testing"
)

func TestRmListener_impl(t *testing.T) {
	var _ net.Listener = new(rmListener)
}

func TestRmListener(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tf, err := ioutil.TempFile("", "plugin")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	path := tf.Name()

	// Close the file
	if err := tf.Close(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Create the listener and test close
	rmL := &rmListener{
		Listener: l,
		Path:     path,
	}
	if err := rmL.Close(); err != nil {
		t.Fatalf("err: %s", err)
	}

	// File should be goe
	if _, err := os.Stat(path); err == nil || !os.IsNotExist(err) {
		t.Fatalf("err: %s", err)
	}
}
