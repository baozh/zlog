
# zlog

一个Go语言的轻量级日志库，支持日志的分级输出、日志文件切分，可指定输出到屏幕、文件。

## 特点

- 日志级别: DEBUG, INFO, WARN, ERROR, FATAL，可在运行期修改日志级别
- 可指定输出到文件、屏幕
- 输出屏幕时，对不同级别的日志，用不同的颜色输出，便于观看
- 简单易用，速度快

## 缘起

之前一直用C++编程，项目中用muduo log库，很好用。学习Go的过程中，试用了很多开源的日志库(glog, log4go, seelog, logrus...)，感觉不太满意(要么太复杂，要么性能不达标)，所以花时间自己撸了一个。我的设计原则是 功能简单够用，速度快。

## 使用

	package main

	import (
		"zlog"
	)

	func main() {
		//设置日志级别
		zlog.SetLogLevel(zlog.DebugLevel)
		//设置输出到文件，传入 日志文件的路径
		zlog.SetWriteTypeFile("./")
		//设置不打印(文件名、行号、函数名)
		zlog.SetPrintFileNameLineNo(false)

		var a int = 1
		var b float64 = 2.0
		var c string = "three"
		var d bool = true
		var e time.Duration = 5 * time.Second

		//输出日志
		zlog.Debugln("Test logging, i:", i, ", int:", a, ", float:", b, ", string:", c, ", bool:", d, ", time.Duration:", e)

		//即时刷出日志到文件中(可在exit前，或者 崩溃前调用)
		zlog.FlushAll()
	}

## 设计

功能需求：

- 日志按级别输出，如果打印级别低于当前日志级别，则不打印。要支持在运行期修改日志级别。
- 日志文件的滚动，滚动的条件有两个：文件大小(例如文件大小超过100MB)和时间(例如每天零点新建一个日志文件，便于快速定位日志文件)。
- 支持输出到屏幕(最好按不同的颜色输出)，便于在开发期间调试。

不需要支持的功能：

- 输出到不同的目的地，如socket, SMTP等。
- 可配置日志输出的格式。
- 不同Goroutine，或者，不同日志级别，写不同的文件。

日志的目的地只有一个：本地文件。往网络写日志消息是不靠谱的，因为诊断日志的功能之一就是诊断网络故障，如果网络有问题，会导致日志输出阻塞。  
日志消息的格式是固定的，不需要运行时配置，这样可以节省解析日志格式字符串、组装日志消息的时间。  
所有日志都顺序输出到同一个文件，否则，需要在不同的文件中跳来跳去(查找事件发生的先后)，比较麻烦。  

日志消息格式：  

    日期  	    时间.微秒   	pid  日志级别  源文件名：行号：函数名 -   正文
    20160609 23:31:21.770367   28599 ERROR    demo.go:33:main.main - Hello

每条日志独占一行，时间戳精确到微秒(便于用日志时间来观测性能)，最好打印Goroutine ID、文件名、行号、函数名，便于调试。由于官方不允许获取Goroutine ID，所以只能用pid来代替。

日志文件的命名格式：  

    日期-时间.basename.主机名.pid.log
    20160609-22171.file_demo.bzh-HP-Pavilion-m4-Notebook-PC.27204.log
    注：basename是可执行文件的名字

性能需求：

- 当输出大量日志时，不阻塞正常的执行流程。
- 在多个Goroutine同时输出日志时，不造成争用(contention)。
- 每秒输出日志100万条。

比如一个进程每秒处理2万个请求，每个请求打印5条日志，这时就需要最少有10w/s的性能。但是如果日志库的性能越高，进程就能腾出更多的资源来作正事。

## 实现

高性能的日志库都需要对**磁盘写操作**友好，一般通过**收集日志串，再批量顺序写文件**来提高性能。所以会有多个生产者Goroutine和一个消费者Goroutine。在实现的时候，会思考一些问题：

1. Buffer如何设计？什么时候唤醒日志协程从Buffer中取数据？
2. 如何减少 业务协程、日志协程 访问Buffer时的锁竞争？
3. 日志串如何组装，才能使它组装速度足够快、且要兼顾接口设计的易用性？
4. 要考虑Goroutine间的锁竞争、消费的速度不慢于生产的速度，以免造成日志的堆积。
5. 什么时候切换写到另一个日志文件？什么时候flush到日志文件？
6. 若日志串写入过多，日志线程来不及消费，怎么办？

![buffer](./buffer.jpg)

Buffer之间数据的流转 如图所示，程序启动时，预分配多个buffer存放到`emptyBuffersQueue`中，业务协程在输出日志时，如果当前`curBuffer`为空、或者空间不够，就用`emptyBufferQueue`中取一个buf，写入日志串，再将原来的`curBuffer`存入到`fullBuffersQueue`中。而日志协程，不停地从`fullBuffersQueue`中取出所有的buffer，批量写入到文件中，然后再存入到`emptyBufferQueue`中。  
这么设计的好处：可重复利用Buffer空间，减少分配大块内存的时间。  

**性能优化的tips：**

1. 日志串空间的分配用sync.pool，减少小对象频繁分配的时间。
2. 日志串header的组装、日志文件名的组装，不用库函数fmt.Sprint()，而是自动手动组装，减少开销。
3. 尽量减少 业务协程、日志协程对共享变量的访问，减少锁冲突。

**日志输出过快、来不及消费，怎么办？**  
直接丢弃多余的日志，启一个协程去等待一个可用的Buffer，Buffer可用后 写入 丢弃日志的开始时间 和 结束时间。  

**获取 `源文件名、行号、函数名`信息，性能损耗严重。**  
在C/C++中，可以用__FILE__, __LINE__, __func__ 在编译期获取这些信息，但是当前Go只支持在运行期(从runtime包)获取，很影响性能(大概会影响一两倍的速度)，但是这些信息在调试开发期间对定位代码很有帮助。我做了个折中，提供一个接口`是否输出 文件名，行号，函数名`，在开发环境，可以输出，方便调试。在生产环境，不输出，以免影响正常业务。  

PS: 在编译期获取 `源文件名、行号、函数名`，需要有编译器的支持。Go在2015的时候，有人提议 增加两个类似__FILE__, __LINE__的宏(见这两个issue, [issue1](https://github.com/Sirupsen/logrus/issues/63), [issue2](https://github.com/golang/go/issues/12876))，不过，被人驳斥 这个做法不符合Go的美学价值观，所以到现在一直没提供。  

**Benchmark：**  
在一台   HP笔记本电脑 上的测试：
配置：CPU 8核 Intel(R) Core(TM) i7-3632QM 2.20GHz，内存8G，磁盘(同步)写入带宽70.0MB/s
系统：Ubuntu 14.04

    BenchmarkZlogPrintFileName_Parallel-8    	      10000000	      1158 ns/op
    BenchmarkZlogPrintFileName_Singal-8                10000000	      2144 ns/op
    BenchmarkZlogNonePrintFileName_Parallel-8	  30000000	       566 ns/op
    BenchmarkZlogNonePrintFileName_Singal-8  	  10000000	      1141 ns/op

单Goroutine写日志的速度大约是: 87.6w/s(设置不写文件名、行号), 46.6w/s(设置写文件名、行号)。多Goroutine同时写日志的速度大约是: 176.6w/s(设置不写文件名、行号), 86.3w/s(设置写文件名、行号)。




