package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	internalmodels "github.com/codetaoist/laojun-marketplace-api/internal/models"
	"github.com/codetaoist/laojun-marketplace-api/internal/plugin"
	"github.com/codetaoist/laojun-shared/database"
	"github.com/codetaoist/laojun-shared/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// 类型别名
type MicroservicePlugin = plugin.MicroservicePlugin

// ExtendedPluginService 扩展插件服务
type ExtendedPluginService struct {
	db                  *database.DB
	inProcessManager    *plugin.PluginLoaderManager
	microserviceManager *plugin.MicroservicePluginManager
	gateway             *plugin.PluginGateway
	logger              *logrus.Logger
}

// NewExtendedPluginService 创建扩展插件服务
func NewExtendedPluginService(
	db *database.DB,
	inProcessManager *plugin.PluginLoaderManager,
	microserviceManager *plugin.MicroservicePluginManager,
	gateway *plugin.PluginGateway,
	logger *logrus.Logger,
) *ExtendedPluginService {
	return &ExtendedPluginService{
		db:                  db,
		inProcessManager:    inProcessManager,
		microserviceManager: microserviceManager,
		gateway:             gateway,
		logger:              logger,
	}
}

// GetExtendedPlugins 获取扩展插件列表
func (s *ExtendedPluginService) GetExtendedPlugins(params internalmodels.PluginSearchParams) ([]internalmodels.ExtendedPlugin, models.PaginationMeta, error) {
	var plugins []internalmodels.ExtendedPlugin
	var totalCount int

	// 构建查询条件
	whereConditions := []string{"p.is_active = true"}
	args := []interface{}{}
	argIndex := 1

	if params.CategoryID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.category_id = $%d", argIndex))
		args = append(args, *params.CategoryID)
		argIndex++
	}

	if params.Featured != nil && *params.Featured {
		whereConditions = append(whereConditions, fmt.Sprintf("p.is_featured = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	if params.Query != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", argIndex, argIndex+1))
		searchTerm := "%" + params.Query + "%"
		args = append(args, searchTerm, searchTerm)
		argIndex += 2
	}

	// 添加插件类型过滤
	if params.Type != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.type = $%d", argIndex))
		args = append(args, *params.Type)
		argIndex++
	}

	if params.Runtime != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.runtime = $%d", argIndex))
		args = append(args, *params.Runtime)
		argIndex++
	}

	if params.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.status = $%d", argIndex))
		args = append(args, *params.Status)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// 获取总数
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM mp_plugins p 
		WHERE %s`, whereClause)

	err := s.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}

	// 构建排序
	orderBy := "p.created_at DESC"
	if params.SortBy != "" {
		switch params.SortBy {
		case "name":
			orderBy = "p.name ASC"
		case "rating":
			orderBy = "p.rating DESC"
		case "downloads":
			orderBy = "p.download_count DESC"
		case "price":
			orderBy = "p.price ASC"
		case "updated":
			orderBy = "p.updated_at DESC"
		}
	}

	// 计算偏移量
	offset := (params.Page - 1) * params.PageSize

	// 查询插件列表
	query := fmt.Sprintf(`
		SELECT 
			p.id, p.name, p.description, p.short_description, p.author, p.developer_id, 
			p.version, p.icon_url, p.banner_url, p.price, p.rating, p.download_count, 
			p.is_featured, p.created_at, p.updated_at, p.category_id,
			p.type, p.runtime, p.status, p.interface_spec, p.dependencies, 
			p.runtime_config, p.security_policy, p.resource_limits,
			p.entry_point, p.exported_symbols, p.binary_path,
			p.docker_image, p.service_port, p.health_check_path, 
			p.grpc_proto_file, p.service_endpoint, p.namespace,
			c.name as category_name, c.icon as category_icon, c.color as category_color
		FROM mp_plugins p
		LEFT JOIN mp_categories c ON p.category_id = c.id
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, models.PaginationMeta{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var plugin internalmodels.ExtendedPlugin
		var categoryName, categoryIcon, categoryColor sql.NullString
		var interfaceSpec, dependencies, runtimeConfig, securityPolicy, resourceLimits sql.NullString
		var entryPoint, exportedSymbols, binaryPath sql.NullString
		var dockerImage, healthCheckPath, grpcProtoFile, serviceEndpoint, namespace sql.NullString
		var servicePort sql.NullInt32

		err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.ShortDescription,
			&plugin.Author, &plugin.DeveloperID, &plugin.Version, &plugin.IconURL,
			&plugin.BannerURL, &plugin.Price, &plugin.Rating, &plugin.DownloadCount,
			&plugin.IsFeatured, &plugin.CreatedAt, &plugin.UpdatedAt, &plugin.CategoryID,
			&plugin.Type, &plugin.Runtime, &plugin.Status, &interfaceSpec, &dependencies,
			&runtimeConfig, &securityPolicy, &resourceLimits,
			&entryPoint, &exportedSymbols, &binaryPath,
			&dockerImage, &servicePort, &healthCheckPath, &grpcProtoFile,
			&serviceEndpoint, &namespace,
			&categoryName, &categoryIcon, &categoryColor,
		)
		if err != nil {
			return nil, models.PaginationMeta{}, err
		}

		// 处理JSON字段
		if interfaceSpec.Valid {
			plugin.InterfaceSpec = &interfaceSpec.String
		}
		if dependencies.Valid {
			var deps []string
			if err := json.Unmarshal([]byte(dependencies.String), &deps); err == nil {
				plugin.Dependencies = deps
			}
		}
		if runtimeConfig.Valid {
			plugin.RuntimeConfig = &runtimeConfig.String
		}
		if securityPolicy.Valid {
			plugin.SecurityPolicy = &securityPolicy.String
		}
		if resourceLimits.Valid {
			plugin.ResourceLimits = &resourceLimits.String
		}

		// 处理进程内插件字段
		if entryPoint.Valid {
			plugin.EntryPoint = &entryPoint.String
		}
		if exportedSymbols.Valid {
			var symbols []string
			if err := json.Unmarshal([]byte(exportedSymbols.String), &symbols); err == nil {
				plugin.ExportedSymbols = symbols
			}
		}
		if binaryPath.Valid {
			plugin.BinaryPath = &binaryPath.String
		}

		// 处理微服务插件字段
		if dockerImage.Valid {
			plugin.DockerImage = &dockerImage.String
		}
		if servicePort.Valid {
			port := int(servicePort.Int32)
			plugin.ServicePort = &port
		}
		if healthCheckPath.Valid {
			plugin.HealthCheckPath = &healthCheckPath.String
		}
		if grpcProtoFile.Valid {
			plugin.GRPCProtoFile = &grpcProtoFile.String
		}
		if serviceEndpoint.Valid {
			plugin.ServiceEndpoint = &serviceEndpoint.String
		}
		if namespace.Valid {
			plugin.Namespace = &namespace.String
		}

		// 设置分类信息
		if categoryName.Valid && plugin.CategoryID != nil {
			var icon *string
			if categoryIcon.Valid {
				icon = &categoryIcon.String
			}
			plugin.Category = &models.Category{
				ID:    *plugin.CategoryID,
				Name:  categoryName.String,
				Icon:  icon,
				Color: categoryColor.String,
			}
		}

		plugins = append(plugins, plugin)
	}

	// 计算分页信息
	totalPages := (totalCount + params.PageSize - 1) / params.PageSize

	meta := models.PaginationMeta{
		Page:       params.Page,
		Limit:      params.PageSize,
		Total:      totalCount,
		TotalPages: totalPages,
	}

	return plugins, meta, nil
}

