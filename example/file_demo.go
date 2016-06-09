package main

import (
	"zlog"
	"time"
	"fmt"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	zlog.SetLogLevel(zlog.DebugLevel)
	zlog.SetWriteTypeFile("./")
	zlog.SetPrintFileNameLineNo(true)

	var a int = 1
	var b float64 = 2.0
	var c string = "three"
	var d bool = true
	var e time.Duration = 5 * time.Second

	t1 := time.Now()
	var times int = 1000000
	for i := 0; i < times; i++ {
		zlog.Debugln("Test logging, i:", i, ", int:", a, ", float:", b, ", string:", c, ", bool:", d, ", time.Duration:", e)
		zlog.Debuglnf("Test logging, i:%d, int:%d, float:%f, string:%s, bool:%d, time.Duration:%v", i, a, b, c, d, e)
		zlog.Infoln("Test logging, i:", i, ", int:", a, ", float:", b, ", string:", c, ", bool:", d, ", time.Duration:", e)
		zlog.Infolnf("Test logging, i:%d, int:%d, float:%f, string:%s, bool:%d, time.Duration:%v", i, a, b, c, d, e)
		zlog.Warnln("Test logging, i:", i, ", int:", a, ", float:", b, ", string:", c, ", bool:", d, ", time.Duration:", e)
		zlog.Warnlnf("Test logging, i:%d, int:%d, float:%f, string:%s, bool:%d, time.Duration:%v", i, a, b, c, d, e)
		zlog.Errorln("Test logging, i:", i, ",int:", a, ", float:", b, ", string:", c, ", bool:", d, ", time.Duration:", e)
		zlog.Errorlnf("Test logging, i:%d, int:%d, float:%f, string:%s, bool:%d, time.Duration:%v", i,a, b, c, d, e)
	}

	t2 := time.Now()
	cost := t2.Sub(t1).Seconds()
	var speed float64 = float64(8*times)/float64(cost)

	fmt.Println("The logging process costs", cost , "s to complete, speed:", speed, "logs/s")

	zlog.FlushAll()

	time.Sleep(1000*time.Second)
}


