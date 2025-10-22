package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/laojun/internal/models"
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
	UserID      uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`   // 操作者ID
	TargetID    uuid.UUID `gorm:"type:uuid" json:"target_id"`          // 目标对象ID
	TargetType  string    `gorm:"size:50;not null" json:"target_type"` // 目标对象类型
	Action      string    `gorm:"size:50;not null" json:"action"`      // 操作类型
	Description string    `gorm:"size:500" json:"description"`         // 操作描述
	OldData     string    `gorm:"type:text" json:"old_data"`           // 变更前数据
	NewData     string    `gorm:"type:text" json:"new_data"`           // 变更后数据
	IPAddress   string    `gorm:"size:45" json:"ip_address"`           // 操作者IP
	UserAgent   string    `gorm:"size:500" json:"user_agent"`          // 用户代理
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
	// 构建角色变更数据
	changeData := s.buildRoleChangeData(targetUserID, targetUsername, oldRoles, newRoles)

	// 序列化角色信息
	oldDataJSON, _ := json.Marshal(s.convertToRoleInfo(oldRoles))
	newDataJSON, _ := json.Marshal(s.convertToRoleInfo(newRoles))

	// 生成操作描述
	description := s.generateRoleChangeDescription(changeData.Changes)

	// 创建审计日志
	auditLog := AuditLog{
		UserID:      operatorID,
		TargetID:    targetUserID,
		TargetType:  "user_roles",
		Action:      "role_assignment",
		Description: fmt.Sprintf("为用户 %s %s", targetUsername, description),
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
		Description: fmt.Sprintf("创建角色: %s", role.Name),
		NewData:     string(roleData),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}

	return s.db.Create(&auditLog).Error
}

// LogRoleUpdate 记录角色更新审计日志
func (s *AuditService) LogRoleUpdate(operatorID uuid.UUID, oldRole, newRole models.Role, ipAddress, userAgent string) error {
	oldData, _ := json.Marshal(oldRole)
	newData, _ := json.Marshal(newRole)

	auditLog := AuditLog{
		UserID:      operatorID,
		TargetID:    newRole.ID,
		TargetType:  "role",
		Action:      "update",
		Description: fmt.Sprintf("更新角色: %s", newRole.Name),
		OldData:     string(oldData),
		NewData:     string(newData),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}

	return s.db.Create(&auditLog).Error
}

// LogRoleDeletion 记录角色删除审计日志
func (s *AuditService) LogRoleDeletion(operatorID uuid.UUID, role models.Role, ipAddress, userAgent string) error {
	roleData, _ := json.Marshal(role)

	auditLog := AuditLog{
		UserID:      operatorID,
		TargetID:    role.ID,
		TargetType:  "role",
		Action:      "delete",
		Description: fmt.Sprintf("删除角色: %s", role.Name),
		OldData:     string(roleData),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now(),
	}

	return s.db.Create(&auditLog).Error
}

// GetAuditLogs 获取审计日志
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
		descriptions = append(descriptions, fmt.Sprintf("新增角色: %v", roleNames))
	}

	if len(changes.Removed) > 0 {
		var roleNames []string
		for _, role := range changes.Removed {
			roleNames = append(roleNames, role.Name)
		}
		descriptions = append(descriptions, fmt.Sprintf("移除角色: %v", roleNames))
	}

	if len(descriptions) == 0 {
		return "角色无变更"
	}

	return fmt.Sprintf("进行了角色变更(%s)", strings.Join(descriptions, ", "))
}
