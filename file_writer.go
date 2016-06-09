package zlog

import (
	"bufio"
	"os"
	"errors"
	"path/filepath"
	"time"
	"runtime"
	"strings"
)

type FileWriter struct {
	logger 		*Logger
	bufWriter 	*bufio.Writer
	file   		*os.File
	nbytes 		uint64    //当前已写入的字节数
	startOfPeriod int64
	count 		int
	logFilePath string
}


const rollSize uint64 = 100*1024*1024
const bufferSize = 256 * 1024
const rollPerSeconds = 60 * 60 *24
const checkRotateEveryN = 102400


func NewFileWriter(logger *Logger, filePath string) (*FileWriter, error) {
	fw := &FileWriter{}
	fw.logger = logger
	fw.logFilePath = filePath

	//判断路径是否存在，如果不存在，则创建
	if err := os.MkdirAll(filePath, 0755); err != nil {
		if !os.IsExist(err) {
			return nil, errors.New("filePath:" + filePath + " create failed!")
		}
	}

	fw.Rotate()
	return fw, nil
}

func (fw *FileWriter) formatHeader(level LogLevel, isPrintFileName bool) *LogMsg {
	msg := recordPool.Get().(*LogMsg)

	// 手动组装日志串，而不是用Sprintf，因为Sprintf比较耗时.
	//组装日志串的格式：
	//日期    时间.微秒    pid   日志级别  源文件名：行号：函数名 -   正文
	now := time.Now()
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	msg.fourDigits(0, year)
	msg.twoDigits(4, int(month))
	msg.twoDigits(6, day)
	msg.logContent[8] = ' '
	msg.twoDigits(9, hour)
	msg.logContent[11] = ':'
	msg.twoDigits(12, minute)
	msg.logContent[14] = ':'
	msg.twoDigits(15, second)
	msg.logContent[17] = '.'
	msg.nDigits(6, 18, now.Nanosecond()/1000, '0')
	msg.logContent[24] = ' '
	msg.nDigits(7, 25, pid, ' ')
	msg.logContent[32] = ' '
	msg.writeIndex = 33
	msg.appendString(LEVEL_FLAGS[level])

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
		msg.appendString(" "+ file + ":")
		msg.appendInt(line)
		msg.appendString(":" + funcName + " - ")
	} else {
		msg.appendString(" - ")
	}
	return msg
}

func (fw *FileWriter) Rotate() error {
	if fw.file != nil {
		fw.Flush()
		fw.file.Close()
	}

	//不用fmt.Sprintf，手动组装文件名
	//文件名格式：日期-时间.basename.主机名.pid.log
	fileName := recordPool.Get().(*LogMsg)
	defer func() {
			fileName.Clear()
			recordPool.Put(fileName)
		}()

	now := time.Now()
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	fileName.fourDigits(0, year)
	fileName.twoDigits(4, int(month))
	fileName.twoDigits(6, day)
	fileName.logContent[8] = '-'
	fileName.twoDigits(9, hour)
	fileName.twoDigits(11, minute)
	fileName.twoDigits(13, second)
	fileName.logContent[14] = '.'
	fileName.writeIndex = 15
	fileName.appendString(baseName + "." + hostName + ".")
	fileName.appendInt(pid)
	fileName.appendString(".log")

	tmpfilepath := filepath.Join(fw.logFilePath, string(fileName.GetBytes()))

	if file, err := os.OpenFile(tmpfilepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); err == nil {
		fw.file = file
	} else {
		return err
	}

	fw.nbytes = 0
	fw.startOfPeriod = now.Unix() / rollPerSeconds * rollPerSeconds
	fw.count = 0

	if fw.bufWriter = bufio.NewWriterSize(fw.file, bufferSize); fw.bufWriter == nil {
		return errors.New("new fileBufWriter failed.")
	}

	return nil
}

func (fw *FileWriter) Write(content []byte) error {
	n, _ := fw.bufWriter.Write(content)
	fw.nbytes += uint64(n)

	if fw.nbytes >= rollSize {
		fw.Rotate()
	} else {
		fw.count++
		if (fw.count > checkRotateEveryN) {
			fw.count = 0
			thisPeriod := time.Now().Unix() / rollPerSeconds * rollPerSeconds
			if thisPeriod != fw.startOfPeriod {
				fw.Rotate()
			}
		}
	}
	return nil
}

func (fw *FileWriter) Flush() {
	fw.bufWriter.Flush()
}
