package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strconv"
)

func main() {
	// 读取文件内容
	data, err := ioutil.ReadFile("test.png")
	if err != nil {
		panic(err)
	}

	// 将文件内容转换为十六进制字符串
	hexdata := hex.EncodeToString(data)

	// 获取 PNG 图片的宽度和高度
	if hexdata[:16] == "89504e470d0a1a0a" {
		width, err := strconv.ParseInt(hexdata[36:40], 16, 32)
		if err != nil {
			panic(err)
		}
		height, err := strconv.ParseInt(hexdata[40:48], 16, 32)
		if err != nil {
			panic(err)
		}
		fmt.Printf("PNG图片宽度：0x%x | %d\nPNG图片高度：0x%x | %d\n", width, width, height, height)
		// fmt.Printf("PNG图片宽度：%s | %d\nPNG图片高度：%s | %d\n", hex.EncodeToString([]byte{byte(width >> 24), byte(width >> 16), byte(width >> 8), byte(width)}), width, hex.EncodeToString([]byte{byte(height >> 24), byte(height >> 16), byte(height >> 8), byte(height)}), height)
	} else {
		fmt.Println("可能不是PNG文件，或文件头有修改。")
	}
}
