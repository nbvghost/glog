package glog

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "runtime"
    "runtime/debug"
    "strings"
    "time"
)

//queue
//日志列对
var _logChanQueue = make(chan string, 1000)
var _logFileTempChan = make(chan string, 1000)
var _logServerStatus = make(chan bool)

//var _closeApp =make(chan bool)

var glogServer = &GlogTCP{}

type ParamValue struct {
    ServerAddr  string
    ServerName  string
    LogFilePath string
    StandardOut bool
    PrintStack  bool
    FileStorage bool
}

var Param = &ParamValue{
    ServerAddr:  "",
    ServerName:  "default",
    LogFilePath: "glog",
    StandardOut:       true,
    PrintStack:  false,
    FileStorage: false,
}

var _glogOut = log.New(os.Stdout, "[TRACE] ", log.LstdFlags|log.Lshortfile)
var _glogErr = log.New(os.Stderr, "[ERROR] ", log.LstdFlags|log.Lshortfile)
var _glogDebug = log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lshortfile)

func Debugf(format string, v ...interface{}) {
    if Param.StandardOut {
        _glogDebug.Output(2, fmt.Sprintf(format, v...))
    }
}
func Debug(v ...interface{}) {
    if Param.StandardOut {
        _glogDebug.Output(2, fmt.Sprintln(v...))
    }
}
func Panic(err error) {
    if Error(err) {
        panic(err)
    }
}
func Print(v ...interface{}) {

    //outText:=fmt.Sprintln(v...)
    _, file, line, ok := runtime.Caller(1)
    if !ok {
        file = "unknown"
        line = 0
    }
    if Param.StandardOut {
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

    filePaths := strings.Split(file, "/")
    if len(filePaths) == 0 {
        filePaths = []string{file}
    }

    outs := make([]interface{}, 0)
    outs = append(outs, filePaths[len(filePaths)-1])
    outs = append(outs, line)
    outs = append(outs, time.Now().Format("2006-01-02 15:04:05"))
    outs = append(outs, Param.ServerName)
    outs = append(outs, v...)

    _logChanQueue <- fmt.Sprintln(outs...)

}
func Trace(v ...interface{}) {

    //outText:=fmt.Sprintln(v...)
    _, file, line, ok := runtime.Caller(1)
    if !ok {
        file = "unknown"
        line = 0
    }
    if Param.StandardOut {
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

    filePaths := strings.Split(file, "/")
    if len(filePaths) == 0 {
        filePaths = []string{file}
    }

    outs := make([]interface{}, 0)
    outs = append(outs, filePaths[len(filePaths)-1])
    outs = append(outs, line)
    outs = append(outs, time.Now().Format("2006-01-02 15:04:05"))
    outs = append(outs, Param.ServerName)
    outs = append(outs, "TRACE")
    outs = append(outs, v...)

    _logChanQueue <- fmt.Sprintln(outs...)

    /*
    	b, _ := json.Marshal(map[string]interface{}{
    		"Time":       time.Now().Format("2006-01-02 15:04:05"),
    		"File":       filePaths[len(filePaths)-1],
    		"Line":       line,
    		"TRACE":      v,
    		"ServerName": Param.ServerName,
    	})
    	_logChanQueue <- string(b)*/

}

func Error(err error) bool {
    if err != nil {
        _, file, line, ok := runtime.Caller(1)
        if !ok {
            file = "unknown"
            line = 0
        }
        if Param.StandardOut {
            _glogErr.Output(2, fmt.Sprintln(err))

        }
        if Param.PrintStack {
            debug.PrintStack()
        }

        outs := make([]interface{}, 0)
        outs = append(outs, file)
        outs = append(outs, line)
        outs = append(outs, time.Now().Format("2006-01-02 15:04:05"))
        outs = append(outs, Param.ServerName)
        outs = append(outs, "ERROR")
        outs = append(outs, err.Error())

        _logChanQueue <- fmt.Sprintln(outs...)

        //log.Println(file, line, err)
        /*b, _ := json.Marshal(map[string]interface{}{
        	"Time":       time.Now().Format("2006-01-02 15:04:05"),
        	"File":       file,
        	"Line":       line,
        	"ERROR":      err.Error(),
        	"ServerName": Param.ServerName,
        })*/
        //conf.LogQueue=append(conf.LogQueue,string(b))
        //_logChanQueue <- string(b)

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
    } else {
        return false
    }
}
func getLogFileName(v time.Time) string {
    logFileName := ""
    if strings.EqualFold(Param.LogFilePath, "") {
        logFileName = Param.ServerName + "/" + v.Format("2006_01_02") + ".log"
        err := os.MkdirAll(Param.ServerName, os.ModePerm)
        if err != nil {
            log.Println(err)
        }
    } else {
        logFileName = Param.LogFilePath + "/" + Param.ServerName + "/" + v.Format("2006_01_02") + ".log"
        err := os.MkdirAll(Param.LogFilePath+"/"+Param.ServerName, os.ModePerm)
        if err != nil {
            log.Println(err)
        }
    }
    return logFileName
}

func init() {

    go func() {

        if strings.EqualFold(Param.ServerAddr, "") == false {
            glogServer.StartTCP(Param.ServerAddr, _logServerStatus)
        }

    }()
    //日志服务
    go func() {
        for v := range _logChanQueue {
            if strings.EqualFold(Param.ServerAddr, "") == false && strings.EqualFold(v, "") == false {
                err := glogServer.Write(v)
                if err != nil && Param.FileStorage {
                    _logFileTempChan <- v
                }
            } else if Param.FileStorage {
                _logFileTempChan <- v
            }
        }
    }()

    //日志写入服务
    go func() {

        if Param.FileStorage == false {
            return
        }

        logFileName := getLogFileName(time.Now())
        var _logFile *os.File
        var logFileWriter *bufio.Writer
        var err error
        _logFile, err = os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND|os.O_SYNC, os.ModePerm)
        if err != nil {
            log.Println(err)
        } else {
            logFileWriter = bufio.NewWriter(_logFile)
        }

        ticker := time.NewTicker(time.Second)
        //var isLogServerOk =false
        //sn:=thread.ListeningSignal()
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

}

func StartLogger(_param *ParamValue) {
    if _param != nil {
        Param = _param
    }

}
