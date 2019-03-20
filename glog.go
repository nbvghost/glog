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
	"strings"
	"time"
)

//queue
//日志列对
var _logChanQueue =make(chan string,10000)
var _logFileTempChan =make(chan string,1000)
var _logServerOk =make(chan bool)
var _logServerUrl=""
var _logFileName="glog"



func Trace(v ...interface{}) {

	_, file, line, _ := runtime.Caller(1)

	go func(_file string, _line int,_v []interface{}) {
		//util.Trace(funcName,file,line,ok)
		for _, va := range _v {
			if va != nil {


				b,_:=json.Marshal(map[string]interface{}{"Time":time.Now().Format("2006-01-02 15:04:05"),"File":_file,"Line":_line,"Trace":va})
				//conf.LogQueue=append(conf.LogQueue,string(b))
				_logChanQueue<-string(b)
				//log.Println(file, line, va)


			}
		}
	}(file,line,v)

}

func Error(err error) bool {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)

		go func(_file string, _line int,_err error) {
			//util.Trace(funcName,file,line,ok)



				//log.Println(file, line, err)
				b,_:=json.Marshal(map[string]interface{}{"Time":time.Now().Format("2006-01-02 15:04:05"),"File":_file,"Line":_line,"Error":_err})
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

				if strings.EqualFold(_logServerUrl,""){
					_logFileTempChan<-_v
					return
				}
				client:=&http.Client{}
				client.Timeout=3*time.Second
				//log.Println(_v)
				buffer := bytes.NewBufferString(_v)
				request,err := http.NewRequest("POST", _logServerUrl, buffer)
				//response,err:=http.Post(conf.Config.LogServer, "text/plain", buffer)
				if err!=nil{
					log.Println(err)
					_logServerOk<-false
					_logFileTempChan<-_v

				}else{
					response,err:=client.Do(request)
					if err!=nil{
						log.Println(err)
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
							log.Println(m["Message"].(string))
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

		logFileName:=_logFileName+"_glog_"+time.Now().Format("2006_01_02")+".log"

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
func NewLogger(params struct{
	Url string
	LogFileName string
	Debug bool
})  {
	_logServerUrl =url
}