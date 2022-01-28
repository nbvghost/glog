package glog

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {

	Param.Tag = "dsfsd"
	Param.AppName = "JSON"
	Param.FormatType = MLF
	Param.StandardOut = true
	Param.FileStorage = true
	Param.ShowHeader = true
	Param.Level = AllLevel

	Start()

	var err = errors.New("dfssdfds")

	Error(err)
	Debug("dsfdsf", "dsfdsfsd", "dsfds", map[string]int{"dsfdsfds": 454})
	Trace(map[string]int{"dsfdsfds": 454})
	Trace("sdfds", 45454, map[string]int{"dsfdsfds": 454})
	Warning(err)
	CheckError(err)
	Panic(err)

	myLogger := NewLogger("54545")
	myLogger.Error(err)
	myLogger.Debug("dsfdsf", "dsfdsfsd", "dsfds", map[string]int{"dsfdsfds": 454})
	myLogger.Trace(map[string]int{"dsfdsfds": 454})
	myLogger.Trace("sdfds", 45454, map[string]int{"dsfdsfds": 454})
	myLogger.Panic(err)

	for {

	}
}