// GetExtendedPlugin 获取单个扩展插件详情
func (s *ExtendedPluginService) GetExtendedPlugin(id uuid.UUID) (*internalmodels.ExtendedPlugin, error) {
	var plugin internalmodels.ExtendedPlugin
	var categoryName, categoryIcon, categoryColor sql.NullString
	var interfaceSpec, dependencies, runtimeConfig, securityPolicy, resourceLimits sql.NullString
	var entryPoint, exportedSymbols, binaryPath sql.NullString
	var dockerImage, healthCheckPath, grpcProtoFile, serviceEndpoint, namespace sql.NullString
	var servicePort sql.NullInt32

	query := `
		SELECT 
			p.id, p.name, p.description, p.short_description, p.author, p.developer_id, 
			p.version, p.icon_url, p.banner_url, p.price, p.rating, p.download_count, 
			p.is_featured, p.created_at, p.updated_at, p.category_id,
			p.type, p.runtime, p.status, p.interface_spec, p.dependencies, 
			p.runtime_config, p.security_policy, p.resource_limits,
			p.entry_point, p.exported_symbols, p.binary_path,
			p.docker_image, p.service_port, p.health_check_path, 
			p.grpc_proto_file, p.service_endpoint, p.namespace,
			c.name as category_name, c.icon as category_icon, c.color as category_color
		FROM mp_plugins p
		LEFT JOIN mp_categories c ON p.category_id = c.id
		WHERE p.id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&plugin.ID, &plugin.Name, &plugin.Description, &plugin.ShortDescription,
		&plugin.Author, &plugin.DeveloperID, &plugin.Version, &plugin.IconURL,
		&plugin.BannerURL, &plugin.Price, &plugin.Rating, &plugin.DownloadCount,
		&plugin.IsFeatured, &plugin.CreatedAt, &plugin.UpdatedAt, &plugin.CategoryID,
		&plugin.Type, &plugin.Runtime, &plugin.Status, &interfaceSpec, &dependencies,
		&runtimeConfig, &securityPolicy, &resourceLimits,
		&entryPoint, &exportedSymbols, &binaryPath,
		&dockerImage, &servicePort, &healthCheckPath, &grpcProtoFile,
		&serviceEndpoint, &namespace,
		&categoryName, &categoryIcon, &categoryColor,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// 处理JSON字段（与上面相同的逻辑）
	if interfaceSpec.Valid {
		plugin.InterfaceSpec = &interfaceSpec.String
	}
	if dependencies.Valid {
		var deps []string
		if err := json.Unmarshal([]byte(dependencies.String), &deps); err == nil {
			plugin.Dependencies = deps
		}
	}
	if runtimeConfig.Valid {
		plugin.RuntimeConfig = &runtimeConfig.String
	}
	if securityPolicy.Valid {
		plugin.SecurityPolicy = &securityPolicy.String
	}
	if resourceLimits.Valid {
		plugin.ResourceLimits = &resourceLimits.String
	}

	// 处理进程内插件字段
	if entryPoint.Valid {
		plugin.EntryPoint = &entryPoint.String
	}
	if exportedSymbols.Valid {
		var symbols []string
		if err := json.Unmarshal([]byte(exportedSymbols.String), &symbols); err == nil {
			plugin.ExportedSymbols = symbols
		}
	}
	if binaryPath.Valid {
		plugin.BinaryPath = &binaryPath.String
	}

	// 处理微服务插件字段
	if dockerImage.Valid {
		plugin.DockerImage = &dockerImage.String
	}
	if servicePort.Valid {
		port := int(servicePort.Int32)
		plugin.ServicePort = &port
	}
	if healthCheckPath.Valid {
		plugin.HealthCheckPath = &healthCheckPath.String
	}
	if grpcProtoFile.Valid {
		plugin.GRPCProtoFile = &grpcProtoFile.String
	}
	if serviceEndpoint.Valid {
		plugin.ServiceEndpoint = &serviceEndpoint.String
	}
	if namespace.Valid {
		plugin.Namespace = &namespace.String
	}

	// 设置分类信息
	if categoryName.Valid && plugin.CategoryID != nil {
		var icon *string
		if categoryIcon.Valid {
			icon = &categoryIcon.String
		}
		plugin.Category = &models.Category{
			ID:    *plugin.CategoryID,
			Name:  categoryName.String,
			Icon:  icon,
			Color: categoryColor.String,
		}
	}

	return &plugin, nil
}

// DeployPlugin 部署插件
func (s *ExtendedPluginService) DeployPlugin(ctx context.Context, pluginID uuid.UUID) error {
	plugin, err := s.GetExtendedPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	if plugin == nil {
		return fmt.Errorf("plugin not found")
	}

	// 更新插件状态为部署中
	if err := s.UpdatePluginStatus(pluginID, internalmodels.StatusDeploying); err != nil {
		return fmt.Errorf("failed to update plugin status: %w", err)
	}

	s.logger.Infof("Deploying plugin %s (type: %s, runtime: %s)", plugin.Name, plugin.Type, plugin.Runtime)

	switch plugin.Type {
	case internalmodels.PluginTypeInProcess:
		err = s.deployInProcessPlugin(ctx, plugin)
	case internalmodels.PluginTypeMicroservice:
		err = s.deployMicroservicePlugin(ctx, plugin)
	default:
		err = fmt.Errorf("unsupported plugin type: %s", plugin.Type)
	}

	if err != nil {
		// 部署失败，更新状态为错误
		s.UpdatePluginStatus(pluginID, internalmodels.StatusError)
		return fmt.Errorf("failed to deploy plugin: %w", err)
	}

	// 部署成功，更新状态为运行中
	if err := s.UpdatePluginStatus(pluginID, internalmodels.StatusRunning); err != nil {
		s.logger.Warnf("Failed to update plugin status to running: %v", err)
	}

	s.logger.Infof("Successfully deployed plugin %s", plugin.Name)
	return nil
}

// GetPluginCallLogs 获取插件调用日志
func (s *ExtendedPluginService) GetPluginCallLogs(pluginID *uuid.UUID, userID *uuid.UUID, success *bool, startTime, endTime *time.Time, page, limit int) ([]internalmodels.PluginCallLog, int, error) {
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// 构建查询条件
	if pluginID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("plugin_id = $%d", argIndex))
		args = append(args, *pluginID)
		argIndex++
	}

	if userID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *userID)
		argIndex++
	}

	if success != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("success = $%d", argIndex))
		args = append(args, *success)
		argIndex++
	}

	if startTime != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *startTime)
		argIndex++
	}

	if endTime != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *endTime)
		argIndex++
	}

	whereClause := "1=1"
	if len(whereConditions) > 0 {
		whereClause = strings.Join(whereConditions, " AND ")
	}

	// 获取总数
	var totalCount int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM mp_plugin_call_logs WHERE %s", whereClause)
	err := s.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// 获取日志列表
	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT id, plugin_id, user_id, method, input_data, output_data, 
			   duration, success, error_msg, client_type, client_ip, user_agent, created_at
		FROM mp_plugin_call_logs 
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []internalmodels.PluginCallLog
	for rows.Next() {
		var log internalmodels.PluginCallLog
		var inputData, outputData sql.NullString

		err := rows.Scan(
			&log.ID, &log.PluginID, &log.UserID, &log.Method,
			&inputData, &outputData, &log.Duration, &log.Success,
			&log.ErrorMsg, &log.ClientType, &log.ClientIP, &log.UserAgent, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// 处理JSON字段
		if inputData.Valid {
			if err := json.Unmarshal([]byte(inputData.String), &log.InputData); err != nil {
				log.InputData = nil
			}
		}

		if outputData.Valid {
			if err := json.Unmarshal([]byte(outputData.String), &log.OutputData); err != nil {
				log.OutputData = nil
			}
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return logs, totalCount, nil
}

// deployInProcessPlugin 部署进程内插件
func (s *ExtendedPluginService) deployInProcessPlugin(ctx context.Context, plugin *internalmodels.ExtendedPlugin) error {
	config := make(map[string]interface{})
	config["id"] = plugin.ID.String()
	config["name"] = plugin.Name

	switch plugin.Runtime {
	case internalmodels.RuntimeGo:
		if plugin.BinaryPath == nil {
			return fmt.Errorf("binary path is required for Go plugins")
		}
		// TODO: 实现Go插件加载逻辑
		return fmt.Errorf("Go plugin loading not implemented yet")
	case internalmodels.RuntimeJS:
		if plugin.BinaryPath == nil {
			return fmt.Errorf("module path is required for JS plugins")
		}
		// TODO: 实现JS插件加载逻辑
		return fmt.Errorf("JS plugin loading not implemented yet")
	default:
		return fmt.Errorf("unsupported runtime for in-process plugin: %s", plugin.Runtime)
	}
}

// deployMicroservicePlugin 部署微服务插件
func (s *ExtendedPluginService) deployMicroservicePlugin(ctx context.Context, plugin *internalmodels.ExtendedPlugin) error {
	if plugin.DockerImage == nil {
		return fmt.Errorf("docker image is required for microservice plugins")
	}

	// TODO: 实现微服务插件部署逻辑
	return fmt.Errorf("microservice plugin deployment not implemented")
}

// StopPlugin 停止插件
func (s *ExtendedPluginService) StopPlugin(ctx context.Context, pluginID uuid.UUID) error {
	plugin, err := s.GetExtendedPlugin(pluginID)
	if err != nil {
		return fmt.Errorf("failed to get plugin: %w", err)
	}

	if plugin == nil {
		return fmt.Errorf("plugin not found")
	}

	s.logger.Infof("Stopping plugin %s", plugin.Name)

	switch plugin.Type {
	case internalmodels.PluginTypeInProcess:
		err = s.inProcessManager.UnloadPlugin(plugin.ID.String())
	case internalmodels.PluginTypeMicroservice:
		err = s.microserviceManager.StopPlugin(plugin.ID.String())
	default:
		err = fmt.Errorf("unsupported plugin type: %s", plugin.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	// 更新插件状态为已停止
	if err := s.UpdatePluginStatus(pluginID, internalmodels.StatusStopped); err != nil {
		s.logger.Warnf("Failed to update plugin status to stopped: %v", err)
	}

	s.logger.Infof("Successfully stopped plugin %s", plugin.Name)
	return nil
}

// RestartPlugin 重启插件
func (s *ExtendedPluginService) RestartPlugin(ctx context.Context, pluginID uuid.UUID) error {
	if err := s.StopPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	time.Sleep(2 * time.Second) // 等待停止完成

	return s.DeployPlugin(ctx, pluginID)
}

// CallPlugin 调用插件
func (s *ExtendedPluginService) CallPlugin(ctx context.Context, request *internalmodels.PluginCallRequest) (*internalmodels.PluginCallResponse, error) {
	// 解析 PluginID
	pluginUUID, err := uuid.Parse(request.PluginID)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin ID format: %w", err)
	}

	// 序列化输入参数
	inputJSON, _ := json.Marshal(request.Params)
	inputDataStr := string(inputJSON)

	// 记录调用日志
	callLog := &internalmodels.PluginCallLog{
		ID:         uuid.New(),
		PluginID:   pluginUUID,
		UserID:     nil, // 从context中获取用户ID
		Method:     request.Method,
		InputData:  &inputDataStr,
		ClientType: "api",
		CreatedAt:  time.Now(),
	}

	startTime := time.Now()

	// 构建网关请求
	gatewayRequest := &plugin.GatewayRequest{
		PluginID:   request.PluginID,
		Method:     request.Method,
		Params:     request.Params,
		ClientType: plugin.ClientTypeAPI,
		RequestID:  uuid.New().String(),
	}

	// 通过网关调用插件
	response, err := s.callPluginViaGateway(ctx, gatewayRequest)

	duration := time.Since(startTime).Milliseconds()
	callLog.Duration = duration
	callLog.Success = err == nil

	if err != nil {
		errorMsg := err.Error()
		callLog.ErrorMsg = &errorMsg
	} else if response != nil {
		outputJSON, _ := json.Marshal(response.Data)
		outputDataStr := string(outputJSON)
		callLog.OutputData = &outputDataStr
	}

	// 异步记录调用日志
	go s.recordCallLog(callLog)

	return response, err
}

// callPluginViaGateway 通过网关调用插件
func (s *ExtendedPluginService) callPluginViaGateway(ctx context.Context, request *plugin.GatewayRequest) (*internalmodels.PluginCallResponse, error) {
	// 确定插件类型
	pluginType, err := s.determinePluginType(request.PluginID)
	if err != nil {
		return nil, err
	}

	var result *plugin.PluginResult

	// 根据插件类型调用
	switch pluginType {
	case "in_process":
		result, err = s.callInProcessPlugin(ctx, request)
	case "microservice":
		result, err = s.callMicroservicePlugin(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported plugin type: %s", pluginType)
	}

	if err != nil {
		return nil, err
	}

	// 转换为内部响应格式
	response := &internalmodels.PluginCallResponse{
		Success:  result.Success,
		Data:     result.Data,
		Duration: result.Duration,
		Metadata: result.Metadata,
	}

	if !result.Success {
		response.Error = &result.Error
	}

	return response, nil
}

// determinePluginType 确定插件类型
func (s *ExtendedPluginService) determinePluginType(pluginID string) (string, error) {
	// 首先尝试从进程内插件管理器获取插件实例
	if _, exists := s.inProcessManager.GetPlugin(pluginID); exists {
		return "in_process", nil
	}

	// 然后尝试从微服务插件管理器获取插件实例
	if _, exists := s.microserviceManager.GetPlugin(pluginID); exists {
		return "microservice", nil
	}

	return "", fmt.Errorf("plugin not found: %s", pluginID)
}

// callInProcessPlugin 调用进程内插件
func (s *ExtendedPluginService) callInProcessPlugin(ctx context.Context, request *plugin.GatewayRequest) (*plugin.PluginResult, error) {
	// 获取插件实例
	_, exists := s.inProcessManager.GetPlugin(request.PluginID)
	if !exists {
		return nil, fmt.Errorf("failed to get in-process plugin: plugin not found")
	}

	// TODO: 实现进程内插件调用逻辑
	// 根据插件接口类型调用相应方法
	return nil, fmt.Errorf("in-process plugin call not implemented")
}

// callMicroservicePlugin 调用微服务插件
func (s *ExtendedPluginService) callMicroservicePlugin(ctx context.Context, request *plugin.GatewayRequest) (*plugin.PluginResult, error) {
	// TODO: 实现微服务插件调用逻辑
	return nil, fmt.Errorf("microservice plugin call not implemented")
}

// recordCallLog 记录调用日志
func (s *ExtendedPluginService) recordCallLog(log *internalmodels.PluginCallLog) {
	query := `
		INSERT INTO mp_plugin_call_logs (
			id, plugin_id, user_id, method, input_data, output_data, 
			duration, success, error_msg, client_type, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := s.db.Exec(query,
		log.ID, log.PluginID, log.UserID, log.Method, log.InputData, log.OutputData,
		log.Duration, log.Success, log.ErrorMsg, log.ClientType, log.CreatedAt,
	)

	if err != nil {
		s.logger.Errorf("Failed to record call log: %v", err)
	}
}

// UpdatePluginStatus 更新插件状态
func (s *ExtendedPluginService) UpdatePluginStatus(pluginID uuid.UUID, status internalmodels.PluginStatus) error {
	query := "UPDATE mp_plugins SET status = $2, updated_at = NOW() WHERE id = $1"
	result, err := s.db.Exec(query, pluginID, status)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("plugin not found")
	}

	return nil
}

