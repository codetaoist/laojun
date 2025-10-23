package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Message string `json:"message"`
	Data    struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expiresAt"`
		User      struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		} `json:"user"`
	} `json:"data"`
}

func main() {
	// 登录获取新token
	loginReq := LoginRequest{
		Username: "admin",
		Password: "admin123",
	}
	
	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		fmt.Printf("序列化登录请求失败: %v\n", err)
		return
	}
	
	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", "http://localhost:8080/api/v1/login", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		return
	}
	
	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("发送请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		return
	}
	
	fmt.Printf("HTTP状态: %d\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", string(body))
	
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("登录失败，状态码: %d\n", resp.StatusCode)
		return
	}
	
	// 解析响应
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		fmt.Printf("解析响应失败: %v\n", err)
		return
	}
	
	if loginResp.Message != "登录成功" {
		fmt.Printf("登录失败: %s\n", loginResp.Message)
		return
	}
	
	fmt.Printf("\n登录成功！\n")
	fmt.Printf("用户ID: %s\n", loginResp.Data.User.ID)
	fmt.Printf("用户名: %s\n", loginResp.Data.User.Username)
	fmt.Printf("邮箱: %s\n", loginResp.Data.User.Email)
	fmt.Printf("Token过期时间: %s\n", loginResp.Data.ExpiresAt)
	fmt.Printf("\nAccess Token:\n%s\n", loginResp.Data.Token)
}