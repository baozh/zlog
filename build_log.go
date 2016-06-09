package zlog

import (
	"fmt"
	"time"
)

func (l *Logger) print(level LogLevel, args ...interface{}) {
	msg := l.writer.formatHeader(level, l.isPrintFileNameLineNo)
	fmt.Fprint(msg, args...)
	msg.appendByte('\n')
	l.writeBuf(msg)
	msg.Clear()
	recordPool.Put(msg)
}

func (l *Logger) printf(level LogLevel, format string, args ...interface{}) {
	msg := l.writer.formatHeader(level, l.isPrintFileNameLineNo)
	fmt.Fprintf(msg, format, args...)
	msg.appendByte('\n')
	l.writeBuf(msg)
	msg.Clear()
	recordPool.Put(msg)
}

func (l *Logger) writeBuf(msg *LogMsg) {
	//写入currentBuf
	notHaveEnoughBuffer := false
	l.curBufMutex.Lock()
	if l.currentBuffer != nil {
		//先判断curBuf 是否有足够的空间写入
		if l.currentBuffer.GetAvailLength() > msg.GetLength() {
			l.currentBuffer.AppendByte(msg.GetBytes())
		} else {
			//如果空间不够，则push 到 fullBufs，然后 申请一个emptyfull，写入 日志串，再唤醒 routine(flushFullBuffers).
			l.fullBuffers.PushBuffer(l.currentBuffer)
			l.currentBuffer = l.emptyBuffers.PopBuffer()
			if l.currentBuffer != nil {
				l.currentBuffer.AppendByte(msg.GetBytes())
			} else {
				notHaveEnoughBuffer = true
			}
		}
	} else {
		notHaveEnoughBuffer = true
	}
	l.curBufMutex.Unlock()

	//currentBuffer 为空，说明 没有可用buf了，消费速度 跟不上 生产速度，则丢弃日志
	if notHaveEnoughBuffer == true {
		//丢弃日志的时候，也要打印相关信息
		//这时 等待 可用的 emptybuffer, 启一个routine 来设置 currentBuf
		if l.isWaitingAvailBuffer == false {
			go WaitingAndSetCurrentBuf(l, time.Now())
			l.isWaitingAvailBuffer = true
		}
	}
}

func (l *Logger) wakeup() {
	l.curBufMutex.Lock()
	if (l.currentBuffer != nil && l.currentBuffer.GetLength() > 0) {
		l.fullBuffers.PushBuffer(l.currentBuffer)
		l.currentBuffer = l.emptyBuffers.PopBuffer()
	}
	l.curBufMutex.Unlock()
}

func WaitingAndSetCurrentBuf(l *Logger, startTime time.Time) {
	//等待可用的empty buf，并设置 currentBuf
	l.emptyBuffers.WaitNewBufferUnlimitedTime()

	endTime := time.Now()
	logStr := "Lost log msg, StartTime:" + startTime.Format("2006-01-02 15:04:05.999999") +
				", EndTime:" + endTime.Format("2006-01-02 15:04:05.999999") + "\n"

	//将提示信息（丢弃日志串）, startTime, endTime 写入到curBufer中
	l.curBufMutex.Lock()
	l.currentBuffer = l.emptyBuffers.PopBuffer()
	l.currentBuffer.AppendString(logStr)
	l.curBufMutex.Unlock()

	l.isWaitingAvailBuffer = false
}



const (
	DEFALUT_LOG_SIZE int = 4000
)

type LogMsg struct {
	logContentTmp  [DEFALUT_LOG_SIZE]byte
	writeIndex int

	logContent []byte
	logContentSize int
}

func NewLogMsg() *LogMsg {
	msg := &LogMsg {}
	msg.logContentSize = DEFALUT_LOG_SIZE
	msg.logContent = msg.logContentTmp[:msg.logContentSize]
	return msg
}

func (log *LogMsg) Clear() {
	log.writeIndex = 0
}

func (log *LogMsg) Avail() int {
	return log.logContentSize - log.writeIndex
}

