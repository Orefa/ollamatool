package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getFilesAndTotalSize(dir string) ([]string, int64, error) {
	var filePaths []string
	var totalSize int64

	// Walk the directory to find all files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only include files, not directories
		if !info.IsDir() {
			// Add file path to the list
			filePaths = append(filePaths, path)
			// Add file size to total size
			totalSize += info.Size()
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return filePaths, totalSize, nil
}

func getFiles(modelName string, ollamaDir string) (filePaths []string, totalSize int64) {
	var library string
	var name string
	var ver string
	if strings.Contains(modelName, "/") {
		arr := strings.Split(modelName, "/")
		library = arr[0]
		arr2 := strings.Split(arr[1], ":")
		name = arr2[0]
		ver = arr2[1]

	} else {
		library = "library"
		arr2 := strings.Split(modelName, ":")
		name = arr2[0]
		ver = arr2[1]
	}

	floc := filepath.Join(ollamaDir, manifests, library, name, ver)
	// fmt.Println(floc)
	filePaths = append(filePaths, floc)
	f, _ := os.Open(floc)
	fst, _ := f.Stat()
	totalSize = totalSize + fst.Size()
	defer f.Close()
	fileBytes, _ := io.ReadAll(f)
	// 创建 Manifest 类型的变量
	var manifest Manifest

	// 反序列化 JSON 数据
	err := json.Unmarshal(fileBytes, &manifest)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}
	configFileName := strings.ReplaceAll(manifest.Config.Digest, ":", "-")
	// fmt.Println(configFileName)
	configFilePath := filepath.Join(ollamaDir, ollamaBlob, configFileName)
	fif, _ := os.Stat(configFilePath)
	totalSize = totalSize + fif.Size()
	filePaths = append(filePaths, configFilePath)
	// compressfile(xzWriter, configFilePath, ollamaDir)
	for _, v := range manifest.Layers {
		layerFileName := strings.ReplaceAll(v.Digest, ":", "-")
		layerFilePath := filepath.Join(ollamaDir, ollamaBlob, layerFileName)
		fif, _ := os.Stat(layerFilePath)
		totalSize = totalSize + fif.Size()
		filePaths = append(filePaths, layerFilePath)
		// compressfile(xzWriter, layerFilePath, ollamaDir)
		// fmt.Println(layerFileName)
	}
	return filePaths, totalSize
}

// 去重文件路径
func removeDuplicatePaths(files []string) ([]string, int64) {
	// 使用map来去重
	uniqueFiles := make(map[string]struct{})
	for _, file := range files {
		uniqueFiles[file] = struct{}{}
	}

	// 转换map为切片
	var result []string
	var totalSize int64
	for file := range uniqueFiles {
		fif, _ := os.Stat(file)
		totalSize += fif.Size()
		result = append(result, file)
	}
	return result, totalSize
}