// GetPluginMetrics 获取插件指标
func (s *ExtendedPluginService) GetPluginMetrics(pluginID uuid.UUID, days int) (*internalmodels.PluginMetrics, error) {
	if days <= 0 {
		days = 7 // 默认7天
	}

	query := `
		SELECT 
			COALESCE(SUM(call_count), 0) as call_count,
			COALESCE(SUM(success_count), 0) as success_count,
			COALESCE(SUM(error_count), 0) as error_count,
			COALESCE(AVG(avg_duration), 0) as avg_duration,
			COALESCE(MAX(max_duration), 0) as max_duration,
			COALESCE(MIN(min_duration), 0) as min_duration,
			COALESCE(SUM(unique_users), 0) as unique_users
		FROM mp_plugin_metrics 
		WHERE plugin_id = $1 AND date >= CURRENT_DATE - INTERVAL '%d days'`

	var metrics internalmodels.PluginMetrics
	err := s.db.QueryRow(fmt.Sprintf(query, days), pluginID).Scan(
		&metrics.CallCount, &metrics.SuccessCount, &metrics.ErrorCount,
		&metrics.AvgDuration, &metrics.MaxDuration, &metrics.MinDuration,
		&metrics.UniqueUsers,
	)

	if err != nil {
		return nil, err
	}

	metrics.PluginID = pluginID
	// 设置当前时间作为查询日期
	metrics.Date = time.Now()

	return &metrics, nil
}

