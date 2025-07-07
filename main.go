package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// var ollamaDir = `E:\.ollama`

var manifests = `manifests\`
var ollamaBlob = `blobs\`

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

	// rootCmd.PersistentFlags().String("OLLAMA_MODELS", "", `Ollama Model存储路径`)

	var exportCmd = &cobra.Command{
		Use: "export",
		Args: func(cmd *cobra.Command, args []string) error {
			modelName, _ := cmd.Flags().GetString("modelName")
			if modelName == "" || !strings.Contains(modelName, ":") {
				// cmd.Help()
				return fmt.Errorf("you must provide correct model names")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			modelPath := Models()
			modelName, _ := cmd.Flags().GetString("modelName")
			if strings.Contains(modelName, ",") {
				arr := strings.Split(modelName, ",")
				var files []string
				for _, v := range arr {
					filePaths, _ := getFiles(v, modelPath)
					files = append(files, filePaths...)
				}
				disfiles, totalSize := removeDuplicatePaths(files)
				outputFile := strings.ReplaceAll(modelName, ":", "-") + ".tar.zst"
				outputFile = strings.ReplaceAll(outputFile, "/", "_")
				outputFile = strings.ReplaceAll(outputFile, ",", "&")

				tarZstFiles(disfiles, totalSize, modelPath, outputFile)
			} else {
				filePaths, totalSize := getFiles(modelName, modelPath)
				outputFile := strings.ReplaceAll(modelName, ":", "-") + ".tar.zst"
				outputFile = strings.ReplaceAll(outputFile, "/", "_")

				// filePaths, totalSize, _ := getFilesAndTotalSize(`D:\dist\lib`)
				tarZstFiles(filePaths, totalSize, modelPath, outputFile)
			}

			// compress(filePaths, totalSize, modelPath, outputFile)
		},
	}
	exportCmd.PersistentFlags().String("modelName", "", `model name`)

	rootCmd.AddCommand(exportCmd)

	var importCommand = &cobra.Command{
		Use: "import",
		Args: func(cmd *cobra.Command, args []string) error {
			importFile, _ := cmd.Flags().GetString("importFile")
			if importFile == "" || !fileExists(importFile) {
				// cmd.Help()
				return fmt.Errorf("you must provide correct import model file")
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			importFile, _ := cmd.Flags().GetString("importFile")
			modelPath := Models()
			// uncompress(importFile, modelPath)
			unZstFile(importFile, modelPath)
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

// Var returns an environment variable stripped of leading and trailing quotes or spaces
func Var(key string) string {
	return strings.Trim(strings.TrimSpace(os.Getenv(key)), "\"'")
}

// Default is $HOME/.ollama/models
func Models() string {
	if s := Var("OLLAMA_MODELS"); s != "" {
		return s
	}

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(home, ".ollama", "models")
}
