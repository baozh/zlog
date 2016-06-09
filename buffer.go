package zlog

import (
	"sync"
	"time"
)

const (
	DEFALUT_BUFFER_SIZE int = 20000*1024
	DEFAULT_BUFFER_NUM  int = 20
)

type LogMsgBuffer struct {
	buffer []byte
	startWriteIndex int   //next write index
	capacity int
}

func NewLogMsgBuffer(bufferSize int) *LogMsgBuffer {
	if bufferSize == 0 {
		return nil
	}

	return &LogMsgBuffer{
		buffer: make([]byte, bufferSize),
		startWriteIndex: 0,
		capacity: bufferSize,
	}
}

func (buf *LogMsgBuffer) AppendString(logStr string) {
	n := copy(buf.buffer[buf.startWriteIndex:], logStr)
	buf.startWriteIndex += n;
}

func (buf *LogMsgBuffer) AppendByte(logByte []byte) {
	n := copy(buf.buffer[buf.startWriteIndex:], logByte)
	buf.startWriteIndex += n;
}

func (buf *LogMsgBuffer) GetLength() int {
	return buf.startWriteIndex
}

func (buf *LogMsgBuffer) GetBytes() []byte {
	return buf.buffer[:buf.startWriteIndex]
}


func (buf *LogMsgBuffer) GetAvailLength() int {
	return buf.capacity - buf.startWriteIndex
}

func (buf *LogMsgBuffer) Clear() {
	if buf.capacity > 0 {
		if (cap(buf.buffer) == buf.capacity) {
			buf.startWriteIndex = 0  //直接复用 之前分配的空间.
		} else {
			buf.buffer = make([]byte, 0, buf.capacity)
			buf.startWriteIndex = 0
		}
	}
}

type BufferContainer struct {
	buffers 		[]*LogMsgBuffer
	bufferCap 		int       //每个buffer的容量
	buffersCap 		int
	mutex 			sync.Mutex
	cufBufCond		*TimeoutCond
}

func NewBufferContainer(bufsSize int, bufsCap int, capPerBuf int) *BufferContainer {
	if capPerBuf == 0 {
		return nil
	}

	buffers := &BufferContainer{}
	buffers.bufferCap = capPerBuf
	buffers.buffersCap = bufsCap
	buffers.buffers = make([]*LogMsgBuffer, bufsSize, bufsCap)
	buffers.cufBufCond = NewTimeoutCond(&buffers.mutex)
	for i:= 0; i < bufsSize; i++ {
		buffers.buffers[i] = NewLogMsgBuffer(capPerBuf)
		if buffers.buffers[i] == nil {
			return nil
		}
	}
	return buffers
}

func (bufs *BufferContainer) PopBuffer() *LogMsgBuffer {
	bufs.mutex.Lock()
	defer bufs.mutex.Unlock()

	if len(bufs.buffers) > 0 {
		buffer := bufs.buffers[0]
		bufs.buffers = bufs.buffers[1:]    //slice 在删除某个元素时，会 自动释放 被删除元素的内存空间吗？
		return buffer
	} else {
		return nil
	}
}

func (bufs *BufferContainer) GetAllBuffersAndClear() []*LogMsgBuffer {
	bufs.mutex.Lock()
	defer bufs.mutex.Unlock()

	retBufs := bufs.buffers
	bufs.buffers = make([]*LogMsgBuffer, 0, bufs.buffersCap)
	return retBufs
}

func (bufs *BufferContainer) PushBuffer(buf *LogMsgBuffer) {
	bufs.mutex.Lock();
	defer bufs.mutex.Unlock();

	bufs.buffers = append(bufs.buffers, buf)

	if bufs.cufBufCond.HasWaiters() == true {
		bufs.cufBufCond.Signal()
	}
}

func (bufs *BufferContainer) PushBuffers(tmpBufs []*LogMsgBuffer) {
	bufs.mutex.Lock();
	defer bufs.mutex.Unlock();

	bufs.buffers = append(bufs.buffers, tmpBufs...)

	if bufs.cufBufCond.HasWaiters() == true {
		bufs.cufBufCond.Signal()
	}
}

func (bufs *BufferContainer) WaitNewBuffer(sec int) {
	bufs.mutex.Lock()
	if len(bufs.buffers) == 0 {  	//unsual usage!
		bufs.cufBufCond.WaitWithTimeout(time.Duration(sec) * time.Second)
	}
	bufs.mutex.Unlock()
}

func (bufs *BufferContainer) WaitNewBufferUnlimitedTime() {
	bufs.mutex.Lock()
	for ; len(bufs.buffers) == 0; {
		bufs.cufBufCond.Wait()
	}
	bufs.mutex.Unlock()
}

func (bufs *BufferContainer) IsEmpty() bool {
	bufs.mutex.Lock()
	defer bufs.mutex.Unlock()
	return len(bufs.buffers) == 0
}