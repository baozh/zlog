package main

import (
	"github.com/baozh/zlog"
	"time"
	"fmt"
)

func main() {
	zlog.SetLogLevel(zlog.DebugLevel)
	zlog.SetWriteTypeConsole()
	zlog.SetPrintFileNameLineNo(true)

	t1 := time.Now()
	var times int = 10
	for i := 0; i < times; i++ {
		zlog.Debugln("this is a debug log, ", i , " for test")
		zlog.Debuglnf("this is a debug log, %d, %s", i , " for test")
		zlog.Infoln("this is a info log, ", i , ", for test")
		zlog.Infolnf("this is a info log, %d, %s", i , " for test")
		zlog.Warnln("this is a warn log, ", i , ", for test")
		zlog.Warnlnf("this is a warn log, %d, %s", i , " for test")
		zlog.Errorln("this is a error log, ", i , ", for test")
		zlog.Errorlnf("this is a error log, %d, %s", i , " for test")
	}

	t2 := time.Now()
	cost := t2.Sub(t1).Seconds()
	var speed float64 = float64(times*8)/float64(cost)

	fmt.Println("The logging process costs", cost , "s to complete, speed:", speed, "tiao/s")

	time.Sleep(1000*time.Second)
}



