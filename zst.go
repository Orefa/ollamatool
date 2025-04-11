package main

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/compress/zstd"
)

// 压缩文件
func tarZstFiles(files []string, totalSize int64, srcDir, tarzstfile string) error {
	startTime := time.Now()

	// 创建目标zip文件
	outFile, err := os.Create(tarzstfile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// 创建 zstd 压缩器
	zstdWriter, err := zstd.NewWriter(outFile)
	if err != nil {
		return err
	}
	defer zstdWriter.Close()

	// 创建 tar.Writer
	tarWriter := tar.NewWriter(zstdWriter)
	defer tarWriter.Close()

	// 压缩文件并显示进度
	var processedSize int64
	for _, file := range files {
		fileToCompress, err := os.Open(file)
		if err != nil {
			return err
		}
		defer fileToCompress.Close()

		// 创建 tar 头部
		stat, err := fileToCompress.Stat()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(stat, "")
		if err != nil {
			log.Fatal(err)
		}

		// 获取相对路径
		relPath, err := filepath.Rel(srcDir, file)
		if err != nil {
			return err
		}
		header.Name = relPath

		// 写入 tar header
		err = tarWriter.WriteHeader(header)
		if err != nil {
			return err
		}

		// 创建 ProgressWriter，包装 xzWriter
		progressWriter := &ProgressWriter{
			writer:         tarWriter,
			compressedSize: processedSize,
			total:          totalSize,
			fileName:       relPath, // 将相对路径作为文件名
		}

		// 复制文件内容
		fileInfo, _ := fileToCompress.Stat()
		_, err = io.Copy(progressWriter, fileToCompress)
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
func unZstFile(tarzstfile string, destDir string) error {
	startTime := time.Now()
	// 打开 tar.zst 文件
	inFile, err := os.Open(tarzstfile)
	if err != nil {
		log.Fatal(err)
	}
	defer inFile.Close()
	// 解压 zstd 文件
	zstdReader, err := zstd.NewReader(inFile)
	if err != nil {
		log.Fatal(err)
	}
	defer zstdReader.Close()

	// 读取 tar 文件
	tarReader := tar.NewReader(zstdReader)

	// 计算总大小
	var totalSize int64
	// 解压 tar 文件中的每个文件
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 完成解压
		}
		if err != nil {
			log.Fatal(err)
		}

		// 累加文件大小
		totalSize += header.Size
	}
	// 创建目标目录
	os.MkdirAll(destDir, 0755)

	// 重新打开 tar 文件进行解压并显示进度
	inFile.Seek(0, io.SeekStart)
	zstdReader, err = zstd.NewReader(inFile)
	if err != nil {
		return err
	}

	tarReader = tar.NewReader(zstdReader)

	// 解压文件并显示进度
	var processedSize int64

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			log.Fatal(err)
		}
		path := filepath.Join(destDir, header.Name)
		// 创建文件或目录
		if header.Typeflag == tar.TypeDir {
			// 创建目录
			err := os.MkdirAll(path, header.FileInfo().Mode().Perm())
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// 创建文件并保留文件权限
			os.MkdirAll(filepath.Dir(path), header.FileInfo().Mode().Perm())
			file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, header.FileInfo().Mode())
			if err != nil {
				log.Println()
				return err
			}
			defer file.Close()

			// 创建 ProgressWriter，包装 xzWriter
			progressWriter := &ProgressWriter{
				writer:         file,
				compressedSize: processedSize,
				total:          totalSize,
				fileName:       header.Name, // 将相对路径作为文件名
				uncompress:     true,
			}
			// 将 tar 中的内容复制到文件
			_, err = io.Copy(progressWriter, tarReader)
			if err != nil {
				log.Fatal(err)
			}
			processedSize += int64(progressWriter.progress)
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("\n解压完成！用时: %v, 解压速度: %.2f MB/s\n",
		duration, float64(totalSize)/1024/1024/duration.Seconds())
	return nil
}
