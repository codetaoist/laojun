package main

import (
	"fmt"
	"log"
	"time"

	"github.com/codetaoist/laojun-shared/utils"
)

func main() {
	fmt.Println("=== 工具包使用示例 ===")

	// 1. 字符串工具
	fmt.Println("\n--- 字符串工具 ---")
	
	testStr := "Hello World"
	fmt.Printf("原始字符串: %s\n", testStr)
	fmt.Printf("是否为空: %t\n", utils.String.IsEmpty(testStr))
	fmt.Printf("是否非空: %t\n", utils.String.IsNotEmpty(testStr))
	fmt.Printf("截断(5字符): %s\n", utils.String.Truncate(testStr, 5))
	fmt.Printf("反转: %s\n", utils.String.Reverse(testStr))
	fmt.Printf("包含'World': %t\n", utils.String.Contains(testStr, "World"))
	
	// 命名转换
	camelCase := "userName"
	snakeCase := "user_name"
	fmt.Printf("驼峰转蛇形 '%s': %s\n", camelCase, utils.String.ToSnakeCase(camelCase))
	fmt.Printf("蛇形转驼峰 '%s': %s\n", snakeCase, utils.String.ToCamelCase(snakeCase))
	
	// 随机字符串
	randomStr := utils.String.RandomString(10)
	fmt.Printf("随机字符串(10位): %s\n", randomStr)

	// 2. 切片工具
	fmt.Println("\n--- 切片工具 ---")
	
	slice1 := []string{"apple", "banana", "cherry"}
	slice2 := []string{"banana", "date", "elderberry"}
	
	fmt.Printf("切片1: %v\n", slice1)
	fmt.Printf("切片2: %v\n", slice2)
	fmt.Printf("包含'banana': %t\n", utils.Slice.Contains(slice1, "banana"))
	
	// 合并和去重
	merged := append(slice1, slice2...)
	unique := utils.Slice.Unique(merged)
	fmt.Printf("合并后: %v\n", merged)
	fmt.Printf("去重后: %v\n", unique)

	// 3. 数字工具
	fmt.Println("\n--- 数字工具 ---")
	
	num1, num2 := 15.5, 23.8
	fmt.Printf("数字1: %.1f, 数字2: %.1f\n", num1, num2)
	fmt.Printf("最大值: %.1f\n", utils.Number.Max(num1, num2))
	fmt.Printf("最小值: %.1f\n", utils.Number.Min(num1, num2))
	fmt.Printf("绝对值(-15.5): %.1f\n", utils.Number.Abs(-15.5))
	fmt.Printf("是否为偶数(16): %t\n", utils.Number.IsEven(16))
	fmt.Printf("是否为偶数(15): %t\n", utils.Number.IsEven(15))

	// 4. 时间工具
	fmt.Println("\n--- 时间工具 ---")
	
	now := time.Now()
	fmt.Printf("当前时间: %s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Printf("格式化时间: %s\n", utils.Time.FormatTime(now, "2006年01月02日 15:04:05"))
	
	// 时间解析
	timeStr := "2024-01-15 14:30:00"
	parsedTime, err := utils.Time.ParseTime(timeStr, "2006-01-02 15:04:05")
	if err != nil {
		log.Printf("时间解析失败: %v", err)
	} else {
		fmt.Printf("解析时间: %s\n", parsedTime.Format("2006-01-02 15:04:05"))
	}

	// 5. 类型转换工具
	fmt.Println("\n--- 类型转换工具 ---")
	
	// 转换为字符串
	intVal := 42
	floatVal := 3.14159
	boolVal := true
	
	fmt.Printf("整数转字符串: %s\n", utils.Convert.ToString(intVal))
	fmt.Printf("浮点数转字符串: %s\n", utils.Convert.ToString(floatVal))
	fmt.Printf("布尔值转字符串: %s\n", utils.Convert.ToString(boolVal))
	
	// 从字符串转换
	strInt := "123"
	strFloat := "45.67"
	strBool := "true"
	
	if convertedInt, err := utils.Convert.ToInt(strInt); err == nil {
		fmt.Printf("字符串转整数: %d\n", convertedInt)
	}
	
	if convertedFloat, err := utils.Convert.ToFloat64(strFloat); err == nil {
		fmt.Printf("字符串转浮点数: %.2f\n", convertedFloat)
	}
	
	if convertedBool, err := utils.Convert.ToBool(strBool); err == nil {
		fmt.Printf("字符串转布尔值: %t\n", convertedBool)
	}

	// 6. JSON工具
	fmt.Println("\n--- JSON工具 ---")
	
	user := map[string]interface{}{
		"id":    1001,
		"name":  "张三",
		"email": "zhangsan@example.com",
		"age":   25,
	}
	
	// 对象转JSON
	jsonStr, err := utils.JSON.ToJSON(user)
	if err != nil {
		log.Printf("JSON序列化失败: %v", err)
	} else {
		fmt.Printf("对象转JSON: %s\n", jsonStr)
		
		// JSON转对象
		var parsedUser map[string]interface{}
		err = utils.JSON.FromJSON(jsonStr, &parsedUser)
		if err != nil {
			log.Printf("JSON反序列化失败: %v", err)
		} else {
			fmt.Printf("JSON转对象: %+v\n", parsedUser)
		}
	}

	// 7. 验证工具
	fmt.Println("\n--- 验证工具 ---")
	
	email := "test@example.com"
	invalidEmail := "invalid-email"
	url := "https://www.example.com"
	invalidUrl := "not-a-url"
	
	fmt.Printf("邮箱验证 '%s': %t\n", email, utils.Validate.IsValidEmail(email))
	fmt.Printf("邮箱验证 '%s': %t\n", invalidEmail, utils.Validate.IsValidEmail(invalidEmail))
	fmt.Printf("URL验证 '%s': %t\n", url, utils.Validate.IsValidURL(url))
	fmt.Printf("URL验证 '%s': %t\n", invalidUrl, utils.Validate.IsValidURL(invalidUrl))

	// 8. 加密工具
	fmt.Println("\n--- 加密工具 ---")
	
	// 生成UUID
	uuid := utils.Crypto.GenerateUUID()
	fmt.Printf("生成UUID: %s\n", uuid)
	
	// 生成随机十六进制字符串
	randomHex, err := utils.Crypto.GenerateRandomHex(16)
	if err != nil {
		log.Printf("生成随机十六进制失败: %v", err)
	} else {
		fmt.Printf("随机十六进制(16字节): %s\n", randomHex)
	}
	
	// 注意：MD5Hash和SHA256Hash方法在当前版本中不可用
	// 如需哈希功能，可以使用标准库的crypto包

	// 9. 分页工具
	fmt.Println("\n--- 分页工具 ---")
	
	// 创建分页对象
	pagination := utils.NewPagination(1, 10) // 第1页，每页10条
	fmt.Printf("分页信息 - 页码: %d, 每页: %d\n", pagination.Page, pagination.PageSize)
	fmt.Printf("偏移量: %d\n", pagination.GetOffset())
	fmt.Printf("限制数量: %d\n", pagination.GetLimit())
	
	// 计算分页信息
	total := int64(95)
	paginationInfo := utils.CalculatePagination(total, pagination.Page, pagination.PageSize)
	fmt.Printf("总记录数: %v\n", paginationInfo["total"])
	fmt.Printf("总页数: %v\n", paginationInfo["total_pages"])
	fmt.Printf("是否有下一页: %v\n", paginationInfo["has_next"])
	fmt.Printf("是否有上一页: %v\n", paginationInfo["has_prev"])

	// 10. Map工具
	fmt.Println("\n--- Map工具 ---")
	
	testMap := map[string]interface{}{
		"name":  "张三",
		"age":   25,
		"email": "zhangsan@example.com",
	}
	
	keys := utils.Map.Keys(testMap)
	values := utils.Map.Values(testMap)
	fmt.Printf("Map键: %v\n", keys)
	fmt.Printf("Map值: %v\n", values)
	
	// 检查键是否存在（使用标准Go语法）
	_, hasName := testMap["name"]
	_, hasPhone := testMap["phone"]
	fmt.Printf("包含键'name': %t\n", hasName)
	fmt.Printf("包含键'phone': %t\n", hasPhone)

	fmt.Println("\n=== 工具包示例完成 ===")
}