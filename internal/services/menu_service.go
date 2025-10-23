package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/codetaoist/laojun/internal/models"
	"github.com/google/uuid"
)

type MenuService struct {
	db *sql.DB
}

// NewMenuService 创建菜单服务
func NewMenuService(db *sql.DB) *MenuService {
	return &MenuService{
		db: db,
	}
}

// GetMenus 获取菜单列表
func (s *MenuService) GetMenus(params models.MenuSearchParams) (*models.PaginatedResponse[models.Menu], error) {
	// 设置默认分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	// 如果是树形模式，返回树形结构
	if params.TreeMode {
		return s.getMenuTree(params)
	}

	// 构建查询条件
	var conditions []string
	var args []interface{}
	argIndex := 1

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", argIndex))
		args = append(args, "%"+params.Search+"%")
		argIndex++
	}

	if params.ParentID != nil && *params.ParentID != "" {
		if *params.ParentID == "null" {
			conditions = append(conditions, "parent_id IS NULL")
		} else {
			parentUUID, err := uuid.Parse(*params.ParentID)
			if err == nil {
				conditions = append(conditions, fmt.Sprintf("parent_id = $%d", argIndex))
				args = append(args, parentUUID)
				argIndex++
			}
		}
	}

	if params.IsHidden != nil {
		conditions = append(conditions, fmt.Sprintf("is_hidden = $%d", argIndex))
		args = append(args, *params.IsHidden)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM sm_menus %s", whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count menus: %w", err)
	}

	// 查询数据
	offset := (params.Page - 1) * params.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, title, path, icon, component, parent_id, sort_order, is_hidden, is_favorite, created_at, updated_at
		FROM sm_menus %s
		ORDER BY sort_order ASC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	rows, err := s.db.Query(dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query menus: %w", err)
	}
	defer rows.Close()

	var menus []models.Menu
	for rows.Next() {
		var menu models.Menu
		err := rows.Scan(
			&menu.ID, &menu.Title, &menu.Path, &menu.Icon, &menu.Component,
			&menu.ParentID, &menu.SortOrder, &menu.IsHidden, &menu.IsFavorite, &menu.CreatedAt, &menu.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan menu: %w", err)
		}
		menus = append(menus, menu)
	}

	return &models.PaginatedResponse[models.Menu]{
		Data:       menus,
		Total:      total,
		Page:       params.Page,
		Limit:      params.PageSize,
		TotalPages: (total + params.PageSize - 1) / params.PageSize,
	}, nil
}

// getMenuTree 获取菜单树形结构
func (s *MenuService) getMenuTree(params models.MenuSearchParams) (*models.PaginatedResponse[models.Menu], error) {
	// 获取所有菜单
	query := `
		SELECT id, title, path, icon, component, parent_id, sort_order, is_hidden, created_at, updated_at
		FROM sm_menus
		ORDER BY sort_order ASC, created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query menus: %w", err)
	}
	defer rows.Close()

	var allMenus []models.Menu
	menuMap := make(map[uuid.UUID]*models.Menu)

	for rows.Next() {
		var menu models.Menu
		err := rows.Scan(
			&menu.ID, &menu.Title, &menu.Path, &menu.Icon, &menu.Component,
			&menu.ParentID, &menu.SortOrder, &menu.IsHidden, &menu.CreatedAt, &menu.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan menu: %w", err)
		}
		allMenus = append(allMenus, menu)
		menuMap[menu.ID] = &allMenus[len(allMenus)-1]
	}

	// 构建树形结构
	var rootMenus []models.Menu
	for i := range allMenus {
		menu := &allMenus[i]
		if menu.ParentID == nil {
			rootMenus = append(rootMenus, *menu)
		} else {
			if parent, exists := menuMap[*menu.ParentID]; exists {
				if parent.Children == nil {
					parent.Children = make([]*models.Menu, 0)
				}
				parent.Children = append(parent.Children, menu)
			}
		}
	}

	return &models.PaginatedResponse[models.Menu]{
		Data:       rootMenus,
		Total:      len(rootMenus),
		Page:       1,
		Limit:      len(rootMenus),
		TotalPages: 1,
	}, nil
}

// GetMenuByID 根据ID获取菜单
func (s *MenuService) GetMenuByID(id uuid.UUID) (*models.Menu, error) {
	query := `
		SELECT id, title, path, icon, component, parent_id, sort_order, is_hidden, created_at, updated_at
		FROM sm_menus
		WHERE id = $1
	`

	var menu models.Menu
	err := s.db.QueryRow(query, id).Scan(
		&menu.ID, &menu.Title, &menu.Path, &menu.Icon, &menu.Component,
		&menu.ParentID, &menu.SortOrder, &menu.IsHidden, &menu.CreatedAt, &menu.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("menu not found")
		}
		return nil, fmt.Errorf("failed to get menu: %w", err)
	}

	return &menu, nil
}

// CreateMenu 创建菜单
func (s *MenuService) CreateMenu(req models.MenuCreateRequest) (*models.Menu, error) {
	// 验证父菜单是否存在
	if req.ParentID != nil {
		_, err := s.GetMenuByID(*req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent menu not found")
		}
	}

	// 如果没有指定排序，设置为最大排序值加1
	if req.SortOrder == 0 {
		var maxSort int
		query := "SELECT COALESCE(MAX(sort_order), 0) FROM sm_menus WHERE parent_id IS NOT DISTINCT FROM $1"
		err := s.db.QueryRow(query, req.ParentID).Scan(&maxSort)
		if err != nil {
			return nil, fmt.Errorf("failed to get max sort order: %w", err)
		}
		req.SortOrder = maxSort + 1
	}

	// 插入菜单
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO sm_menus (id, title, path, icon, component, parent_id, sort_order, is_hidden, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, title, path, icon, component, parent_id, sort_order, is_hidden, created_at, updated_at
	`

	var menu models.Menu
	err := s.db.QueryRow(query, id, req.Title, req.Path, req.Icon, req.Component,
		req.ParentID, req.SortOrder, req.IsHidden, now, now).Scan(
		&menu.ID, &menu.Title, &menu.Path, &menu.Icon, &menu.Component,
		&menu.ParentID, &menu.SortOrder, &menu.IsHidden, &menu.CreatedAt, &menu.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create menu: %w", err)
	}

	return &menu, nil
}

