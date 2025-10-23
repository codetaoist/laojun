package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func main() {
	// 使用之前获取的有效JWT token
	jwtToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYWYxZjZmMzQtY2VkMi00YmFmLTg4M2ItNmE1ODgwMTFhNzRlIiwidXNlcm5hbWUiOiJhZG1pbiIsImVtYWlsIjoiYWRtaW5AbGFvanVuIiwiaXNfYWRtaW4iOnRydWUsImlzcyI6Imxhb2p1biIsInN1YiI6ImFmMWY2ZjM0LWNlZDItNGJhZi04ODNiLTZhNTg4MDExYTc0ZSIsImV4cCI6MTc2MTIzNzQ3NiwibmJmIjoxNzYxMTUxMDc2LCJpYXQiOjE3NjExNTEwNzZ9.fcwQbLHPBAiZfVZAwZr2mUDAgdOySnfW0Lna5BpzlWA"
	
	// 测试需要管理员权限的API端点
	testCases := []struct {
		name        string
		url         string
		method      string
		description string
	}{
		{"权限列表", "http://localhost:8080/api/v1/permissions", "GET", "获取权限列表"},
		{"角色列表", "http://localhost:8080/api/v1/roles", "GET", "获取角色列表"},
		{"权限模板", "http://localhost:8080/api/v1/permission-templates", "GET", "获取权限模板"},
		{"设备类型", "http://localhost:8080/api/v1/permissions/device-types", "GET", "获取设备类型"},
		{"模块列表", "http://localhost:8080/api/v1/permissions/modules", "GET", "获取模块列表"},
		{"用户组列表", "http://localhost:8080/api/v1/permissions/user-groups", "GET", "获取用户组列表"},
		{"系统健康检查", "http://localhost:8080/health", "GET", "系统健康检查"},
		{"系统指标", "http://localhost:8080/metrics", "GET", "系统指标"},
	}
	
	fmt.Printf("🔍 测试super_admin权限对各种管理员API的访问\n")
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))
	
	successCount := 0
	totalCount := len(testCases)
	
	for i, testCase := range testCases {
		fmt.Printf("%d. %s\n", i+1, testCase.name)
		fmt.Printf("   描述: %s\n", testCase.description)
		fmt.Printf("   URL: %s\n", testCase.url)
		
		success := testAPI(jwtToken, testCase.url, testCase.method)
		if success {
			successCount++
		}
		fmt.Println()
	}
	
	fmt.Printf("%s\n", strings.Repeat("=", 60))
	fmt.Printf("📊 测试总结: %d/%d 个API端点访问成功\n", successCount, totalCount)
	
	if successCount == totalCount {
		fmt.Printf("🎉 所有测试通过！super_admin权限正常工作\n")
	} else if successCount > 0 {
		fmt.Printf("⚠️  部分测试通过，super_admin权限基本正常\n")
	} else {
		fmt.Printf("❌ 所有测试失败，super_admin权限可能有问题\n")
	}
}

func testAPI(jwtToken, url, method string) bool {
	// 创建HTTP请求
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Printf("   ❌ 创建请求失败: %v\n", err)
		return false
	}
	
	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("   ❌ 发送请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("   ❌ 读取响应失败: %v\n", err)
		return false
	}
	
	fmt.Printf("   HTTP状态: %d\n", resp.StatusCode)
	
	// 判断是否成功
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("   ✅ 访问成功 - super_admin权限生效\n")
		
		// 如果响应内容不太长，显示部分内容
		if len(body) < 200 {
			fmt.Printf("   响应: %s\n", string(body))
		} else {
			fmt.Printf("   响应长度: %d 字节\n", len(body))
		}
		return true
	} else if resp.StatusCode == 401 {
		fmt.Printf("   ❌ 认证失败 - Token可能无效\n")
		fmt.Printf("   响应: %s\n", string(body))
		return false
	} else if resp.StatusCode == 403 {
		fmt.Printf("   ❌ 权限不足 - super_admin权限未生效\n")
		fmt.Printf("   响应: %s\n", string(body))
		return false
	} else if resp.StatusCode == 404 {
		fmt.Printf("   ⚠️  端点不存在 - API路径可能不正确\n")
		fmt.Printf("   响应: %s\n", string(body))
		return false
	} else {
		fmt.Printf("   ❌ 访问失败 - 状态码 %d\n", resp.StatusCode)
		fmt.Printf("   响应: %s\n", string(body))
		return false
	}
}