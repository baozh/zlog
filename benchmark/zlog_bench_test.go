package bench

import (
	"testing"
	"zlog"
	"time"
)

//go test -bench="." -benchtime=10s -v
//会默认以 当前cpu的核数 设置 最大的并发处理数
func BenchmarkZlogPrintFileName_Parallel(b *testing.B) {
	zlog.SetLogLevel(zlog.DebugLevel)
	zlog.SetWriteTypeFile("./")
	zlog.SetPrintFileNameLineNo(true)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var a int = 1
		var b float64 = 2.0
		var c string = "three"
		var d bool = true
		var e time.Duration = 5 * time.Second
		i := 0
		for pb.Next() {
			i++
			zlog.Debugln("Test logging, int:", a, ", float:", b, ", string:", c, ", bool:", d, ", time.Duration:", e)
		}
	})
}

func BenchmarkZlogPrintFileName_Singal(bb *testing.B) {
	zlog.SetLogLevel(zlog.DebugLevel)
	zlog.SetWriteTypeFile("./")
	zlog.SetPrintFileNameLineNo(true)

	var a int = 1
	var b float64 = 2.0
	var c string = "three"
	var d bool = true
	var e time.Duration = 5 * time.Second
	bb.ResetTimer()
	bb.StartTimer()
	for i := 0; i < bb.N; i++ {
		i++
		zlog.Debugln("Test logging, int:", a, ", float:", b, ", string:", c, ", bool:", d, ", time.Duration:", e)
	}
}

func BenchmarkZlogNonePrintFileName_Parallel(b *testing.B) {
	zlog.SetLogLevel(zlog.DebugLevel)
	zlog.SetWriteTypeFile("./")
	zlog.SetPrintFileNameLineNo(false)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var a int = 1
		var b float64 = 2.0
		var c string = "three"
		var d bool = true
		var e time.Duration = 5 * time.Second
		i := 0
		for pb.Next() {
			i++
			zlog.Debugln("Test logging, int:", a, ", float:", b, ", string:", c, ", bool:", d, ", time.Duration:", e)
		}
	})
}

func BenchmarkZlogNonePrintFileName_Singal(bb *testing.B) {
	zlog.SetLogLevel(zlog.DebugLevel)
	zlog.SetWriteTypeFile("./")
	zlog.SetPrintFileNameLineNo(false)

	var a int = 1
	var b float64 = 2.0
	var c string = "three"
	var d bool = true
	var e time.Duration = 5 * time.Second
	bb.ResetTimer()
	bb.StartTimer()
	for i := 0; i < bb.N; i++ {
		i++
		zlog.Debugln("Test logging, int:", a, ", float:", b, ", string:", c, ", bool:", d, ", time.Duration:", e)
	}
}
