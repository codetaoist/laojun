package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/laojun-admin-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditService 审计服务
type AuditService struct {
	db *gorm.DB
}

// NewAuditService 创建审计服务
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// AuditLog 审计日志模型
type AuditLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`   // Operator ID
	TargetID    uuid.UUID `gorm:"type:uuid" json:"target_id"`          // Target object ID
	TargetType  string    `gorm:"size:50;not null" json:"target_type"` // Target object type
	Action      string    `gorm:"size:50;not null" json:"action"`      // Action type
	Description string    `gorm:"size:500" json:"description"`         // Action description
	OldData     string    `gorm:"type:text" json:"old_data"`           // Data before change
	NewData     string    `gorm:"type:text" json:"new_data"`           // Data after change
	IPAddress   string    `gorm:"size:45" json:"ip_address"`           // Operator IP
	UserAgent   string    `gorm:"size:500" json:"user_agent"`          // User agent
	CreatedAt   time.Time `json:"created_at"`
}

// RoleChangeData 角色变更数据
type RoleChangeData struct {
	UserID   uuid.UUID   `json:"user_id"`
	Username string      `json:"username"`
	OldRoles []RoleInfo  `json:"old_roles"`
	NewRoles []RoleInfo  `json:"new_roles"`
	Changes  RoleChanges `json:"changes"`
}

// RoleInfo 角色信息
type RoleInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
}

// RoleChanges 角色变更详情
type RoleChanges struct {
	Added   []RoleInfo `json:"added"`
	Removed []RoleInfo `json:"removed"`
}

// LogRoleAssignment 记录角色分配审计日志
func (s *AuditService) LogRoleAssignment(
	operatorID uuid.UUID,
	targetUserID uuid.UUID,
	targetUsername string,
	oldRoles []models.Role,
	newRoles []models.Role,
	ipAddress string,
	userAgent string,
) error {
	// Build role change data
	changeData := s.buildRoleChangeData(targetUserID, targetUsername, oldRoles, newRoles)

	// Serialize role info
	oldDataJSON, _ := json.Marshal(s.convertToRoleInfo(oldRoles))
	newDataJSON, _ := json.Marshal(s.convertToRoleInfo(newRoles))

	// Generate operation description
	description := s.generateRoleChangeDescription(changeData.Changes)

	// Create audit log
	auditLog := AuditLog{
		UserID:      operatorID,
		TargetID:    targetUserID,
		TargetType:  "user_roles",
		Action:      "role_assignment",
		Description: fmt.Sprintf("Role assignment for user %s: %s", targetUsername, description),
		OldData:     string(oldDataJSON),
		NewData:     string(newDataJSON),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}

	return s.db.Create(&auditLog).Error
}

// LogRoleCreation 记录角色创建审计日志
func (s *AuditService) LogRoleCreation(operatorID uuid.UUID, role models.Role, ipAddress, userAgent string) error {
	roleData, _ := json.Marshal(role)

	auditLog := AuditLog{
		UserID:      operatorID,
		TargetID:    role.ID,
		TargetType:  "role",
		Action:      "create",
		Description: fmt.Sprintf("Create role: %s", role.Name),
		NewData:     string(roleData),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}

	return s.db.Create(&auditLog).Error
}

// LogRoleUpdate logs role update audit log
func (s *AuditService) LogRoleUpdate(operatorID uuid.UUID, oldRole, newRole models.Role, ipAddress, userAgent string) error {
	oldData, _ := json.Marshal(oldRole)
	newData, _ := json.Marshal(newRole)

	auditLog := AuditLog{
		UserID:      operatorID,
		TargetID:    newRole.ID,
		TargetType:  "role",
		Action:      "update",
		Description: fmt.Sprintf("Update role: %s", newRole.Name),
		OldData:     string(oldData),
		NewData:     string(newData),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}

	return s.db.Create(&auditLog).Error
}

// LogRoleDeletion logs role deletion audit log
func (s *AuditService) LogRoleDeletion(operatorID uuid.UUID, role models.Role, ipAddress, userAgent string) error {
	roleData, _ := json.Marshal(role)

	auditLog := AuditLog{
		UserID:      operatorID,
		TargetID:    role.ID,
		TargetType:  "role",
		Action:      "delete",
		Description: fmt.Sprintf("Delete role: %s", role.Name),
		OldData:     string(roleData),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}

	return s.db.Create(&auditLog).Error
}

// GetAuditLogs gets audit logs
func (s *AuditService) GetAuditLogs(page, pageSize int, filters map[string]interface{}) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := s.db.Model(&AuditLog{})

	// 应用过滤条件
	if userID, ok := filters["user_id"]; ok {
		query = query.Where("user_id = ?", userID)
	}
	if targetType, ok := filters["target_type"]; ok {
		query = query.Where("target_type = ?", targetType)
	}
	if action, ok := filters["action"]; ok {
		query = query.Where("action = ?", action)
	}
	if startDate, ok := filters["start_date"]; ok {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filters["end_date"]; ok {
		query = query.Where("created_at <= ?", endDate)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// buildRoleChangeData 构建角色变更数据
func (s *AuditService) buildRoleChangeData(userID uuid.UUID, username string, oldRoles, newRoles []models.Role) RoleChangeData {
	oldRoleMap := make(map[uuid.UUID]models.Role)
	newRoleMap := make(map[uuid.UUID]models.Role)

	for _, role := range oldRoles {
		oldRoleMap[role.ID] = role
	}
	for _, role := range newRoles {
		newRoleMap[role.ID] = role
	}

	var added, removed []RoleInfo

	// 找出新增的角色
	for _, role := range newRoles {
		if _, exists := oldRoleMap[role.ID]; !exists {
			added = append(added, s.convertRoleToInfo(role))
		}
	}

	// 找出移除的角色
	for _, role := range oldRoles {
		if _, exists := newRoleMap[role.ID]; !exists {
			removed = append(removed, s.convertRoleToInfo(role))
		}
	}

	return RoleChangeData{
		UserID:   userID,
		Username: username,
		OldRoles: s.convertToRoleInfo(oldRoles),
		NewRoles: s.convertToRoleInfo(newRoles),
		Changes: RoleChanges{
			Added:   added,
			Removed: removed,
		},
	}
}

// convertToRoleInfo 转换角色为角色信息
func (s *AuditService) convertToRoleInfo(roles []models.Role) []RoleInfo {
	var roleInfos []RoleInfo
	for _, role := range roles {
		roleInfos = append(roleInfos, s.convertRoleToInfo(role))
	}
	return roleInfos
}

// convertRoleToInfo 转换单个角色为角色信息
func (s *AuditService) convertRoleToInfo(role models.Role) RoleInfo {
	var desc string
	if role.Description != nil {
		desc = *role.Description
	}
	return RoleInfo{
		ID:          role.ID,
		Name:        role.Name,
		Description: desc,
		IsSystem:    role.IsSystem,
	}
}

// generateRoleChangeDescription 生成角色变更描述
func (s *AuditService) generateRoleChangeDescription(changes RoleChanges) string {
	var descriptions []string

	if len(changes.Added) > 0 {
		var roleNames []string
		for _, role := range changes.Added {
			roleNames = append(roleNames, role.Name)
		}
		descriptions = append(descriptions, fmt.Sprintf("Added roles: %v", roleNames))
	}

	if len(changes.Removed) > 0 {
		var roleNames []string
		for _, role := range changes.Removed {
			roleNames = append(roleNames, role.Name)
		}
		descriptions = append(descriptions, fmt.Sprintf("Removed roles: %v", roleNames))
	}

	if len(descriptions) == 0 {
		return "No role changes"
	}

	return fmt.Sprintf("Role changes: %s", strings.Join(descriptions, ", "))
}
