package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL = "http://localhost:8080/api/v1"
	adminURL = "http://localhost:8080/api/v1"
)

type LoginRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	Captcha   string `json:"captcha"`
	CaptchaID string `json:"captcha_key"`
}

type LoginResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Token string `json:"token"`
		User  struct {
			ID       string `json:"id"`
			Username string `json:"username"`
			Role     string `json:"role"`
		} `json:"user"`
	} `json:"data"`
	Message string `json:"message"`
}

type ReviewRequest struct {
	Result          string `json:"result"`
	Notes           string `json:"notes"`
	RejectionReason string `json:"rejectionReason,omitempty"`
}

type AppealRequest struct {
	Reason         string `json:"reason"`
	AdditionalInfo string `json:"additionalInfo,omitempty"`
}

func main() {
	fmt.Println("=== 插件审核API接口测试 ===")

	// 1. 获取验证码
	fmt.Println("\n1. 获取验证码...")
	captchaResp, err := getCaptcha()
	if err != nil {
		fmt.Printf("获取验证码失败: %v\n", err)
		return
	}
	
	var captchaID string
	if id, ok := captchaResp["key"]; ok && id != nil {
		captchaID = id.(string)
		fmt.Printf("验证码ID: %s\n", captchaID)
	} else if id, ok := captchaResp["captcha_id"]; ok && id != nil {
		captchaID = id.(string)
		fmt.Printf("验证码ID: %s\n", captchaID)
	} else {
		fmt.Println("验证码ID为空，使用默认值")
		captchaID = "test-captcha-id"
	}
	
	// 获取验证码明文（调试用）
	captchaCode := ""
	if captchaID != "" && captchaID != "test-captcha-id" {
		debugResp, err := http.Get(fmt.Sprintf("%s/auth/captcha/code?key=%s", adminURL, captchaID))
		if err == nil {
			defer debugResp.Body.Close()
			debugBody, _ := io.ReadAll(debugResp.Body)
			fmt.Printf("验证码调试API响应: %s\n", string(debugBody))
			
			var debugResult map[string]interface{}
			if json.Unmarshal(debugBody, &debugResult) == nil {
				if data, ok := debugResult["data"].(map[string]interface{}); ok {
					if code, ok := data["code"].(string); ok {
						captchaCode = code
						fmt.Printf("验证码明文: %s\n", captchaCode)
					}
				}
			}
		}
	}
	
	if captchaCode == "" {
		captchaCode = "test" // 默认验证码
	}
	fmt.Printf("使用验证码: %s\n\n", captchaCode)

	// 2. 登录获取token
	fmt.Println("\n2. 登录获取token...")
	token, err := login("admin", "admin123", captchaCode, captchaID)
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return
	}
	fmt.Printf("登录成功，Token: %s...\n", token[:20])

	// 3. 测试获取审核队列
	fmt.Println("\n3. 测试获取审核队列...")
	err = testGetReviewQueue(token)
	if err != nil {
		fmt.Printf("获取审核队列失败: %v\n", err)
	}

	// 4. 测试获取审核统计
	fmt.Println("\n4. 测试获取审核统计...")
	err = testGetReviewStats(token)
	if err != nil {
		fmt.Printf("获取审核统计失败: %v\n", err)
	}

	// 5. 测试获取审核员工作负载
	fmt.Println("\n5. 测试获取审核员工作负载...")
	err = testGetReviewerWorkload(token)
	if err != nil {
		fmt.Printf("获取审核员工作负载失败: %v\n", err)
	}

	// 6. 测试获取我的审核任务
	fmt.Println("\n6. 测试获取我的审核任务...")
	err = testGetMyReviewTasks(token)
	if err != nil {
		fmt.Printf("获取我的审核任务失败: %v\n", err)
	}

	// 7. 测试审核插件（如果有插件的话）
	fmt.Println("\n7. 测试审核插件...")
	err = testReviewPlugin(token, "test-plugin-id")
	if err != nil {
		fmt.Printf("审核插件失败: %v\n", err)
	}

	// 8. 测试创建申诉
	fmt.Println("\n8. 测试创建申诉...")
	err = testCreateAppeal(token, "test-plugin-id")
	if err != nil {
		fmt.Printf("创建申诉失败: %v\n", err)
	}

	// 9. 测试获取申诉列表
	fmt.Println("\n9. 测试获取申诉列表...")
	err = testGetAppeals(token)
	if err != nil {
		fmt.Printf("获取申诉列表失败: %v\n", err)
	}

	// 10. 测试自动审核
	fmt.Println("\n10. 测试自动审核...")
	err = testAutoReview(token, "test-plugin-id")
	if err != nil {
		fmt.Printf("自动审核失败: %v\n", err)
	}

	fmt.Println("\n=== 测试完成 ===")
}

func getCaptcha() (map[string]interface{}, error) {
	resp, err := http.Get(adminURL + "/auth/captcha")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	fmt.Printf("验证码API响应: %s\n", string(body))
	
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	
	if data, ok := result["data"].(map[string]interface{}); ok {
		return data, nil
	}
	
	return result, nil
}

func login(username, password, captcha, captchaID string) (string, error) {
	loginReq := LoginRequest{
		Username:  username,
		Password:  password,
		Captcha:   captcha,
		CaptchaID: captchaID,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", err
	}
	
	fmt.Printf("发送登录请求参数: %s\n", string(jsonData))

	resp, err := http.Post(adminURL+"/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	fmt.Printf("登录API响应: %s\n", string(body))
	
	var loginResp LoginResponse
	err = json.Unmarshal(body, &loginResp)
	if err != nil {
		return "", fmt.Errorf("JSON解析失败: %v, 响应内容: %s", err, string(body))
	}

	if !loginResp.Success {
		return "", fmt.Errorf("登录失败: %s", loginResp.Message)
	}

	return loginResp.Data.Token, nil
}

func makeAuthenticatedRequest(method, url, token string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

func testGetReviewQueue(token string) error {
	resp, err := makeAuthenticatedRequest("GET", adminURL+"/plugin-review/queue", token, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
	return nil
}

func testGetReviewStats(token string) error {
	resp, err := makeAuthenticatedRequest("GET", adminURL+"/plugin-review/stats", token, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
	return nil
}

func testGetReviewerWorkload(token string) error {
	resp, err := makeAuthenticatedRequest("GET", adminURL+"/plugin-review/workload", token, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
	return nil
}

func testGetMyReviewTasks(token string) error {
	resp, err := makeAuthenticatedRequest("GET", adminURL+"/plugin-review/my-tasks", token, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
	return nil
}

func testReviewPlugin(token, pluginID string) error {
	reviewReq := ReviewRequest{
		Result:          "approved",
		Notes:           "测试审核通过",
		RejectionReason: "",
	}

	resp, err := makeAuthenticatedRequest("POST", adminURL+"/plugin-review/review/"+pluginID, token, reviewReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
	return nil
}

func testCreateAppeal(token, pluginID string) error {
	appealReq := AppealRequest{
		Reason:         "测试申诉原因",
		AdditionalInfo: "这是一个测试申诉",
	}

	resp, err := makeAuthenticatedRequest("POST", adminURL+"/plugin-review/appeal/"+pluginID, token, appealReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
	return nil
}

func testGetAppeals(token string) error {
	resp, err := makeAuthenticatedRequest("GET", adminURL+"/plugin-review/appeals", token, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
	return nil
}

func testAutoReview(token, pluginID string) error {
	resp, err := makeAuthenticatedRequest("POST", adminURL+"/plugin-review/auto-review/"+pluginID, token, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应: %s\n", string(body))
	return nil
}