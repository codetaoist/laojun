package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// User 用户结构
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

// UsersResponse 用户列表响应结构
type UsersResponse struct {
	Message string `json:"message"`
	Data    struct {
		Users []User `json:"users"`
		Total int    `json:"total"`
		Page  int    `json:"page"`
		Limit int    `json:"limit"`
	} `json:"data"`
}

func main() {
	// 使用之前获取的有效JWT token
	jwtToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYWYxZjZmMzQtY2VkMi00YmFmLTg4M2ItNmE1ODgwMTFhNzRlIiwidXNlcm5hbWUiOiJhZG1pbiIsImVtYWlsIjoiYWRtaW5AbGFvanVuIiwiaXNfYWRtaW4iOnRydWUsImlzcyI6Imxhb2p1biIsInN1YiI6ImFmMWY2ZjM0LWNlZDItNGJhZi04ODNiLTZhNTg4MDExYTc0ZSIsImV4cCI6MTc2MTIzNzQ3NiwibmJmIjoxNzYxMTUxMDc2LCJpYXQiOjE3NjExNTEwNzZ9.fcwQbLHPBAiZfVZAwZr2mUDAgdOySnfW0Lna5BpzlWA"
	
	// 测试不同的用户API端点
	testCases := []struct {
		name   string
		url    string
		method string
	}{
		{"获取用户列表", "http://localhost:8080/api/v1/users", "GET"},
		{"获取用户统计", "http://localhost:8080/api/v1/users/stats", "GET"},
		{"获取当前用户信息", "http://localhost:8080/api/v1/users/me", "GET"},
	}
	
	for _, testCase := range testCases {
		fmt.Printf("测试: %s\n", testCase.name)
		fmt.Printf("URL: %s\n", testCase.url)
		
		err := testAPI(jwtToken, testCase.url, testCase.method)
		if err != nil {
			fmt.Printf("  错误: %v\n", err)
		}
		fmt.Println()
	}
}

func testAPI(jwtToken, url, method string) error {
	// 创建HTTP请求
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	
	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}
	
	fmt.Printf("  HTTP状态: %d\n", resp.StatusCode)
	fmt.Printf("  响应内容: %s\n", string(body))
	
	// 如果是成功响应，尝试解析用户数据
	if resp.StatusCode == http.StatusOK {
		if url == "http://localhost:8080/api/v1/users" {
			var usersResp UsersResponse
			if err := json.Unmarshal(body, &usersResp); err == nil {
				fmt.Printf("  解析结果: 找到 %d 个用户\n", usersResp.Data.Total)
				if len(usersResp.Data.Users) > 0 {
					fmt.Printf("  第一个用户: %s (%s)\n", usersResp.Data.Users[0].Username, usersResp.Data.Users[0].Email)
				}
			}
		}
		fmt.Printf("  结果: ✅ 访问成功 - super_admin权限生效\n")
	} else {
		fmt.Printf("  结果: ❌ 访问失败 - 状态码 %d\n", resp.StatusCode)
	}
	
	return nil
}