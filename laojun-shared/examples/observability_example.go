package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/codetaoist/laojun-shared/observability"
)

func main() {
	// 创建可观测性配置
	config := observability.DefaultConfig()
	config.WithServiceInfo("observability-example", "1.0.0", "development").
		EnableMonitoring().
		EnableTracing().
		WithExportFormats("prometheus", "jaeger")

	// 创建可观测性实例
	obs, err := observability.NewObservability(config)
	if err != nil {
		log.Fatalf("创建可观测性实例失败: %v", err)
	}
	defer obs.Close()

	ctx := context.Background()

	// 示例 1: 基础操作
	fmt.Println("=== 示例 1: 基础操作 ===")
	op1, _ := obs.StartOperation(ctx, "basic-operation")
	op1.SetAttribute("example", "basic")
	op1.IncrementCounter("operations_count", 1)
	
	// 模拟工作
	time.Sleep(100 * time.Millisecond)
	
	op1.SetStatus(observability.OperationStatusSuccess)
	op1.Finish()
	
	fmt.Printf("基础操作完成，耗时: %v\n", op1.Duration())

	// 示例 2: HTTP操作
	fmt.Println("\n=== 示例 2: HTTP操作 ===")
	op2, _ := obs.StartOperation(ctx, "http-request", 
		observability.WithOperationType(observability.OperationTypeHTTP),
		observability.WithAttribute("http.method", "GET"),
		observability.WithAttribute("http.url", "/api/users"))
	
	op2.SetAttribute("http.method", "GET")
	op2.SetAttribute("http.url", "/api/users")
	op2.SetAttribute("http.status_code", 200)
	op2.AddEvent("request_started", observability.EventAttribute{Key: "timestamp", Value: time.Now()})
	
	// 模拟HTTP处理
	time.Sleep(50 * time.Millisecond)
	
	op2.RecordHistogram("http_request_duration_seconds", 0.05)
	op2.SetStatus(observability.OperationStatusSuccess)
	op2.Finish()
	
	fmt.Printf("HTTP操作完成，耗时: %v\n", op2.Duration())

	// 示例 3: 数据库操作与嵌套操作
	fmt.Println("\n=== 示例 3: 数据库操作 ===")
	op3, _ := obs.StartOperation(ctx, "database-query", 
		observability.WithOperationType(observability.OperationTypeDatabase),
		observability.WithAttribute("db.operation", "SELECT"),
		observability.WithAttribute("db.table", "users"))
	
	op3.SetAttribute("db.statement", "SELECT * FROM users WHERE id = ?")
	op3.SetAttribute("db.connection_string", "postgres://localhost:5432/mydb")
	
	// 嵌套操作: 获取连接
	connOp := op3.StartChild("db-connection-acquire")
	connOp.SetAttribute("db.operation", "connection_acquire")
	time.Sleep(10 * time.Millisecond)
	connOp.SetStatus(observability.OperationStatusSuccess)
	connOp.Finish()
	
	// 嵌套操作: 执行查询
	queryOp := op3.StartChild("db-query-execute")
	queryOp.SetAttribute("db.operation", "query_execute")
	queryOp.AddEvent("query_start", 
		observability.EventAttribute{Key: "query", Value: "SELECT * FROM users WHERE id = ?"},
		observability.EventAttribute{Key: "params", Value: []interface{}{12345}},
	)
	
	// 模拟查询执行
	time.Sleep(30 * time.Millisecond)
	
	queryOp.RecordSummary("db_query_rows_returned", 1)
	queryOp.SetStatus(observability.OperationStatusSuccess)
	queryOp.Finish()
	
	op3.SetStatus(observability.OperationStatusSuccess)
	op3.Finish()
	
	fmt.Printf("数据库操作完成，耗时: %v\n", op3.Duration())

	// 示例 4: 错误处理
	fmt.Println("\n=== 示例 4: 错误处理 ===")
	errorOp, _ := obs.StartOperation(ctx, "error-operation")
	errorOp.SetAttribute("operation.type", "error_simulation")
	
	// 模拟错误
	simulatedError := fmt.Errorf("连接超时，等待5秒后失败")
	errorOp.SetError(simulatedError)
	errorOp.IncrementCounter("errors_total", 1)
	
	errorOp.Finish()
	
	fmt.Printf("错误操作完成，耗时: %v\n", errorOp.Duration())

	// 示例 5: 缓存操作
	fmt.Println("\n=== 示例 5: 缓存操作 ===")
	cacheOp, _ := obs.StartOperation(ctx, "cache-get", 
		observability.WithOperationType(observability.OperationTypeCache),
		observability.WithAttribute("cache.operation", "GET"),
		observability.WithAttribute("cache.key", "user:12345"))
	
	cacheOp.SetAttribute("cache.key", "user:12345")
	cacheOp.SetAttribute("cache.hit", false)
	cacheOp.AddEvent("cache_miss", observability.EventAttribute{Key: "key", Value: "user:12345"})
	
	time.Sleep(5 * time.Millisecond)
	
	cacheOp.SetGauge("cache_hit_ratio", 0.85)
	cacheOp.SetStatus(observability.OperationStatusSuccess)
	cacheOp.Finish()
	
	fmt.Printf("缓存操作完成，耗时: %v\n", cacheOp.Duration())

	// 示例 6: 外部服务操作
	fmt.Println("\n=== 示例 6: 外部服务操作 ===")
	extOp, _ := obs.StartOperation(ctx, "external-service-call", 
		observability.WithOperationType(observability.OperationTypeExternal),
		observability.WithAttribute("service.name", "payment-service"),
		observability.WithAttribute("service.operation", "process_payment"))
	
	extOp.SetAttribute("service.name", "payment-service")
	extOp.SetAttribute("service.version", "2.1.0")
	extOp.SetAttribute("request.id", "req-789")
	
	// 模拟外部服务调用
	time.Sleep(200 * time.Millisecond)
	
	extOp.RecordHistogram("external_service_response_time", 0.2)
	extOp.SetStatus(observability.OperationStatusSuccess)
	extOp.Finish()
	
	fmt.Printf("外部服务操作完成，耗时: %v\n", extOp.Duration())

	// 示例 7: 复杂操作，包含多个指标和事件
	fmt.Println("\n=== 示例 7: 复杂操作 ===")
	complexOp, _ := obs.StartOperation(ctx, "complex-operation",
		observability.WithAttributes(map[string]interface{}{
			"component": "business_logic",
			"version": "1.2.3",
		}),
		observability.WithTimeout(5*time.Second),
	)
	
	// 操作的多个阶段
	phases := []string{"validation", "processing", "persistence", "notification"}
	
	for i, phase := range phases {
		phaseOp := complexOp.StartChild(fmt.Sprintf("phase-%s", phase))
		phaseOp.SetAttribute("phase.name", phase)
		phaseOp.SetAttribute("phase.index", i)
		
		// 模拟阶段工作
		phaseTime := time.Duration(20+i*10) * time.Millisecond
		time.Sleep(phaseTime)
		
		phaseOp.RecordHistogram("phase_duration_seconds", phaseTime.Seconds())
		phaseOp.AddEvent(fmt.Sprintf("phase_%s_completed", phase))
		phaseOp.SetStatus(observability.OperationStatusSuccess)
		phaseOp.Finish()
		
		fmt.Printf("  阶段 %s 完成，耗时: %v\n", phase, phaseOp.Duration())
	}
	
	complexOp.SetStatus(observability.OperationStatusSuccess)
	complexOp.FinishWithOptions(observability.FinishOptions{
		ExtraAttributes: map[string]interface{}{
			"total_phases": len(phases),
			"success_rate": 1.0,
			"phases_completed": len(phases),
			"final_status": "success",
		},
	})
	
	fmt.Printf("复杂操作完成，耗时: %v\n", complexOp.Duration())

	// 示例 8: 健康检查和导出
	fmt.Println("\n=== 示例 8: 健康检查和导出 ===")
	
	// 检查健康状态
	healthStatus := obs.HealthCheck()
	if healthStatus.Status == "healthy" {
		fmt.Println("可观测性系统健康状态良好")
	} else {
		fmt.Printf("可观测性系统健康状态异常: %s\n", healthStatus.Status)
	}
	
	// 导出数据
	err = obs.Export(ctx, "prometheus")
	if err != nil {
		log.Printf("导出可观测性数据失败: %v", err)
	} else {
		fmt.Println("成功导出可观测性数据")
	}

	fmt.Println("\n=== 可观测性示例成功完成! ===")
}