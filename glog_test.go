package glog

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {

	Param.FileStorage = false
	Param.StandardOut = true
	Param.FormatType = JSON
	Start()

	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Test", args: args{err: errors.New("test error")}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Error(tt.args.err)
			Debug("dsfdsf", "dsfdsfsd", "dsfds", map[string]int{"dsfdsfds": 454})
			Trace(map[string]int{"dsfdsfds": 454})
			Trace("sdfds", 45454, map[string]int{"dsfdsfds": 454})
			Panic(tt.args.err)

		})
	}

	for {
	}
}
