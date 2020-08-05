package glog

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {

	Param.FileStorage = true
	StartLogger(Param)

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
			//Error(tt.args.err)
			//Debug("dsfdsf")
			//Panic(tt.args.err)
			Trace("dsfds", map[string]int{"dsfdsfds": 454})

		})
	}

	select {}
}
