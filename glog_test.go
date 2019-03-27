package glog

import (
	"testing"
	"time"
)

func BenchmarkTrace(b *testing.B) {
	Param.Debug=false
	Param.ServerAddr=":9090"
	Param.FileStorage =true
	go func() {

		StartLogger(Param)
	}()


	for i:=0;i<b.N;i++{
		Trace(map[string]interface{}{"dfs":54})
	}


}
func TestTrace(t *testing.T) {
	Param.Debug=true
	Param.ServerAddr=":9090"
	Param.FileStorage =true

	StartLogger(Param)

	for {
		Trace("sdfdsf")
		time.Sleep(1*time.Second)
	}


}