package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	adminURL = "http://localhost:8080/api/v1"
)

func main() {
	fmt.Println("=== 公开API接口测试 ===")

	// 测试健康检查
	fmt.Println("\n1. 测试健康检查...")
	testHealthCheck()

	// 测试系统信息
	fmt.Println("\n2. 测试系统信息...")
	testSystemInfo()

	// 测试验证码API
	fmt.Println("\n3. 测试验证码API...")
	testCaptchaAPI()

	fmt.Println("\n=== 公开API测试完成 ===")
}

func testHealthCheck() error {
	resp, err := http.Get(adminURL + "/health")
	if err != nil {
		fmt.Printf("健康检查请求失败: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return err
	}

	fmt.Printf("健康检查状态码: %d\n", resp.StatusCode)
	fmt.Printf("健康检查响应: %s\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("✅ 健康检查通过")
	} else {
		fmt.Println("❌ 健康检查失败")
	}

	return nil
}

func testSystemInfo() error {
	resp, err := http.Get(adminURL + "/system/info")
	if err != nil {
		fmt.Printf("系统信息请求失败: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return err
	}

	fmt.Printf("系统信息状态码: %d\n", resp.StatusCode)
	fmt.Printf("系统信息响应: %s\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("✅ 系统信息获取成功")
		
		// 解析JSON响应
		var result map[string]interface{}
		if json.Unmarshal(body, &result) == nil {
			if data, ok := result["data"].(map[string]interface{}); ok {
				fmt.Printf("系统版本: %v\n", data["version"])
				fmt.Printf("系统名称: %v\n", data["name"])
				fmt.Printf("启动时间: %v\n", data["start_time"])
			}
		}
	} else {
		fmt.Println("❌ 系统信息获取失败")
	}

	return nil
}

func testCaptchaAPI() error {
	resp, err := http.Get(adminURL + "/auth/captcha")
	if err != nil {
		fmt.Printf("验证码请求失败: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return err
	}

	fmt.Printf("验证码状态码: %d\n", resp.StatusCode)
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ 验证码获取成功")
		
		// 解析JSON响应
		var result map[string]interface{}
		if json.Unmarshal(body, &result) == nil {
			if data, ok := result["data"].(map[string]interface{}); ok {
				if key, ok := data["key"].(string); ok {
					fmt.Printf("验证码Key: %s\n", key)
					
					// 测试验证码调试接口
					time.Sleep(100 * time.Millisecond) // 稍等一下
					testCaptchaDebug(key)
				}
			}
		}
	} else {
		fmt.Printf("❌ 验证码获取失败: %s\n", string(body))
	}

	return nil
}

func testCaptchaDebug(key string) error {
	resp, err := http.Get(fmt.Sprintf("%s/auth/captcha/code?key=%s", adminURL, key))
	if err != nil {
		fmt.Printf("验证码调试请求失败: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return err
	}

	fmt.Printf("验证码调试状态码: %d\n", resp.StatusCode)
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ 验证码调试接口正常")
		
		// 解析JSON响应
		var result map[string]interface{}
		if json.Unmarshal(body, &result) == nil {
			if data, ok := result["data"].(map[string]interface{}); ok {
				if code, ok := data["code"].(string); ok {
					fmt.Printf("验证码明文: %s\n", code)
				}
			}
		}
	} else {
		fmt.Printf("❌ 验证码调试接口失败: %s\n", string(body))
	}

	return nil
}