package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	fmt.Println("测试验证码API...")
	
	// 测试验证码API
	resp, err := http.Get("http://localhost:8080/api/v1/auth/captcha")
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ 验证码API正常工作")
	} else {
		fmt.Println("❌ 验证码API返回错误")
	}
}