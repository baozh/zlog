package zlog

import (
	"sync"
	"time"
)

//设置 输出到文件
func SetWriteTypeFile(logFilePath string) error {
	fw, err := NewFileWriter(defaultLogger, logFilePath)
	defaultLogger.writer = fw
	return err
}

//设置 输出到屏幕
func SetWriteTypeConsole() error {
	fw := NewConsoleWriter()
	defaultLogger.writer = fw
	return nil
}

//设置日志级别
func SetLogLevel(level LogLevel) {
	if (defaultLogger != nil) {
		defaultLogger.currentLevel = level
	}
}

//设置 是否 在日志中打印 （文件名，行号，函数名）
//由于 （文件名，行号，函数名）信息是在运行期获取，会影响性能，建议 在测试开发期间 设置打印，在生产环境中 设置不打印.
func SetPrintFileNameLineNo(isAble bool) {
	defaultLogger.isPrintFileNameLineNo = isAble
}

//即时刷出日志到文件中(可在exit前，或者 崩溃前调用)
func FlushAll() {
	defaultLogger.wakeup()
	time.Sleep(1 * time.Second)   //等1秒（等待 刷日志routine 将buffer中的日志写入文件）
	defaultLogger.writer.Flush()
}

//停止 打印
func StopLogging() {
	if (defaultLogger != nil) {
		defaultLogger.isRunning = false
	}
}

func Debugln(args ...interface{}) {
	if defaultLogger.isRunning == true && defaultLogger.currentLevel <= DebugLevel {
		defaultLogger.print(DebugLevel, args...)
	}
}

func Debuglnf(format string, args ...interface{}) {
	if defaultLogger.isRunning == true && defaultLogger.currentLevel <= DebugLevel {
		defaultLogger.printf(DebugLevel, format, args...)
	}
}

func Infoln(args ...interface{}) {
	if defaultLogger.isRunning == true && defaultLogger.currentLevel <= InfoLevel {
		defaultLogger.print(InfoLevel, args...)
	}
}

func Infolnf(format string, args ...interface{}) {
	if defaultLogger.isRunning == true && defaultLogger.currentLevel <= InfoLevel {
		defaultLogger.printf(InfoLevel, format, args...)
	}
}

func Warnln(args ...interface{}) {
	if defaultLogger.isRunning == true && defaultLogger.currentLevel <= WarnLevel {
		defaultLogger.print(WarnLevel, args...)
	}
}

func Warnlnf(format string, args ...interface{}) {
	if defaultLogger.isRunning == true && defaultLogger.currentLevel <= WarnLevel {
		defaultLogger.printf(WarnLevel, format, args...)
	}
}

func Errorln(args ...interface{}) {
	if defaultLogger.isRunning == true && defaultLogger.currentLevel <= ErrorLevel {
		defaultLogger.print(ErrorLevel, args...)
	}
}

func Errorlnf(format string, args ...interface{}) {
	if defaultLogger.isRunning == true && defaultLogger.currentLevel <= ErrorLevel {
		defaultLogger.printf(ErrorLevel, format, args...)
	}
}

func Fatalln(args ...interface{}) {
	if defaultLogger.isRunning == true && defaultLogger.currentLevel <= FatalLevel {
		defaultLogger.print(FatalLevel, args...)
	}
}

func Fatallnf(format string, args ...interface{}) {
	if defaultLogger.isRunning == true && defaultLogger.currentLevel <= FatalLevel {
		defaultLogger.printf(FatalLevel, format, args...)
	}
}

// default
var (
	defaultLogger *Logger = nil
	recordPool  *sync.Pool
)

type Writer interface {
	formatHeader(level LogLevel, isPrintFileName bool) *LogMsg
	Write(content []byte) error
	Flush()
}

type Logger struct {
	writer     		Writer
	currentLevel 	  	LogLevel            //当前日志级别
	currentBuffer  		*LogMsgBuffer
	curBufMutex     	sync.Mutex
	flushInterval   	int                 //刷出日志的间隔 (单位：秒)
	emptyBuffers    	*BufferContainer
	fullBuffers     	*BufferContainer
	isRunning		bool   		    /* atomic */
	isWaitingAvailBuffer  	bool		    /* atomic */
	isPrintFileNameLineNo  	bool
}

func init() {
	defaultLogger = NewLogger()
	recordPool = &sync.Pool{
		New: func() interface{} {
			return NewLogMsg()
		},
	}
}

func NewLogger() *Logger {
	if defaultLogger != nil {
		return defaultLogger
	}

	logger := new(Logger)
	logger.currentLevel = DebugLevel
	logger.writer = NewConsoleWriter()
	logger.flushInterval = 3
	logger.emptyBuffers = NewBufferContainer(DEFAULT_BUFFER_NUM, DEFAULT_BUFFER_NUM, DEFALUT_BUFFER_SIZE)
	logger.fullBuffers = NewBufferContainer(0, DEFAULT_BUFFER_NUM, DEFALUT_BUFFER_SIZE)
	logger.currentBuffer = logger.emptyBuffers.PopBuffer()
	logger.isRunning = true
	logger.isWaitingAvailBuffer = false
	logger.isPrintFileNameLineNo = true
	go flushFullBuffers(logger)

	return logger
}

func flushFullBuffers(logger *Logger) {
	for ;logger.isRunning; {

		logger.fullBuffers.WaitNewBuffer(logger.flushInterval)

		logger.curBufMutex.Lock()
		if (logger.currentBuffer != nil && logger.currentBuffer.GetLength() > 0) {
			logger.fullBuffers.PushBuffer(logger.currentBuffer)
			logger.currentBuffer = logger.emptyBuffers.PopBuffer()
		}
		logger.curBufMutex.Unlock()

		if (logger.fullBuffers.IsEmpty()) {
			continue
		}

		//将fullBuffers中的内容 写入文件中
		tmpBuffers := logger.fullBuffers.GetAllBuffersAndClear()
		for _, buf := range tmpBuffers {
			logger.writer.Write(buf.GetBytes())
			buf.Clear()
		}

		logger.emptyBuffers.PushBuffers(tmpBuffers)
		logger.writer.Flush()
	}

	//flush余下的内容
	logger.writer.Flush()
}

