package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	// 定义表名映射
	tableMapping := map[string]string{
		"lj_device_types":           "sm_device_types",
		"lj_modules":                "sm_modules",
		"lj_user_group_members":     "ug_user_group_members",
		"lj_user_group_permissions": "ug_user_group_permissions",
		"lj_permission_templates":   "ug_permission_templates",
		"lj_permissions":            "az_permissions",
		"lj_menus":                  "sm_menus",
		"lj_jwt_keys":               "ua_jwt_keys",
		"lj_icon_library":           "sys_icons",
		"lj_system_settings":        "sys_settings",
		"lj_audit_logs":             "sys_audit_logs",
		"lj_extended_permissions":   "pe_extended_permissions",
		"lj_user_device_permissions": "pe_user_device_permissions",
		"lj_user_groups":            "ug_user_groups",
		"lj_users":                  "ua_admin",
		"lj_roles":                  "az_roles",
		"lj_user_roles":             "az_user_roles",
		"lj_menu_configs":           "sm_menu_configs",
		"lj_user_menu_preferences":  "sm_user_menu_preferences",
	}

	// 需要修复的文件列表
	files := []string{
		"internal/services/permission_service.go",
		"internal/services/system_service.go",
		"internal/services/jwt_key_service.go",
		"internal/services/icon_service.go",
		"internal/services/menu_service.go",
		"pkg/shared/middleware/auth.go",
	}

	for _, file := range files {
		err := fixTableNamesInFile(file, tableMapping)
		if err != nil {
			log.Printf("修复文件 %s 失败: %v", file, err)
		} else {
			fmt.Printf("成功修复文件: %s\n", file)
		}
	}

	fmt.Println("表名前缀修复完成!")
}

func fixTableNamesInFile(filename string, tableMapping map[string]string) error {
	// 读取文件内容
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// 替换表名
	for oldTable, newTable := range tableMapping {
		contentStr = strings.ReplaceAll(contentStr, oldTable, newTable)
	}

	// 写回文件
	err = ioutil.WriteFile(filename, []byte(contentStr), 0644)
	if err != nil {
		return err
	}

	return nil
}