package helper

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
)

//获取md5值
func Md5(str string) string {
	sum := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	return string(sum)
}

//读取文件
func FileGetContents(file string) ([]byte, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

//写入文件
func FilePutContents(file string, content string, append bool) error {
	var f *os.File
	var err error
	os.Create(file)
	if append {
		f, err = os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	} else {
		f, err = os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	}
	defer f.Close()
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	return err
}
