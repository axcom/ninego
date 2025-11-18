//全局对象单元
package skit

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

//当前项目根目录(当前执行文件在根目录的子目录下时)
func RootDir() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	realPath, err := filepath.EvalSymlinks(exPath)
	if err != nil {
		panic(err)
	}
	return filepath.Dir(realPath)
}

//App文件名（无路径）
func AppFileName() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	return filepath.Base(path)
}

//App程序名（无文件后缀）
func AppName() string {
	filename := AppFileName()
	fileSuffix := path.Ext(filename)
	return strings.TrimSuffix(filename, fileSuffix)
}

//App全路径文件名
func AppPath() string {
	/*file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	return strings.Replace(path, "\\", "/", -1)*/
	s, err := exec.LookPath(os.Args[0])
	CheckErr(err)
	i := strings.LastIndex(s, "\\")
	path := string(s[0 : i+1])
	return path
}

//App文件目录名
func AppDir() string {
	path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return strings.Replace(path, "\\", "/", -1)
}

//判断文件或文件夹是否存在
func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
	/*if err == nil {
		return true, err
	}
	if os.IsNotExist(err) { //IsNotExist表示不存在、否则表示不确定
		return false, nil
	}
	return false, err*/
}

// return length in bytes for regular files
func FileSize(file string) int64 {
	f, e := os.Stat(file)
	if e != nil {
		return 0
	}
	return f.Size()
}

// joinFilePath joins path & file into a single string
func JoinFilePath(path, file string) string {
	return filepath.Join(path, file)
}

// ChangeFileExt 修改文件名的扩展名。
// 如果 Extension 为空，则移除扩展名。
// 如果 Extension 不以"."开头，函数会自动添加。
// 示例:
// ChangeFileExt("image.jpg", "png")      -> "image.png"
// ChangeFileExt("document.txt", ".pdf")  -> "document.pdf"
// ChangeFileExt("archive.zip", "")       -> "archive"
func ChangeFileExt(fileName, extension string) string {
	// 获取不带扩展名的文件名
	withoutExt := filepath.Base(fileName[:len(fileName)-len(filepath.Ext(fileName))])

	if extension == "" {
		return withoutExt
	}

	// 确保扩展名以"."开头
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	return withoutExt + extension
}

// ChangeFilePath 修改文件的路径部分，保持文件名和扩展名不变。
// 如果 Path 为空，则返回仅包含文件名的结果。
// 示例:
// ChangeFilePath("/home/user/doc.txt", "/tmp")  -> "/tmp/doc.txt"
// ChangeFilePath("C:\\Users\\file.exe", "D:\\Data") -> "D:\\Data\\file.exe"
// ChangeFilePath("image.png", "")            -> "image.png"
func ChangeFilePath(fileName, path string) string {
	return filepath.Join(path, filepath.Base(fileName))
}

// ExtractFilePath 从文件名中提取路径部分（包括最后一个目录分隔符）。
// 如果 fileName 没有路径部分，则返回空字符串。
// 这与 Delphi 的 ExtractFilePath 行为完全一致。
// 示例:
// ExtractFilePath("/home/user/doc.txt")  -> "/home/user/"
// ExtractFilePath("C:\\Users\\file.exe") -> "C:\\Users\\"
// ExtractFilePath("image.png")           -> ""
func ExtractFilePath(fileName string) string {
	dir := filepath.Dir(fileName)
	if dir == "." || dir == fileName {
		return ""
	}
	return addTrailingSeparator(dir)
}

// ExtractFileDir 从文件名中提取目录部分（不包括最后一个目录分隔符）。
// 如果 fileName 没有路径部分，则返回"."。
// 示例:
// ExtractFileDir("/home/user/doc.txt")  -> "/home/user"
// ExtractFileDir("C:\\Users\\file.exe") -> "C:\\Users"
// ExtractFileDir("image.png")           -> "."
func ExtractFileDir(fileName string) string {
	return filepath.Dir(fileName)
}

// ExtractFileName 从文件名（包括路径）中提取文件名和扩展名。
// 示例:
// ExtractFileName("/home/user/doc.txt")  -> "doc.txt"
// ExtractFileName("C:\\Users\\file.exe") -> "file.exe"
// ExtractFileName("/usr/bin/ls")         -> "ls"
func ExtractFileName(fileName string) string {
	return filepath.Base(fileName)
}

// ExtractFileExt 从文件名中提取扩展名（包括开头的"."）。
// 如果文件没有扩展名，则返回空字符串。
// 示例:
// ExtractFileExt("image.jpg")  -> ".jpg"
// ExtractFileExt("document")   -> ""
// ExtractFileExt("archive.tar.gz") -> ".gz"
func ExtractFileExt(fileName string) string {
	return filepath.Ext(fileName)
}

// addTrailingSeparator 确保路径字符串以系统的路径分隔符结尾。
func addTrailingSeparator(path string) string {
	if path == "" {
		return ""
	}
	sep := string(filepath.Separator)
	if !strings.HasSuffix(path, sep) {
		path += sep
	}
	return path
}
