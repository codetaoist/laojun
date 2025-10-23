package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PermissionCheckRequest 权限检查请求结构
type PermissionCheckRequest struct {
	UserID     string `json:"user_id"`
	DeviceType string `json:"device_type"`
	Module     string `json:"module"`
	Resource   string `json:"resource"`
	Action     string `json:"action"`
}

// PermissionCheckResponse 权限检查响应结构
type PermissionCheckResponse struct {
	Message string `json:"message"`
	Data    struct {
		HasPermission bool   `json:"has_permission"`
		Reason        string `json:"reason"`
	} `json:"data"`
}

func main() {
	// 硬编码的用户ID和JWT token（从之前的验证中获得）
	userID := "af1f6f34-ced2-4baf-883b-6a588011a74e"
	jwtToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYWYxZjZmMzQtY2VkMi00YmFmLTg4M2ItNmE1ODgwMTFhNzRlIiwidXNlcm5hbWUiOiJhZG1pbiIsImVtYWlsIjoiYWRtaW5AbGFvanVuIiwiaXNfYWRtaW4iOnRydWUsImlzcyI6Imxhb2p1biIsInN1YiI6ImFmMWY2ZjM0LWNlZDItNGJhZi04ODNiLTZhNTg4MDExYTc0ZSIsImV4cCI6MTc2MTIzNzQ3NiwibmJmIjoxNzYxMTUxMDc2LCJpYXQiOjE3NjExNTEwNzZ9.fcwQbLHPBAiZfVZAwZr2mUDAgdOySnfW0Lna5BpzlWA"
	
	// 测试不同的权限检查
	testCases := []PermissionCheckRequest{
		{
			UserID:     userID,
			DeviceType: "web",
			Module:     "user",
			Resource:   "users",
			Action:     "read",
		},
		{
			UserID:     userID,
			DeviceType: "web",
			Module:     "user",
			Resource:   "users",
			Action:     "create",
		},
		{
			UserID:     userID,
			DeviceType: "web",
			Module:     "user",
			Resource:   "users",
			Action:     "update",
		},
		{
			UserID:     userID,
			DeviceType: "web",
			Module:     "user",
			Resource:   "users",
			Action:     "delete",
		},
		{
			UserID:     userID,
			DeviceType: "web",
			Module:     "permission",
			Resource:   "roles",
			Action:     "read",
		},
		{
			UserID:     userID,
			DeviceType: "web",
			Module:     "permission",
			Resource:   "permissions",
			Action:     "read",
		},
	}
	
	// 测试权限检查API
	for i, testCase := range testCases {
		fmt.Printf("测试案例 %d: 检查用户 %s 在设备 %s 模块 %s 对资源 %s 的 %s 权限\n", 
			i+1, testCase.UserID[:8], testCase.DeviceType, testCase.Module, testCase.Resource, testCase.Action)
		
		result, err := checkPermission(jwtToken, testCase)
		if err != nil {
			fmt.Printf("  错误: %v\n", err)
		} else {
			fmt.Printf("  结果: 有权限 = %t, 原因 = %s\n", result.Data.HasPermission, result.Data.Reason)
		}
		fmt.Println()
	}
}

func checkPermission(jwtToken string, req PermissionCheckRequest) (*PermissionCheckResponse, error) {
	// 准备请求数据
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}
	
	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", "http://localhost:8080/api/v1/permissions/check", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	
	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+jwtToken)
	
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	
	fmt.Printf("  HTTP状态: %d\n", resp.StatusCode)
	fmt.Printf("  响应内容: %s\n", string(body))
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API返回错误状态: %d, 响应: %s", resp.StatusCode, string(body))
	}
	
	// 解析响应
	var result PermissionCheckResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	
	return &result, nil
}