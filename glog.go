package glog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

//queue
//日志列对
var _logChanQueue =make(chan string,10000)
var _logFileTempChan =make(chan string,1000)
var _logServerOk =make(chan bool)


var Param = &ParamValue{
	ServerUrl:"",
	ServerName:"default",
	LogFileName:"glog",
	Debug:true,
	PrintStack:false,
}



var _glogOut =log.New(os.Stdout,"[TRACE] ",log.LstdFlags|log.LUTC|log.Llongfile)
var _glogErr =log.New(os.Stderr,"[ERROR] ",log.LstdFlags|log.LUTC|log.Llongfile)
var _glogDebug =log.New(os.Stdout,"[DEBUG] ",log.LstdFlags|log.LUTC|log.Llongfile)

func Debug(v ...interface{})  {
	if Param.Debug{
		_glogDebug.Output(2, fmt.Sprintln(v...))
	}

}
func Trace(v ...interface{}) {
	outText:=fmt.Sprintln(v...)

	if strings.EqualFold(outText,""){
		return
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok{
		file = "unknown"
		line = 0
	}
	if Param.Debug{
		 _glogOut.Output(2, outText)
	}
	go func(_file string, _line int,_v string) {
		//NumGoroutine:=runtime.NumGoroutine()
		//GOOS:=runtime.GOOS
		//mem:=&runtime.MemStats{}
		//runtime.ReadMemStats(mem)

		/*memAll:=0
		memFree:=0
		memUsed:=0

		sysInfo := new(syscall.sys)
		err := syscall.Sysinfo(sysInfo)
		if err == nil {
			memAll = sysInfo.Totalram * uint32(syscall.Getpagesize())
			memFree = sysInfo.Freeram * uint32(syscall.Getpagesize())
			memUsed = memAll - memUsed
		}*/


		b,_:=json.Marshal(map[string]interface{}{
			"Time":time.Now().Format("2006-01-02 15:04:05"),
			"File":_file,
			"Line":_line,
			"TRACE":_v,
			"ServerName":Param.ServerName,
		})
		//conf.LogQueue=append(conf.LogQueue,string(b))
		_logChanQueue<-string(b)
		//log.Println(file, line, va)

	}(file,line,outText)

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

		go func(_file string, _line int,_err error) {
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


		}(file,line,err)

		return true
	}else{
		return false
	}
}
func init()  {



	//日志服务
	go func() {

		for v := range _logChanQueue {

			go func(_v string) {

				if strings.EqualFold(Param.ServerUrl,""){
					_logFileTempChan<-_v
					return
				}
				client:=&http.Client{}
				client.Timeout=3*time.Second
				//log.Println(_v)
				buffer := bytes.NewBufferString(_v)
				request,err := http.NewRequest("POST", Param.ServerUrl, buffer)
				//response,err:=http.Post(conf.Config.LogServer, "text/plain", buffer)
				if err!=nil{
					log.Panicln(err)
					_logServerOk<-false
					_logFileTempChan<-_v
				}else{
					response,err:=client.Do(request)
					if err!=nil{
						//log.Panicln(err)
						//log.Println(err)
						_logServerOk<-false
						_logFileTempChan<-_v
						return
					}
					//网络通了

					_logServerOk<-true

					//return &gweb.JsonResult{Data:map[string]interface{}{"Success":false,"Message":err.Error()}}
					b,err:=ioutil.ReadAll(response.Body)
					if err!=nil{
						_logFileTempChan<-_v
						return
					}
					defer response.Body.Close()
					m:=make(map[string]interface{})
					err=json.Unmarshal(b,&m)
					if err!=nil{
						_logFileTempChan<-_v
						return
					}
					if m["Success"]!=nil && m["Message"]!=nil{

						if m["Success"].(bool)==false{
							//log.Println(m["Message"].(string))
							_logFileTempChan<-_v
							return
						}

					}else{
						_logFileTempChan<-_v
						return
					}

				}

			}(v)

		}

	}()

	//日志服务
	go func() {

		logFileName:=Param.LogFileName+"_glog_"+time.Now().Format("2006_01_02")+".log"

		var logFile *os.File
		logFile,_= os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)

		var isLogServerOk =false
		for{
			select {
			case v:=<-time.NewTicker(1*time.Hour).C:
				fmt.Println(v)

				logFile,_= os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)

			case v:=<-_logServerOk:

				if v{
					if isLogServerOk==false{
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
									log.Println(lf.Truncate(0))
									lf.Sync()
									lf.Close()
									logFile, _ = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
								}
								break
							}
							_logChanQueue<-l
						}


					}
					isLogServerOk=true
					//logFile.Sync()
					//logFile.Close()
					//logFile=nil
				}else{
					isLogServerOk=false
				}
			case v :=<-_logFileTempChan:
				//log.Println(v)
				if logFile!=nil{
					logFile.WriteString(fmt.Sprintf("%v\n", v))
					logFile.Sync()
				}

			}
		}
		//logFile.Sync()

	}()

	_logServerOk<-true
}
type ParamValue struct {
	ServerUrl string
	ServerName string
	LogFileName string
	Debug bool
	PrintStack bool
}
func NewLogger(_param *ParamValue){
	if _param!=nil{
		Param = _param
	}

}