// UpdateMenu 更新菜单
func (s *MenuService) UpdateMenu(id uuid.UUID, req models.MenuUpdateRequest) (*models.Menu, error) {
	// 检查菜单是否存在
	existingMenu, err := s.GetMenuByID(id)
	if err != nil {
		return nil, err
	}

	// 验证父菜单是否存在
	if req.ParentID != nil && *req.ParentID != uuid.Nil {
		// 检查是否会造成循环引用
		if *req.ParentID == id {
			return nil, fmt.Errorf("cannot set menu as its own parent")
		}

		// 检查父菜单是否存在
		_, err := s.GetMenuByID(*req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent menu not found")
		}

		// 检查是否会造成循环引用（深度检查）
		if s.wouldCreateCycle(id, *req.ParentID) {
			return nil, fmt.Errorf("would create circular reference")
		}
	}

	// 构建更新语句
	var setParts []string
	var args []interface{}
	argIndex := 1

	if req.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}

	if req.Path != nil {
		setParts = append(setParts, fmt.Sprintf("path = $%d", argIndex))
		args = append(args, req.Path)
		argIndex++
	}

	if req.Icon != nil {
		setParts = append(setParts, fmt.Sprintf("icon = $%d", argIndex))
		args = append(args, req.Icon)
		argIndex++
	}

	if req.Component != nil {
		setParts = append(setParts, fmt.Sprintf("component = $%d", argIndex))
		args = append(args, req.Component)
		argIndex++
	}

	if req.ParentID != nil {
		setParts = append(setParts, fmt.Sprintf("parent_id = $%d", argIndex))
		args = append(args, req.ParentID)
		argIndex++
	}

	if req.SortOrder != nil {
		setParts = append(setParts, fmt.Sprintf("sort_order = $%d", argIndex))
		args = append(args, *req.SortOrder)
		argIndex++
	}

	if req.IsHidden != nil {
		setParts = append(setParts, fmt.Sprintf("is_hidden = $%d", argIndex))
		args = append(args, *req.IsHidden)
		argIndex++
	}

	if req.IsFavorite != nil {
		setParts = append(setParts, fmt.Sprintf("is_favorite = $%d", argIndex))
		args = append(args, *req.IsFavorite)
		argIndex++
	}

	if len(setParts) == 0 {
		return existingMenu, nil
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE sm_menus 
		SET %s
		WHERE id = $%d
		RETURNING id, title, path, icon, component, parent_id, sort_order, is_hidden, created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex)

	var menu models.Menu
	err = s.db.QueryRow(query, args...).Scan(
		&menu.ID, &menu.Title, &menu.Path, &menu.Icon, &menu.Component,
		&menu.ParentID, &menu.SortOrder, &menu.IsHidden, &menu.CreatedAt, &menu.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update menu: %w", err)
	}

	return &menu, nil
}

// DeleteMenu 删除菜单
func (s *MenuService) DeleteMenu(id uuid.UUID) error {
	// 检查是否有子菜单
	var childCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM sm_menus WHERE parent_id = $1", id).Scan(&childCount)
	if err != nil {
		return fmt.Errorf("failed to check child menus: %w", err)
	}

	if childCount > 0 {
		return fmt.Errorf("cannot delete menu with child menus")
	}

	// 删除菜单
	result, err := s.db.Exec("DELETE FROM sm_menus WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete menu: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("menu not found")
	}

	return nil
}

