package glog

import (
	"github.com/nbvghost/gweb/therad"
	"testing"
	"time"
)

func BenchmarkTrace(b *testing.B) {
	Param.Debug=false
	Param.ServerAddr=":9090"
	Param.FileStorage =true

	therad.NewCoroutine(func() {
		StartLogger(Param)
	}, func(v interface{}, stack []byte) {
		Debug(v)
		Debug(string(stack))

	})

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