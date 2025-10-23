package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type CaptchaResponse struct {
	Message string `json:"message"`
	Data    struct {
		Image string `json:"image"`
		Key   string `json:"key"`
	} `json:"data"`
}

type CaptchaDebugResponse struct {
	Message string `json:"message"`
	Data    struct {
		Key  string `json:"key"`
		Code string `json:"code"`
	} `json:"data"`
}

func main() {
	fmt.Println("测试验证码API和调试接口...")
	
	// 1. 获取验证码
	resp, err := http.Get("http://localhost:8080/api/v1/auth/captcha")
	if err != nil {
		fmt.Printf("获取验证码失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}

	fmt.Printf("验证码API状态码: %d\n", resp.StatusCode)
	
	if resp.StatusCode != 200 {
		fmt.Printf("验证码API失败: %s\n", string(body))
		return
	}

	var captchaResp CaptchaResponse
	if err := json.Unmarshal(body, &captchaResp); err != nil {
		fmt.Printf("解析验证码响应失败: %v\n", err)
		return
	}

	fmt.Printf("验证码Key: %s\n", captchaResp.Data.Key)
	fmt.Printf("验证码图片长度: %d字符\n", len(captchaResp.Data.Image))

	// 2. 测试调试接口
	debugURL := fmt.Sprintf("http://localhost:8080/api/v1/auth/captcha/code?key=%s", captchaResp.Data.Key)
	debugResp, err := http.Get(debugURL)
	if err != nil {
		fmt.Printf("获取验证码调试信息失败: %v\n", err)
		return
	}
	defer debugResp.Body.Close()

	debugBody, err := io.ReadAll(debugResp.Body)
	if err != nil {
		fmt.Printf("读取调试响应失败: %v\n", err)
		return
	}

	fmt.Printf("调试API状态码: %d\n", debugResp.StatusCode)
	
	if debugResp.StatusCode != 200 {
		fmt.Printf("调试API失败: %s\n", string(debugBody))
		return
	}

	var debugRespData CaptchaDebugResponse
	if err := json.Unmarshal(debugBody, &debugRespData); err != nil {
		fmt.Printf("解析调试响应失败: %v\n", err)
		return
	}

	fmt.Printf("验证码明文: %s\n", debugRespData.Data.Code)
	fmt.Println("✅ 验证码API和调试接口都正常工作")
}