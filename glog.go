package glog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nbvghost/gweb/therad"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

//queue
//日志列对
var _logChanQueue =make(chan string,100000)
var _logFileTempChan =make(chan string,1000)
var _logServerStatus =make(chan bool)


var glogServer =&TCP{}
type ParamValue struct {
	ServerAddr string
	ServerName string
	LogFileName string
	Debug bool
	PrintStack bool
	FileStorage bool
}
var Param = &ParamValue{
	ServerAddr:"",
	ServerName:"default",
	LogFileName:"glog",
	Debug:true,
	PrintStack:false,
	FileStorage:false,
}



var _glogOut =log.New(os.Stdout,"[TRACE] ",log.LstdFlags|log.LUTC|log.Lshortfile)
var _glogErr =log.New(os.Stderr,"[ERROR] ",log.LstdFlags|log.LUTC|log.Lshortfile)
var _glogDebug =log.New(os.Stdout,"[DEBUG] ",log.LstdFlags|log.LUTC|log.Lshortfile)

func Debugf(format string, v ...interface{})  {
	if Param.Debug{
		_glogDebug.Output(2, fmt.Sprintf(format,v...))
	}
}
func Debug(v ...interface{})  {
	if Param.Debug{
		_glogDebug.Output(2, fmt.Sprintln(v...))
	}
}
func Trace(v ...interface{}) {
	//outText:=fmt.Sprintln(v...)
	_, file, line, ok := runtime.Caller(1)
	if !ok{
		file = "unknown"
		line = 0
	}
	if Param.Debug{
		 _glogOut.Output(2, fmt.Sprintln(v...))
	}

	for index:=range v{
		b,_:=json.Marshal(map[string]interface{}{
			"Time":time.Now().Format("2006-01-02 15:04:05"),
			"File":file,
			"Line":line,
			"TRACE":v[index],
			"ServerName":Param.ServerName,
		})
		_logChanQueue<-string(b)
	}



	/*go func(_file string, _line int,_v []interface{}) {


		b,_:=json.Marshal(map[string]interface{}{
			"Time":time.Now().Format("2006-01-02 15:04:05"),
			"File":_file,
			"Line":_line,
			"TRACE":_v,
			"ServerName":Param.ServerName,
		})
		_logChanQueue<-string(b)


	}(file,line,v)*/

}

func Error(err error) bool {
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		if !ok{
			file = "unknown"
			line = 0
		}
		if Param.Debug{
			_glogErr.Output(2, fmt.Sprintln(err))
			if Param.PrintStack{
				debug.PrintStack()
			}
		}
		//log.Println(file, line, err)
		b,_:=json.Marshal(map[string]interface{}{
			"Time":time.Now().Format("2006-01-02 15:04:05"),
			"File":file,
			"Line":line,
			"ERROR":err,
			"ServerName":Param.ServerName,
		})
		//conf.LogQueue=append(conf.LogQueue,string(b))
		_logChanQueue<-string(b)


		/*go func(_file string, _line int,_err error) {
			//util.Trace(funcName,file,line,ok)
				//log.Println(file, line, err)
				b,_:=json.Marshal(map[string]interface{}{
					"Time":time.Now().Format("2006-01-02 15:04:05"),
					"File":_file,
					"Line":_line,
					"ERROR":_err,
					"ServerName":Param.ServerName,
				})
				//conf.LogQueue=append(conf.LogQueue,string(b))
				_logChanQueue<-string(b)


		}(file,line,err)*/

		return true
	}else{
		return false
	}
}
func init()  {

	therad.NewCoroutine(func() {
		<-_startChan
		if strings.EqualFold(Param.ServerAddr,"")==false{
			glogServer.StartTCP(Param.ServerAddr,_logServerStatus)
		}
	}, func(v interface{}, stack []byte) {
		Debug(v)
		Debug(string(stack))

	})

	//日志服务
	therad.NewCoroutine(func() {
		for v := range _logChanQueue {

			if strings.EqualFold(Param.ServerAddr,"")==false{
				err:=glogServer.Write(v)
				if err!=nil{
					_logFileTempChan<-v
				}
			}
		}
	}, func(v interface{}, stack []byte) {
		Debug(v)
		Debug(string(stack))

	})

	//日志服务
	therad.NewCoroutine(func() {
		<-_startChan
		logFileName:=Param.LogFileName+"_glog_"+time.Now().Format("2006_01_02")+".log"

		var logFile *os.File
		logFile,_= os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)

		//var isLogServerOk =false
		for{
			select {
			case v:=<-time.NewTicker(1*time.Hour).C:
				fmt.Println(v)

				logFile,_= os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)

			case v:=<-_logServerStatus:

				if v{

					bf,err:=ioutil.ReadAll(logFile)
					if err!=nil{
						return
					}
					fd := bytes.NewBuffer(bf)
					logFile.Close()
					for{
						l, err := fd.ReadString('\n')
						if err != nil {
							if err == io.EOF {
								lf, _ := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
								lf.Truncate(0)
								lf.Sync()
								lf.Close()
								logFile, _ = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
							}
							break
						}
						_logChanQueue<-l
					}


					//if isLogServerOk==false{



					//}
					//isLogServerOk=true
					//logFile.Sync()
					//logFile.Close()
					//logFile=nil
				}else{
					//isLogServerOk=false
				}
			case v :=<-_logFileTempChan:
				//log.Println(v)
				if Param.FileStorage{
					//ioutil.WriteFile(logFileName,[]byte(fmt.Sprintln( v)),os.ModeAppend)
					if logFile!=nil{
						logFile.WriteString(fmt.Sprintln( v))
						logFile.Sync()
					}
				}


			}
		}
		//logFile.Sync()
	}, func(v interface{}, stack []byte) {
		Debug(v)
		Debug(string(stack))

	})

}
var _startChan =make(chan bool)

func StartLogger(_param *ParamValue){
	if _param!=nil{
		Param = _param
	}
	_startChan<-true
	_startChan<-true
}