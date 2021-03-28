//*********************************************************************************
//@Auth:蔡君君
//@Date:2021/03/19 22:30
//@File:log.go
//@Pack:logs
//@Proj:Finger
//@Ides:GoLand
//@Desc:自定义日志模块
//*********************************************************************************
package logs

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
	"runtime/debug"
)

const (
	OutFile = 1 << iota
	OutConsole
)

type Level int

const (
	OFF Level = iota
	FATAL
	ERROR
	WARNING
	INFO
	DEBUG
	ALL
)

var (
	levelFlags = []string{"", "FATAL", "ERROR", "WARNS", "INFOS", "DEBUG", ""}
	tmFormat   = "20060102"
	logDir     = "logs"
	logName    = "noah"
	logExt     = "log"
	logSize    = 20 * 1024 * 1024
	logLevel   = ALL
)

//Auth:2021/03/19 22:31 周五 蔡君君
//Desc:日志操作结构
type logs struct {
	mu   sync.Mutex
	t    *time.Time
	f    *os.File
	l    *log.Logger
	c    bool
	size int
	flag int
}

var l logs

//Auth:2021/03/19 22:32:43 周五 蔡君君
//Desc:初始化日志模块
func InitLog(fileDir, fileName, fileExt string) error {
	logDir = fileDir
	logName = fileName
	logExt = fileExt

	t, err := time.Parse(tmFormat, time.Now().Format(tmFormat))
	if err != nil {
		fmt.Println("create log handler failed,err:", err)
		return err
	}

	l = logs{t: &t, size: 0, flag: OutFile | OutConsole, c: false}
	logLevel = ALL
	if l.flag&OutFile == 0 {
		return nil
	}

	//l.mu.Lock()
	//l.splitLogDir()
	//l.mu.Unlock()

	return nil
}

//Auth:2021/03/19 22:34:08 周五 蔡君君
//Desc:设置日志等级
func SetLogLevel(level int) {
	logLevel = Level(level)
}

//Auth:2021/03/19 22:34:31 周五 蔡君君
//Desc:设置日志输出模式 输出到文件:1, 输出到控制台:2, 默认同时输出到文件和控制台
func SetOutput(out int) {
	l.flag = out
}

//Auth:2021/03/19 22:35:20 周五 蔡君君
//Desc:设置单个日志文件大小,default:20 * 1024 * 1024
func SetLogSize(size int) {
	logSize = size
}

//Auth:2021/03/19 22:36:21 周五 蔡君君
//Desc:判断日志文件是否需要分割
func (l *logs) needSplit() bool {
	if l.size > logSize*1024*1024 { //按照大小分割
		return true
	}

	//按照日期分割
	t, err := time.Parse(tmFormat, time.Now().Format(tmFormat))
	if err != nil {
		return false
	}
	if t.After(*l.t) {
		return true
	}
	return false
}

//Auth:2021/03/19 22:37:17 周五 蔡君君
//Desc:检测日志文件，输出到文件才需要检测分割
func (l *logs) fileCheck() {
	if l.flag&OutFile != 0 && (l.needSplit() || !l.c) {
		l.mu.Lock()
		l.splitLogDir()
		l.mu.Unlock()
	}
}

//Auth:2021/03/19 22:38:19 周五 蔡君君
//Desc:分割日志文件
func (l *logs) splitLogDir() {
	now := time.Now()
	dir := fmt.Sprintf("%s/%s", logDir, now.Format(tmFormat))
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("mkdir failed,err:", err)
		return
	}

	for i := 0; i < 99; i++ {
		fullName := fmt.Sprintf("%s/%s.%s_%d", dir, logName, logExt, i)
		_, err = os.Stat(fullName)
		if err != nil && os.IsNotExist(err) {
			fd, err := os.Create(fullName)
			if err != nil {
				fmt.Println("create file handler, failed,err:", err)
				return
			}
			t, _ := time.Parse(tmFormat, time.Now().Format(tmFormat))
			if l.f != nil {
				_ = l.f.Close()
			}
			l.t = &t
			l.f = fd
			l.size = 0
			l.l = log.New(l.f, "", 0)
			l.c = true
			break
		}
	}

}

//Auth:2021/03/19 22:38:46 周五 蔡君君
//Desc:获取日志记录的前缀，包括等级、时间、调用文件、行号以及函数
func getPrefix(level Level) string {
	now := time.Now()
	prefix := fmt.Sprintf("<%s>[%s-%03.3f]", levelFlags[level], now.Format("20060102-15:04:05"), float32(now.Nanosecond())/1000000000)

	pc, file, line, ok := runtime.Caller(2)
	if ok {
		name := runtime.FuncForPC(pc).Name()
		m,e := debug.ReadBuildInfo()
		file = path.Base(file)
		if e {
			t := strings.TrimPrefix(name, m.Path+"/")
			p := strings.Split(t, ".")[0]
			if t == name{
				return fmt.Sprintf("%smain/%s(%d)", prefix, file, line)
			}
			return fmt.Sprintf("%s%s/%s(%d)", prefix, p, file, line)
		}
		sp :=strings.Split("xxx/" + name, "/")
		return fmt.Sprintf("%s%s/%s(%d)", prefix,strings.Split(sp[len(sp)-1], ".")[0] ,file, line)
	}
	return prefix + ": "
}

//Auth:2021/03/19 22:39:54 周五 蔡君君
//Desc:自动格式化日志记录
func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat("%+v ", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat("%+v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}

//Auth:2021/03/19 22:40:24 周五 蔡君君
//Desc:日志写入操作
func (l *logs) base(prefix, log string) {
	str := fmt.Sprintf("%s%s", prefix, log)
	if l.flag&OutFile != 0 {
		l.size += len(str)
		l.l.Println(str)
	}
	if l.flag&OutConsole != 0 {
		fmt.Println(str)
	}
}

//Auth:2021/03/19 22:40:54 周五 蔡君君
//Desc:DEBUG日志封装
func Debug(f string, v ...interface{}) {
	if DEBUG > logLevel {
		return
	}
	l.fileCheck()
	msg := formatLog(f, v...)
	l.base(getPrefix(DEBUG), msg)
}

//Auth:2021/03/19 22:41:28 周五 蔡君君
//Desc:INFOS日志封装
func Infos(f interface{}, v ...interface{}) {
	if INFO > logLevel {
		return
	}
	l.fileCheck()
	msg := formatLog(f, v...)
	l.base(getPrefix(INFO), msg)
}

//Auth:2021/03/19 22:41:58 周五 蔡君君
//Desc:WARNS日志封装
func Warns(f interface{}, v ...interface{}) {
	if WARNING > logLevel {
		return
	}
	l.fileCheck()
	msg := formatLog(f, v...)
	l.base(getPrefix(WARNING), msg)
}

//Auth:2021/03/19 22:42:15 周五 蔡君君
//Desc:ERROR日志封装
func Error(f interface{}, v ...interface{}) {
	if ERROR > logLevel {
		return
	}
	l.fileCheck()
	msg := formatLog(f, v...)
	l.base(getPrefix(ERROR), msg)
}

//Auth:2021/03/19 22:42:38 周五 蔡君君
//Desc:FATAL日志封装
func Fatal(f interface{}, v ...interface{}) {
	if FATAL > logLevel {
		return
	}
	l.fileCheck()
	msg := formatLog(f, v...)
	l.base(getPrefix(FATAL), msg)
}