// BatchDeleteMenus 批量删除菜单
func (s *MenuService) BatchDeleteMenus(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return fmt.Errorf("no menu IDs provided")
	}

	// 检查是否有子菜单
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM sm_menus WHERE parent_id IN (%s)", strings.Join(placeholders, ","))
	var childCount int
	err := s.db.QueryRow(query, args...).Scan(&childCount)
	if err != nil {
		return fmt.Errorf("failed to check child menus: %w", err)
	}

	if childCount > 0 {
		return fmt.Errorf("cannot delete menus with child menus")
	}

	// 批量删除
	deleteQuery := fmt.Sprintf("DELETE FROM sm_menus WHERE id IN (%s)", strings.Join(placeholders, ","))
	_, err = s.db.Exec(deleteQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to batch delete menus: %w", err)
	}

	return nil
}

// MoveMenu 移动菜单
func (s *MenuService) MoveMenu(id uuid.UUID, req models.MenuMoveRequest) error {
	// 验证目标父菜单是否存在
	if req.TargetParentID != nil && *req.TargetParentID != uuid.Nil {
		if *req.TargetParentID == id {
			return fmt.Errorf("cannot move menu to itself")
		}

		_, err := s.GetMenuByID(*req.TargetParentID)
		if err != nil {
			return fmt.Errorf("target parent menu not found")
		}

		if s.wouldCreateCycle(id, *req.TargetParentID) {
			return fmt.Errorf("would create circular reference")
		}
	}

	// 开始事务
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 更新菜单的父级和排序
	_, err = tx.Exec(`
		UPDATE sm_menus 
		SET parent_id = $1, sort_order = $2, updated_at = $3
		WHERE id = $4
	`, req.TargetParentID, req.TargetIndex, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to move menu: %w", err)
	}

	// 重新排序同级菜单
	err = s.reorderSiblings(tx, req.TargetParentID)
	if err != nil {
		return fmt.Errorf("failed to reorder siblings: %w", err)
	}

	return tx.Commit()
}

// GetMenuStats 获取菜单统计信息
func (s *MenuService) GetMenuStats() (*models.MenuStats, error) {
	var stats models.MenuStats

	// 总菜单数
	err := s.db.QueryRow("SELECT COUNT(*) FROM sm_menus").Scan(&stats.TotalMenus)
	if err != nil {
		return nil, fmt.Errorf("failed to get total menus: %w", err)
	}

	// 可见菜单数
	err = s.db.QueryRow("SELECT COUNT(*) FROM sm_menus WHERE is_hidden = false").Scan(&stats.VisibleMenus)
	if err != nil {
		return nil, fmt.Errorf("failed to get visible menus: %w", err)
	}

	// 隐藏菜单数
	stats.HiddenMenus = stats.TotalMenus - stats.VisibleMenus

	// 最大深度（这里简化实现，实际可能需要递归查询）
	stats.MaxDepth = 3 // 暂时设为固定值，实际应根据数据库结构动态计算
	return &stats, nil
}

// wouldCreateCycle 检查是否会创建循环引用
func (s *MenuService) wouldCreateCycle(menuID, targetParentID uuid.UUID) bool {
	// 简化实现：检查目标父菜单的所有祖先是否包含当前菜单
	currentParentID := &targetParentID
	for currentParentID != nil {
		if *currentParentID == menuID {
			return true
		}

		var nextParentID *uuid.UUID
		err := s.db.QueryRow("SELECT parent_id FROM sm_menus WHERE id = $1", *currentParentID).Scan(&nextParentID)
		if err != nil {
			break
		}
		currentParentID = nextParentID
	}

	return false
}

// reorderSiblings 重新排序同级菜单
func (s *MenuService) reorderSiblings(tx *sql.Tx, parentID *uuid.UUID) error {
	// 获取同级菜单并重新排序
	query := `
		UPDATE sm_menus 
		SET sort_order = subquery.new_order
		FROM (
			SELECT id, ROW_NUMBER() OVER (ORDER BY sort_order, created_at) as new_order
			FROM sm_menus 
			WHERE parent_id IS NOT DISTINCT FROM $1
		) AS subquery
		WHERE sm_menus.id = subquery.id
	`

	_, err := tx.Exec(query, parentID)
	return err
}

// 新增：切换收藏状态
func (s *MenuService) ToggleFavorite(id uuid.UUID) (*models.Menu, error) {
	query := `
		UPDATE sm_menus 
		SET is_favorite = NOT is_favorite, updated_at = $2
		WHERE id = $1
		RETURNING id, title, path, icon, component, parent_id, sort_order, is_hidden, is_favorite, created_at, updated_at
	`
	var menu models.Menu
	err := s.db.QueryRow(query, id, time.Now()).Scan(
		&menu.ID, &menu.Title, &menu.Path, &menu.Icon, &menu.Component,
		&menu.ParentID, &menu.SortOrder, &menu.IsHidden, &menu.IsFavorite, &menu.CreatedAt, &menu.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle favorite: %w", err)
	}
	return &menu, nil
}
