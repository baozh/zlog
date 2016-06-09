package zlog

import (
	"fmt"
	"os"
	"time"
	"runtime"
	"strings"
)

type ConsoleWriter struct {
}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{}
}

func (fw *ConsoleWriter) formatHeader(level LogLevel, isPrintFileName bool) *LogMsg {
	//组装日志串的格式：
	//日期    时间.微秒    pid   日志级别  源文件名：行号：函数名 -   正文
	now := time.Now()
	year, month, day := now.Date()
	hour, minute, second := now.Clock()

	msg := recordPool.Get().(*LogMsg)
	var header string

	if isPrintFileName {
		//获取源文件名，行号，函数名
		var funcName string
		pc, file, line, ok := runtime.Caller(3)
		if ok {
			slash := strings.LastIndex(file, "/")
			if slash >= 0 {
				file = file[slash+1:]
			}

			funcPtr := runtime.FuncForPC(pc)
			if funcPtr != nil {
				funcName = funcPtr.Name()
			}
		} else {
			file = "NoneFileName"
			line = 1
			funcName = "NoneFuncName"
		}

		//不同的日志级别，用不同的颜色输出
		switch level {
		case DebugLevel:
			header = fmt.Sprintf("%04d%02d%02d %02d:%02d:%02d.%.06d %07d \033[34m%s\033[0m %s:%d:%s - ",
				year, month, day, hour, minute, second, now.Nanosecond()/1000, pid, LEVEL_FLAGS[level], file, line, funcName)
		case InfoLevel:
			header = fmt.Sprintf("%04d%02d%02d %02d:%02d:%02d.%.06d %07d \033[32m%s\033[0m %s:%d:%s - ",
				year, month, day, hour, minute, second, now.Nanosecond()/1000, pid, LEVEL_FLAGS[level], file, line, funcName)
		case WarnLevel:
			header = fmt.Sprintf("%04d%02d%02d %02d:%02d:%02d.%.06d %07d \033[33m%s\033[0m %s:%d:%s - ",
				year, month, day, hour, minute, second, now.Nanosecond()/1000, pid, LEVEL_FLAGS[level], file, line, funcName)
		case ErrorLevel:
			header = fmt.Sprintf("%04d%02d%02d %02d:%02d:%02d.%.06d %07d \033[31m%s\033[0m %s:%d:%s - ",
				year, month, day, hour, minute, second, now.Nanosecond()/1000, pid, LEVEL_FLAGS[level], file, line, funcName)
		case FatalLevel:
			header = fmt.Sprintf("%04d%02d%02d %02d:%02d:%02d.%.06d %07d \033[35m%s\033[0m %s:%d:%s - ",
				year, month, day, hour, minute, second, now.Nanosecond()/1000, pid, LEVEL_FLAGS[level], file, line, funcName)
		}
	} else {
		//不同的日志级别，用不同的颜色输出
		switch level {
		case DebugLevel:
			header = fmt.Sprintf("%04d%02d%02d %02d:%02d:%02d.%.06d %07d \033[34m%s\033[0m - ",
				year, month, day, hour, minute, second, now.Nanosecond() / 1000, pid, LEVEL_FLAGS[level])
		case InfoLevel:
			header = fmt.Sprintf("%04d%02d%02d %02d:%02d:%02d.%.06d %07d \033[32m%s\033[0m - ",
				year, month, day, hour, minute, second, now.Nanosecond() / 1000, pid, LEVEL_FLAGS[level])
		case WarnLevel:
			header = fmt.Sprintf("%04d%02d%02d %02d:%02d:%02d.%.06d %07d \033[33m%s\033[0m - ",
				year, month, day, hour, minute, second, now.Nanosecond() / 1000, pid, LEVEL_FLAGS[level])
		case ErrorLevel:
			header = fmt.Sprintf("%04d%02d%02d %02d:%02d:%02d.%.06d %07d \033[31m%s\033[0m - ",
				year, month, day, hour, minute, second, now.Nanosecond() / 1000, pid, LEVEL_FLAGS[level])
		case FatalLevel:
			header = fmt.Sprintf("%04d%02d%02d %02d:%02d:%02d.%.06d %07d \033[35m%s\033[0m - ",
				year, month, day, hour, minute, second, now.Nanosecond() / 1000, pid, LEVEL_FLAGS[level])
		}
	}

	msg.setString(header)
	return msg
}

func (fw *ConsoleWriter) Write(content []byte) error {
	fmt.Fprint(os.Stdout, string(content))
	return nil
}

func (fw *ConsoleWriter) Flush() {
}
