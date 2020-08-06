package glog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FormatType string

const CLF FormatType = "CLF"
const JSON FormatType = "JSON"

//queue
//日志列对
var _logChanQueue = make(chan string, 1000)
var _logFileTempChan = make(chan string, 1000)
var _logServerStatus = make(chan bool)

//var _closeApp =make(chan bool)

var glogServer = &GlogTCP{}

type paramValue struct {
	PushAddr    string
	Name        string
	FormatType  FormatType
	LogFilePath string
	StandardOut bool
	FileStorage bool
}

var Param = &paramValue{
	PushAddr:    "",
	Name:        "default",
	LogFilePath: "glog",
	StandardOut: true,
	FormatType:  CLF,
	FileStorage: false,
}

var _glogOut = log.New(os.Stdout, "[TRACE] ", log.LstdFlags|log.Lshortfile)
var _glogErr = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile)
var _glogDebug = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lshortfile)

func Debug(v ...interface{}) {
	if Param.StandardOut {
		_glogDebug.Output(2, fmt.Sprintln(v...))
	}
}
func getSource(pc uintptr, file string, line int, ok bool) (uintptr, string, int) {

	return pc, file, line
}
func Panic(err error) {
	pc, file, line := getSource(runtime.Caller(1))
	if err != nil {
		format(pc, file, line, "PANIC", []interface{}{err.Error()})
		panic(err)
	}
}
func Trace(v ...interface{}) {
	pc, file, line := getSource(runtime.Caller(1))

	if Param.StandardOut {
		_glogOut.Output(2, fmt.Sprintln(v...))
	}

	format(pc, file, line, "TRACE", v)

}

func Error(err error) bool {
	if err != nil {
		pc, file, line := getSource(runtime.Caller(1))

		if Param.StandardOut {
			_glogErr.Output(2, fmt.Sprintln(err)+string(debug.Stack()))

		}

		format(pc, file, line, "ERROR", []interface{}{err.Error()})

		return true
	} else {
		return false
	}
}

func format(pc uintptr, file string, line int, level string, values []interface{}) {

	filePath := ""

	filePaths := strings.Split(file, "/")
	if len(filePaths) <= 2 {
		filePath = file
	} else {
		filePath = filePaths[len(filePaths)-2] + "/" + filePaths[len(filePaths)-1]
	}

	logs := make(map[int]interface{})
	for i := 0; i < len(values); i++ {
		logs[i] = values[i]
	}

	if Param.FormatType == JSON {

		outs := make(map[string]interface{}, 0)
		outs["File"] = filePath + ":" + strconv.Itoa(line)
		outs["Time"] = time.Now().Format("2006-01-02 15:04:05")
		outs["Name"] = Param.Name
		outs["Level"] = level
		outs["PC"] = fmt.Sprintf("%x", pc)

		outs["Logs"] = logs

		b, _ := json.Marshal(outs)

		_logChanQueue <- fmt.Sprintln(string(b))

	} else {
		outs := make([]interface{}, 0)
		outs = append(outs, filePath+":"+strconv.Itoa(line))
		outs = append(outs, line)
		outs = append(outs, "pc:"+strconv.Itoa(int(pc)))
		outs = append(outs, time.Now().Format("2006-01-02 15:04:05"))
		outs = append(outs, Param.Name)

		b, _ := json.Marshal(logs)

		outs = append(outs, string(b))

		_logChanQueue <- fmt.Sprintln(outs...)
	}

}

func getLogFileName(v time.Time) string {
	logFileName := ""
	if strings.EqualFold(Param.LogFilePath, "") {
		logFileName = Param.Name + "/" + v.Format("2006_01_02") + ".log"
		err := os.MkdirAll(Param.Name, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	} else {
		logFileName = Param.LogFilePath + "/" + Param.Name + "/" + v.Format("2006_01_02") + ".log"
		err := os.MkdirAll(Param.LogFilePath+"/"+Param.Name, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	return logFileName
}

var logFileWriter *bufio.Writer

func initGLog() {
	wait := &sync.WaitGroup{}
	wait.Add(1)
	go func() {
		wait.Done()

		if strings.EqualFold(Param.PushAddr, "") == false {
			glogServer.StartTCP(Param.PushAddr, _logServerStatus)
		}

	}()
	wait.Add(1)
	//日志服务
	go func() {
		wait.Done()
		for v := range _logChanQueue {
			if strings.EqualFold(Param.PushAddr, "") == false && strings.EqualFold(v, "") == false {
				err := glogServer.Write(v)
				if err != nil && Param.FileStorage {
					_logFileTempChan <- v
				}
			} else if Param.FileStorage {
				_logFileTempChan <- v
			}
		}
	}()

	wait.Add(1)
	//日志写入服务
	go func() {
		wait.Done()

		if Param.FileStorage == false {
			return
		}

		logFileName := getLogFileName(time.Now())
		var _logFile *os.File

		var err error
		_logFile, err = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_SYNC, os.ModePerm)
		if err != nil {
			log.Println(err)
		} else {
			logFileWriter = bufio.NewWriter(_logFile)
		}

		ticker := time.NewTicker(time.Second)

		defer ticker.Stop()

		for {
			select {
			case v := <-ticker.C:
				//fmt.Println(v)
				//_logFileName:=Param.LogFilePath+"_glog_"+v.Format("2006_01_02")+".log"
				_logFileName := getLogFileName(v)
				if strings.EqualFold(logFileName, _logFileName) == false {
					logFileName = _logFileName
					if _logFile != nil {
						_logFile.Close()
					}
					if logFileWriter != nil {
						logFileWriter.Flush()
					}
					_logFile, err = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_SYNC, os.ModePerm)
					if err != nil {
						log.Println(err)
					} else {
						logFileWriter = bufio.NewWriter(_logFile)

					}
				}

				if logFileWriter != nil {
					logFileWriter.Flush()
				}

			case v := <-_logServerStatus:

				if v {

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
				} else {
					//isLogServerOk=false
				}
			case v := <-_logFileTempChan:
				//log.Println(v)
				if Param.FileStorage {
					//ioutil.WriteFile(logFileName,[]byte(fmt.Sprintln( v)),os.ModeAppend)
					if logFileWriter != nil {
						n, err := logFileWriter.WriteString(v)
						if err != nil {
							_glogErr.Println(n, err)
						}
						//logFile.Sync()
					}
				}

			}
		}
	}()
	wait.Wait()
}
func Stop() {
	if logFileWriter != nil {
		logFileWriter.Flush()
	}
}
func Start() {

	initGLog()

}
