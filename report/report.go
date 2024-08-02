package report

import (
	"io"
	"unsafe"
)

// Report 定义一个 io.Writer 结构体
type Report struct {
	level string
	cb    func(level string, message string)
}

// Write 定义Write方法以实现 io.Writer
func (r Report) Write(p []byte) (n int, err error) {
	r.cb(r.level, *(*string)(unsafe.Pointer(&p)))
	return 0, nil
}

// IoWrite 定义工厂函数
func IoWrite(level string, cb func(level string, message string)) io.Writer {
	return Report{
		level: level,
		cb:    cb,
	}
}