// CreateExtendedPlugin 创建扩展插件
func (s *ExtendedPluginService) CreateExtendedPlugin(plugin *internalmodels.ExtendedPlugin) error {
	plugin.ID = uuid.New()
	plugin.CreatedAt = time.Now()
	plugin.UpdatedAt = time.Now()

	// 序列化JSON字段
	dependenciesJSON, _ := json.Marshal(plugin.Dependencies)
	exportedSymbolsJSON, _ := json.Marshal(plugin.ExportedSymbols)

	query := `
		INSERT INTO mp_plugins (
			id, name, description, short_description, author, version, 
			category_id, price, is_free, is_featured, is_active,
			icon_url, banner_url, screenshots, tags, requirements, changelog,
			type, runtime, status, interface_spec, dependencies, 
			runtime_config, security_policy, resource_limits,
			entry_point, exported_symbols, binary_path,
			docker_image, service_port, health_check_path, 
			grpc_proto_file, service_endpoint, namespace,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
			$18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36
		)`

	_, err := s.db.Exec(query,
		plugin.ID, plugin.Name, plugin.Description, plugin.ShortDescription,
		plugin.Author, plugin.Version, plugin.CategoryID, plugin.Price,
		plugin.IsFree, plugin.IsFeatured, plugin.IsActive,
		plugin.IconURL, plugin.BannerURL, plugin.Screenshots, plugin.Tags,
		plugin.Requirements, plugin.Changelog,
		plugin.Type, plugin.Runtime, plugin.Status, plugin.InterfaceSpec, string(dependenciesJSON),
		plugin.RuntimeConfig, plugin.SecurityPolicy, plugin.ResourceLimits,
		plugin.EntryPoint, string(exportedSymbolsJSON), plugin.BinaryPath,
		plugin.DockerImage, plugin.ServicePort, plugin.HealthCheckPath,
		plugin.GRPCProtoFile, plugin.ServiceEndpoint, plugin.Namespace,
		plugin.CreatedAt, plugin.UpdatedAt,
	)

	return err
}

