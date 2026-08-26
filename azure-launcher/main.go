package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	payloadURL    string
	payloadSHA256 string
)

const (
	runtimeFolder = "正文校样室-运行文件"
	appExecutable = "正文校样室.exe"
)

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	messageBoxW   = user32.NewProc("MessageBoxW")
	mbIconError   = uintptr(0x10)
)

func main() {
	if payloadURL == "" || payloadSHA256 == "" {
		showMessage("正文校样室 · 蔚蓝版", "启动器尚未完成发布配置，请重新下载正式版本。", mbIconError)
		return
	}

	executablePath, err := os.Executable()
	if err != nil {
		showMessage("正文校样室 · 蔚蓝版", "无法确定启动器位置："+err.Error(), mbIconError)
		return
	}

	baseDirectory := filepath.Dir(executablePath)
	runtimeDirectory := filepath.Join(baseDirectory, runtimeFolder)
	applicationPath := filepath.Join(runtimeDirectory, appExecutable)

	if !isRegularFile(applicationPath) {
		if err := installPayload(baseDirectory, runtimeDirectory); err != nil {
			showMessage("正文校样室 · 蔚蓝版", "首次准备运行文件失败。\n\n"+err.Error()+"\n\n请检查网络连接后重新启动。", mbIconError)
			return
		}
	}

	command := exec.Command(applicationPath)
	command.Dir = runtimeDirectory
	if err := command.Start(); err != nil {
		showMessage("正文校样室 · 蔚蓝版", "运行文件已准备完成，但无法启动正文校样室。\n\n"+err.Error(), mbIconError)
	}
}

func installPayload(baseDirectory, runtimeDirectory string) error {
	temporaryDirectory := runtimeDirectory + ".installing"
	archivePath := filepath.Join(baseDirectory, ".morroblog-azure-payload.tar.gz")

	if err := os.RemoveAll(temporaryDirectory); err != nil {
		return fmt.Errorf("无法清理临时文件：%w", err)
	}
	if err := os.MkdirAll(temporaryDirectory, 0o755); err != nil {
		return fmt.Errorf("无法创建运行文件夹：%w", err)
	}
	defer os.RemoveAll(temporaryDirectory)
	defer os.Remove(archivePath)

	if err := downloadPayload(archivePath); err != nil {
		return err
	}
	if err := verifySHA256(archivePath, payloadSHA256); err != nil {
		return err
	}
	if err := extractArchive(archivePath, temporaryDirectory); err != nil {
		return err
	}
	if !isRegularFile(filepath.Join(temporaryDirectory, appExecutable)) {
		return errors.New("下载的运行载荷不完整")
	}

	if err := os.RemoveAll(runtimeDirectory); err != nil {
		return fmt.Errorf("无法替换旧运行文件：%w", err)
	}
	if err := os.Rename(temporaryDirectory, runtimeDirectory); err != nil {
		return fmt.Errorf("无法完成运行文件安装：%w", err)
	}
	return nil
}

func downloadPayload(destination string) error {
	attempts := []struct {
		name   string
		client *http.Client
	}{
		{
			name: "直连",
			client: &http.Client{
				Timeout: 12 * time.Minute,
				Transport: &http.Transport{
					Proxy: nil,
				},
			},
		},
		{
			name:   "系统代理",
			client: &http.Client{Timeout: 12 * time.Minute},
		},
	}

	var failures []string
	for _, attempt := range attempts {
		if err := downloadWithClient(attempt.client, destination); err == nil {
			return nil
		} else {
			failures = append(failures, attempt.name+"："+err.Error())
			_ = os.Remove(destination)
		}
	}
	return fmt.Errorf("无法下载运行文件。已依次尝试直连和系统代理：%s", strings.Join(failures, "；"))
}

func downloadWithClient(client *http.Client, destination string) error {
	response, err := client.Get(payloadURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回 %s", response.Status)
	}

	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, response.Body); err != nil {
		return err
	}
	return nil
}

func verifySHA256(filePath, expected string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("无法校验下载文件：%w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("无法读取下载文件：%w", err)
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return errors.New("下载文件校验失败，已拒绝使用损坏或不匹配的运行文件")
	}
	return nil
}

func extractArchive(archivePath, destination string) error {
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("无法打开运行文件压缩包：%w", err)
	}
	defer archiveFile.Close()

	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		return fmt.Errorf("无法解压运行文件：%w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	cleanDestination := filepath.Clean(destination)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("无法读取运行文件压缩包：%w", err)
		}

		target := filepath.Clean(filepath.Join(destination, header.Name))
		if target != cleanDestination && !strings.HasPrefix(target, cleanDestination+string(os.PathSeparator)) {
			return errors.New("运行文件压缩包包含非法路径")
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(file, tarReader)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		case tar.TypeSymlink, tar.TypeLink:
			return errors.New("运行文件压缩包包含不允许的链接文件")
		}
	}
	return nil
}

func isRegularFile(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && !info.IsDir()
}

func showMessage(title, message string, flags uintptr) {
	titleUTF16, _ := syscall.UTF16PtrFromString(title)
	messageUTF16, _ := syscall.UTF16PtrFromString(message)
	messageBoxW.Call(0, uintptr(unsafe.Pointer(messageUTF16)), uintptr(unsafe.Pointer(titleUTF16)), flags)
}
