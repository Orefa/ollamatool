package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// var ollamaDir = `E:\.ollama`

var manifests = `models\manifests\registry.ollama.ai\`
var ollamaBlob = `models\blobs\`

func main() {
	// 获取当前执行文件的路径
	executablePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 获取文件名
	fileName := filepath.Base(executablePath)

	var rootCmd = &cobra.Command{
		Use: fileName,
		// Short: "This is a root command",
		// Long:  "A longer description of the root command",
	}
	// 隐藏命令，不代表不存在
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.PersistentFlags().String("OLLAMA_MODELS", "", `Ollama Model存储路径`)

	var exportCmd = &cobra.Command{
		Use: "export",
		Args: func(cmd *cobra.Command, args []string) error {
			modelPath, _ := cmd.Flags().GetString("OLLAMA_MODELS")
			if modelPath == "" {
				// cmd.Help()
				return fmt.Errorf("you must provide Ollama Model存储路径")
			}
			modelName, _ := cmd.Flags().GetString("modelName")
			if modelName == "" || !strings.Contains(modelName, ":") {
				// cmd.Help()
				return fmt.Errorf("you must provide correct model names")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			modelPath, _ := cmd.Flags().GetString("OLLAMA_MODELS")
			modelName, _ := cmd.Flags().GetString("modelName")
			if strings.Contains(modelName, ",") {
				arr := strings.Split(modelName, ",")
				var files []string
				for _, v := range arr {
					filePaths, _ := getFiles(v, modelPath)
					files = append(files, filePaths...)
				}
				disfiles, totalSize := removeDuplicatePaths(files)
				outputFile := strings.ReplaceAll(modelName, ":", "-") + ".zip"
				outputFile = strings.ReplaceAll(outputFile, "/", "_")
				outputFile = strings.ReplaceAll(outputFile, ",", "&")

				zipFiles(disfiles, totalSize, modelPath, outputFile)
			} else {
				filePaths, totalSize := getFiles(modelName, modelPath)
				outputFile := strings.ReplaceAll(modelName, ":", "-") + ".zip"
				outputFile = strings.ReplaceAll(outputFile, "/", "_")

				// filePaths, totalSize, _ := getFilesAndTotalSize(`D:\dist\lib`)
				zipFiles(filePaths, totalSize, modelPath, outputFile)
			}

			// compress(filePaths, totalSize, modelPath, outputFile)
		},
	}
	exportCmd.PersistentFlags().String("modelName", "", `model name`)

	rootCmd.AddCommand(exportCmd)

	var importCommand = &cobra.Command{
		Use: "import",
		Args: func(cmd *cobra.Command, args []string) error {
			modelPath, _ := cmd.Flags().GetString("OLLAMA_MODELS")
			if modelPath == "" {
				// cmd.Help()
				return fmt.Errorf("you must provide Ollama Model存储路径")
			}
			importFile, _ := cmd.Flags().GetString("importFile")
			if importFile == "" || !fileExists(importFile) {
				// cmd.Help()
				return fmt.Errorf("you must provide correct import model file")
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			importFile, _ := cmd.Flags().GetString("importFile")
			modelPath, _ := cmd.Flags().GetString("OLLAMA_MODELS")
			// uncompress(importFile, modelPath)
			unzipFile(importFile, modelPath)
		},
	}
	importCommand.PersistentFlags().String("importFile", "", `import model file`)

	rootCmd.AddCommand(importCommand)
	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// outputFile = "output.xz"

	// uncompress(outputFile, "myfiles")
}

func fileExists(filePath string) bool {
	_, err := os.Open(filePath) // 尝试打开文件
	if err != nil {
		if os.IsNotExist(err) {
			return false // 文件不存在
		}
		return false // 其他错误
	}
	return true // 文件存在
}
