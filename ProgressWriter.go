package main

import (
	"fmt"
	"io"
)

// ProgressWriter 用于显示每个文件的进度
type ProgressWriter struct {
	writer         io.Writer
	compressedSize int64
	total          int64
	progress       int64
	fileName       string
	uncompress     bool
}

func (p *ProgressWriter) Write(b []byte) (n int, err error) {
	n, err = p.writer.Write(b)
	if err != nil {
		return n, err
	}
	p.progress += int64(n)

	// 每次写入数据后显示进度
	p.showProgress()
	return n, err
}

func (p *ProgressWriter) showProgress() {
	if p.total == 0 {
		return
	}
	progressPercent := float64(p.progress+p.compressedSize) / float64(p.total) * 100
	if p.uncompress {
		fmt.Printf("\r解压进度: %.2f%%", progressPercent)
	} else {
		fmt.Printf("\r压缩进度: %.2f%%", progressPercent)
	}
	// fmt.Printf("\rCompressing file '%s': %.2f%%", p.fileName, progressPercent)
}