// UpdateExtendedPlugin 更新扩展插件
func (s *ExtendedPluginService) UpdateExtendedPlugin(id uuid.UUID, plugin *internalmodels.ExtendedPlugin) error {
	plugin.UpdatedAt = time.Now()

	// 序列化JSON字段
	dependenciesJSON, _ := json.Marshal(plugin.Dependencies)
	exportedSymbolsJSON, _ := json.Marshal(plugin.ExportedSymbols)

	query := `
		UPDATE mp_plugins SET 
			name = $2, description = $3, short_description = $4, author = $5,
			version = $6, category_id = $7, price = $8, is_free = $9,
			is_featured = $10, is_active = $11, icon_url = $12, banner_url = $13,
			screenshots = $14, tags = $15, requirements = $16, changelog = $17,
			type = $18, runtime = $19, status = $20, interface_spec = $21, dependencies = $22,
			runtime_config = $23, security_policy = $24, resource_limits = $25,
			entry_point = $26, exported_symbols = $27, binary_path = $28,
			docker_image = $29, service_port = $30, health_check_path = $31,
			grpc_proto_file = $32, service_endpoint = $33, namespace = $34,
			updated_at = $35
		WHERE id = $1`

	result, err := s.db.Exec(query,
		id, plugin.Name, plugin.Description, plugin.ShortDescription,
		plugin.Author, plugin.Version, plugin.CategoryID, plugin.Price,
		plugin.IsFree, plugin.IsFeatured, plugin.IsActive,
		plugin.IconURL, plugin.BannerURL, plugin.Screenshots, plugin.Tags,
		plugin.Requirements, plugin.Changelog,
		plugin.Type, plugin.Runtime, plugin.Status, plugin.InterfaceSpec, string(dependenciesJSON),
		plugin.RuntimeConfig, plugin.SecurityPolicy, plugin.ResourceLimits,
		plugin.EntryPoint, string(exportedSymbolsJSON), plugin.BinaryPath,
		plugin.DockerImage, plugin.ServicePort, plugin.HealthCheckPath,
		plugin.GRPCProtoFile, plugin.ServiceEndpoint, plugin.Namespace,
		plugin.UpdatedAt,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("plugin not found")
	}

	return nil
}
