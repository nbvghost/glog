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

const DebugLevel Level = 1
const TraceLevel Level = 1 << 1
const ErrorLevel Level = 1 << 2
const PanicLeveL Level = 1 << 3

const MoreDebugLevel Level = DebugLevel | TraceLevel | ErrorLevel | PanicLeveL
const MoreTraceLevel Level = TraceLevel | ErrorLevel | PanicLeveL
const MoreErrorLevel Level = ErrorLevel | PanicLeveL
const MorePanicLeveL Level = PanicLeveL

type Level int

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

var _glogOut = log.New(os.Stdout, "[TRACE] ", log.LstdFlags|log.Lshortfile)
var _glogErr = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile)
var _glogDebug = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lshortfile)

type paramValue struct {
	PushAddr    string
	Tag         string
	AppName     string
	FormatType  FormatType
	LogFilePath string
	StandardOut bool
	FileStorage bool
	ShowHeader  bool
	Level       Level
}

type logger struct {
	param     *paramValue
	skip      int
	calldepth int
}

func (log *logger) Debug(v ...interface{}) {
	if log.param.Level&DebugLevel == DebugLevel {
		_glogDebug.Output(log.calldepth, fmt.Sprintln(v...))
	}
}
func (log *logger) Trace(v ...interface{}) {
	if log.param.Level&TraceLevel == TraceLevel {
		pc, file, line := log.getSource()
		out := log.format(pc, file, line, "TRACE", v)
		if log.param.StandardOut {
			_glogOut.Output(log.calldepth, out)
		}
	}

}

func (log *logger) Error(err error) bool {

	if err != nil {
		if log.param.Level&ErrorLevel == ErrorLevel {
			pc, file, line := log.getSource()
			out := log.format(pc, file, line, "ERROR", []interface{}{map[string]interface{}{
				"ErrorMessage": err.Error(),
				"Stack":        string(debug.Stack()),
			}})
			if log.param.StandardOut {
				_glogErr.Output(log.calldepth, out)
			}
		}
		return true
	} else {
		return false
	}
}

func (log *logger) Panic(err error) {
	if log.param.Level&PanicLeveL == PanicLeveL {
		pc, file, line := log.getSource()
		if err != nil {
			out := log.format(pc, file, line, "PANIC", []interface{}{map[string]interface{}{
				"ErrorMessage": err.Error(),
				"Stack":        string(debug.Stack()),
			}})
			if log.param.StandardOut {
				_glogErr.Output(log.calldepth, out)
			}
			os.Exit(1)
		}
	}

}

func (log *logger) getSource() (uintptr, string, int) {
	pc, file, line, _ := runtime.Caller(log.skip)
	return pc, file, line
}
func (log *logger) getSourceByZero() (uintptr, string, int) {
	pc, file, line, _ := runtime.Caller(0)
	return pc, file, line
}
func (log *logger) format(pc uintptr, file string, line int, level string, values []interface{}) string {

	filePath := getLastTowPath(file)

	_, glogFilePath, _ := log.getSourceByZero()

	version := strings.Split(getLastTowPath(glogFilePath), "/")[0]

	if log.param.FormatType == JSON {

		outs := make(map[string]interface{}, 0)
		outs["File"] = filePath + ":" + strconv.Itoa(line)
		outs["Time"] = time.Now().Format("2006-01-02 15:04:05")
		outs["AppName"] = log.param.AppName
		outs["Tag"] = log.param.Tag
		outs["Level"] = level
		outs["Version"] = version
		outs["PC"] = fmt.Sprintf("%x", pc)

		if len(values) == 1 {
			outs["Logs"] = values[0]
		} else {
			logs := make(map[int]interface{})
			for i := 0; i < len(values); i++ {
				logs[i] = values[i]
			}
			outs["Logs"] = logs
		}

		b, _ := json.Marshal(outs)

		_logChanQueue <- fmt.Sprintln(string(b))

		return fmt.Sprintln(string(b))

	} else {
		outs := make([]interface{}, 0)
		//outs = append(outs, filePath+":"+strconv.Itoa(line))
		outs = append(outs, "PC:"+strconv.Itoa(int(pc)))
		//outs = append(outs, time.Now().Format("2006-01-02 15:04:05"))
		outs = append(outs, log.param.AppName)
		outs = append(outs, log.param.Tag)
		outs = append(outs, version)
		outs = append(outs, values...)

		_logChanQueue <- fmt.Sprintln(outs...)

		return fmt.Sprintln(outs...)
	}

}
func getLastTowPath(file string) string {
	filePath := ""

	filePaths := strings.Split(file, "/")
	if len(filePaths) <= 2 {
		filePath = file
	} else {
		filePath = filePaths[len(filePaths)-2] + "/" + filePaths[len(filePaths)-1]
	}
	return filePath
}
func NewLogger(Tag string) *logger {
	cp := func() paramValue {

		return *Param

	}()
	cp.Tag = Tag
	return &logger{param: &cp, skip: 2, calldepth: 2}
}

var defaultLogger = &logger{param: Param, skip: 3, calldepth: 3}

func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}
func Panic(err error) {
	defaultLogger.Panic(err)
}
func Trace(v ...interface{}) {
	defaultLogger.Trace(v...)
}

func Error(err error) bool {
	return defaultLogger.Error(err)
}

var Param = &paramValue{
	PushAddr:    "",
	Tag:         "",
	AppName:     "default",
	LogFilePath: "glog",
	StandardOut: false,
	FormatType:  CLF,
	FileStorage: false,
	Level:       TraceLevel,
}

func getLogFileName(v time.Time) string {
	logFileName := ""
	if strings.EqualFold(Param.LogFilePath, "") {
		logFileName = Param.AppName + "/" + v.Format("2006_01_02") + ".log"
		err := os.MkdirAll(Param.AppName, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	} else {
		logFileName = Param.LogFilePath + "/" + Param.AppName + "/" + v.Format("2006_01_02") + ".log"
		err := os.MkdirAll(Param.LogFilePath+"/"+Param.AppName, os.ModePerm)
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

var once = &sync.Once{}

func Start() {

	once.Do(func() {

		if Param.FormatType == JSON && Param.ShowHeader == false {
			_glogOut = log.New(os.Stdout, "", 0)
			_glogErr = log.New(os.Stderr, "", 0)
			_glogDebug = log.New(os.Stdout, "", 0)
		}

		initGLog()
	})

}
