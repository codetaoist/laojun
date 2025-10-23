package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func main() {
	// 这是之前获取的JWT token
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mjk2NTU2ODMsImlhdCI6MTcyOTY1MjA4MywiaXNzIjoidGFpc2hhbmdsYW9qdW4iLCJzdWIiOiJhZjFmNmYzNC1jZWQyLTRiYWYtODgzYi02YTU4ODAxMWE3NGUiLCJ1c2VyX2lkIjoiYWYxZjZmMzQtY2VkMi00YmFmLTg4M2ItNmE1ODgwMTFhNzRlIiwidXNlcm5hbWUiOiJhZG1pbiJ9.signature"

	fmt.Println("解码JWT token...")
	
	// 分割JWT token
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		log.Fatal("无效的JWT token格式")
	}

	// 解码header
	fmt.Println("\n=== JWT Header ===")
	headerData, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		log.Fatal("解码header失败:", err)
	}
	
	var header map[string]interface{}
	if err := json.Unmarshal(headerData, &header); err != nil {
		log.Fatal("解析header JSON失败:", err)
	}
	
	headerJSON, _ := json.MarshalIndent(header, "", "  ")
	fmt.Println(string(headerJSON))

	// 解码payload
	fmt.Println("\n=== JWT Payload ===")
	payloadData, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		log.Fatal("解码payload失败:", err)
	}
	
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		log.Fatal("解析payload JSON失败:", err)
	}
	
	payloadJSON, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Println(string(payloadJSON))

	// 提取关键信息
	fmt.Println("\n=== 关键信息 ===")
	if userID, ok := payload["user_id"].(string); ok {
		fmt.Printf("用户ID: %s\n", userID)
	}
	if username, ok := payload["username"].(string); ok {
		fmt.Printf("用户名: %s\n", username)
	}
	if sub, ok := payload["sub"].(string); ok {
		fmt.Printf("Subject: %s\n", sub)
	}
}