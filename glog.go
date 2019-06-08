package glog

import (
	"encoding/json"
	"fmt"
	"github.com/nbvghost/gweb/thread"

	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

//queue
//日志列对
var _logChanQueue =make(chan string,1000)
var _logFileTempChan =make(chan string,1000)
var _logServerStatus =make(chan bool)


var glogServer =&GlogTCP{}
type ParamValue struct {
	ServerAddr string
	ServerName string
	LogFilePath string
	Debug bool
	PrintStack bool
	FileStorage bool
}
var Param = &ParamValue{
	ServerAddr:"",
	ServerName:"default",
	LogFilePath:"glog",
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
	/*for index:=range v{
		b,_:=json.Marshal(map[string]interface{}{
			"Time":time.Now().Format("2006-01-02 15:04:05"),
			"File":file,
			"Line":line,
			"TRACE":v[index],
			"ServerName":Param.ServerName,
		})
		_logChanQueue<-string(b)
	}*/

	filePaths:=strings.Split(file,"/")
	if len(filePaths)==0{
		filePaths=[]string{file}
	}
	b,_:=json.Marshal(map[string]interface{}{
		"Time":time.Now().Format("2006-01-02 15:04:05"),
		"File":filePaths[len(filePaths)-1],
		"Line":line,
		"TRACE":v,
		"ServerName":Param.ServerName,
	})
	_logChanQueue<-string(b)


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
			"ERROR":err.Error(),
			"ServerName":Param.ServerName,
		})
		//conf.LogQueue=append(conf.LogQueue,string(b))
		_logChanQueue<-string(b)

		/*go func(_file string, _line int,_err error) max_relay_log_size{
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
func getLogFileName(v time.Time) string {
	logFileName:=""
	if strings.EqualFold(Param.LogFilePath,""){
		logFileName=Param.ServerName+"/"+v.Format("2006_01_02")+".log"
		err:=os.MkdirAll(Param.ServerName,os.ModePerm)
		if err!=nil{
			log.Println(err)
		}
	}else{
		logFileName=Param.LogFilePath+"/"+Param.ServerName+"/"+v.Format("2006_01_02")+".log"
		err:=os.MkdirAll(Param.LogFilePath+"/"+Param.ServerName,os.ModePerm)
		if err!=nil{
			log.Println(err)
		}
	}
	return logFileName
}
func init()  {

	thread.NewCoroutine(func() {
		<-_startChan
		if strings.EqualFold(Param.ServerAddr,"")==false{
			glogServer.StartTCP(Param.ServerAddr,_logServerStatus)
		}
	}, func(v interface{}, stack []byte) {
		log.Println(v)
		log.Println(string(stack))
	})

	//日志服务
	thread.NewCoroutine(func() {
		for v := range _logChanQueue {
			if strings.EqualFold(Param.ServerAddr,"")==false && strings.EqualFold(v,"")==false{
				err:=glogServer.Write(v)
				if err!=nil{
					_logFileTempChan<-v
				}
			}else if Param.FileStorage{
				_logFileTempChan<-v
			}
		}
	}, func(v interface{}, stack []byte) {
		log.Println(v)
		log.Println(string(stack))

	})

	//日志服务
	thread.NewCoroutine(func() {
		<-_startChan
		logFileName:=getLogFileName(time.Now())
		var logFile *os.File
		logFile,err:= os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_SYNC, os.ModePerm)
		if err!=nil{
			log.Println(err)
		}


		ticker:=time.NewTicker(time.Second)
		//var isLogServerOk =false
		defer ticker.Stop()
		for{
			select {
			case v:=<-ticker.C:
				//fmt.Println(v)
				//_logFileName:=Param.LogFilePath+"_glog_"+v.Format("2006_01_02")+".log"
				_logFileName:=getLogFileName(v)
				if strings.EqualFold(logFileName,_logFileName)==false{
					logFileName=_logFileName
					if logFile!=nil{
						logFile.Close()
					}
					logFile,_= os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_SYNC, os.ModePerm)
				}

			case v:=<-_logServerStatus:

				if v{

					/*bf,err:=ioutil.ReadAll(logFile)
					if err!=nil{
						return
					}
					fd := bytes.NewBuffer(bf)
					logFile.Close()
					for{
						l, err := fd.ReadString('\n')
						if err != nil {
							if err == io.EOF {
								lf, _ := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_SYNC, os.ModePerm)
								lf.Truncate(0)
								lf.Sync()
								lf.Close()
								logFile, _ = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_SYNC, os.ModePerm)
							}
							break
						}
						_logChanQueue<-l
					}*/
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
						//logFile.Sync()
					}
				}


			}
		}
		//logFile.Sync()
	}, func(v interface{}, stack []byte) {
		log.Println(v)
		log.Println(string(stack))

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