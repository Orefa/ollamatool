package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// 压缩文件
func zipFiles(files []string, totalSize int64, srcDir, zipFileName string) error {
	startTime := time.Now()

	// 创建目标zip文件
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// 创建zip writer
	zw := zip.NewWriter(zipFile)
	defer zw.Close()

	// 压缩文件并显示进度
	var processedSize int64
	for _, file := range files {
		fileToZip, err := os.Open(file)
		if err != nil {
			return err
		}
		defer fileToZip.Close()

		// 获取相对路径
		relPath, err := filepath.Rel(srcDir, file)
		if err != nil {
			return err
		}

		// 创建zip中的文件头
		zipHeader, err := zw.Create(relPath)
		if err != nil {
			return err
		}

		// 创建 ProgressWriter，包装 xzWriter
		progressWriter := &ProgressWriter{
			writer:         zipHeader,
			compressedSize: processedSize,
			total:          totalSize,
			fileName:       relPath, // 将相对路径作为文件名
		}

		// 复制文件内容
		fileInfo, _ := fileToZip.Stat()
		_, err = io.Copy(progressWriter, fileToZip)
		if err != nil {
			return err
		}

		processedSize += fileInfo.Size()
		// progress := float64(processedSize) / float64(totalSize) * 100
		// fmt.Printf("\r压缩进度: %.2f%%", progress)
	}

	duration := time.Since(startTime)
	fmt.Printf("\n压缩完成！用时: %v, 压缩速度: %.2f MB/s\n",
		duration, float64(totalSize)/1024/1024/duration.Seconds())
	return nil
}

// 解压文件
func unzipFile(zipFileName string, destDir string) error {
	startTime := time.Now()

	// 打开zip文件
	r, err := zip.OpenReader(zipFileName)
	if err != nil {
		return err
	}
	defer r.Close()

	// 计算总大小
	var totalSize int64
	for _, f := range r.File {
		totalSize += int64(f.UncompressedSize64)
	}

	// 创建目标目录
	os.MkdirAll(destDir, 0755)

	// 解压文件并显示进度
	var processedSize int64
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		path := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), 0755)
			outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				rc.Close()
				return err
			}
			// 创建 ProgressWriter，包装 xzWriter
			progressWriter := &ProgressWriter{
				writer:         outFile,
				compressedSize: processedSize,
				total:          totalSize,
				fileName:       f.Name, // 将相对路径作为文件名
			}
			_, err = io.Copy(progressWriter, rc)
			outFile.Close()
			if err != nil {
				rc.Close()
				return err
			}
		}
		rc.Close()

		processedSize += int64(f.UncompressedSize64)
		// progress := float64(processedSize) / float64(totalSize) * 100
		// fmt.Printf("\r解压进度: %.2f%%", progress)
	}

	duration := time.Since(startTime)
	fmt.Printf("\n解压完成！用时: %v, 解压速度: %.2f MB/s\n",
		duration, float64(totalSize)/1024/1024/duration.Seconds())
	return nil
}