func (log *LogMsg) GetLength() int {
	return log.writeIndex
}

func (log *LogMsg) GetBytes() []byte {
	return log.logContent[:log.writeIndex]
}

func (log *LogMsg) appendInt(value int) {
	n := log.someDigits(log.writeIndex, value)
	log.writeIndex += n
}

func (log *LogMsg) appendByte(value byte) {
	if (log.logContentSize - log.writeIndex > 1) {
		log.logContent[log.writeIndex] = value
		log.writeIndex++
	}
}

func (log *LogMsg) setString(value string) {
	n := copy(log.logContent[0:], value)
	log.writeIndex += n
}

func (log *LogMsg) appendString(value string) {
	n := copy(log.logContent[log.writeIndex:], value)
	log.writeIndex += n
}

func (log *LogMsg) Write(value []byte) (retN int, err error) {
	if (log.logContentSize - log.writeIndex > len(value)) {
		retN := copy(log.logContent[log.writeIndex:], value)
		log.writeIndex += retN
	} else {
		tmp := log.logContent
		log.logContent = make([]byte, 2*(log.writeIndex + len(value)))
		copy(log.logContent[0:], tmp[:log.writeIndex])
		retN := copy(log.logContent[log.writeIndex:], value)
		log.writeIndex += retN
	}
	return retN, nil
}

// Some custom tiny helper functions to print the log header efficiently.
const digits = "0123456789"
//const digits_other = " 123456789"

// twoDigits formats a zero-prefixed two-digit integer at buf.tmp[i].
func (log *LogMsg) twoDigits(index, value int) {
	log.logContent[index+1] = digits[value%10]
	value /= 10
	log.logContent[index] = digits[value%10]
}

func (log *LogMsg) fourDigits(index, value int) {
	log.logContent[index+3] = digits[value%10]
	value /= 10
	log.logContent[index+2] = digits[value%10]
	value /= 10
	log.logContent[index+1] = digits[value%10]
	value /= 10
	log.logContent[index] = digits[value%10]
}

//func (log *LogMsg) sixDigits(index, value int) {
//	log.logContent[index+5] = digits[value%10]
//	value /= 10
//	log.logContent[index+4] = digits[value%10]
//	value /= 10
//	log.logContent[index+3] = digits[value%10]
//	value /= 10
//	log.logContent[index+2] = digits[value%10]
//	value /= 10
//	log.logContent[index+1] = digits[value%10]
//	value /= 10
//	log.logContent[index] = digits[value%10]
//}
//
//func (log *LogMsg) sevenDigits(index, value int) {
//	log.logContent[index+6] = digits_other[value%10]
//	value /= 10
//	log.logContent[index+5] = digits_other[value%10]
//	value /= 10
//	log.logContent[index+4] = digits_other[value%10]
//	value /= 10
//	log.logContent[index+3] = digits_other[value%10]
//	value /= 10
//	log.logContent[index+2] = digits_other[value%10]
//	value /= 10
//	log.logContent[index+1] = digits_other[value%10]
//	value /= 10
//	log.logContent[index] = digits_other[value%10]
//}

// nDigits formats an n-digit integer at buf.tmp[i],
// padding with pad on the left.
// It assumes d >= 0.
func (log *LogMsg) nDigits(len, index, value int, pad byte) {
	remain := len - 1
	for ; remain >= 0 && value > 0; remain-- {
		log.logContent[index+remain] = digits[value%10]
		value /= 10
	}
	for ; remain >= 0; remain-- {
		log.logContent[index+remain] = pad
	}
}

// someDigits formats a zero-prefixed variable-width integer at buf.tmp[i].
func (log *LogMsg) someDigits(index, value int) int {
	// Print into the top, then copy down. We know there's space for at least
	// a 10-digit number.
	beginWriteIndex := len(log.logContent)
	for {
		beginWriteIndex--
		log.logContent[beginWriteIndex] = digits[value%10]
		value /= 10
		if value == 0 {
			break
		}
	}
	return copy(log.logContent[index:], log.logContent[beginWriteIndex:])
}
