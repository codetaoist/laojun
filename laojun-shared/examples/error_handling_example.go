package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/codetaoist/laojun-shared/observability"
)

func main() {
	fmt.Println("=== 错误处理和边界情况测试 ===")

	// 创建默认配置
	config := observability.DefaultConfig()
	config.ServiceName = "error-handling-test"
	config.ServiceVersion = "1.0.0"
	config.Environment = "test"
	
	// 启用监控和追踪
	config.EnableMonitoring()
	config.EnableTracing()

	// 初始化可观测性
	obs, err := observability.NewObservability(config)
	if err != nil {
		log.Fatalf("初始化可观测性失败: %v", err)
	}
	defer obs.Close()

	testNilParameters(obs)
	testEmptyParameters(obs)
	testFinishedOperations(obs)
	testClosedObservability(obs)
	testConcurrentAccess(obs)

	fmt.Println("\n=== 错误处理测试完成! ===")
}

func testNilParameters(obs observability.Observability) {
	fmt.Println("\n--- 测试 nil 参数处理 ---")

	// 测试 nil context
	op1, _ := obs.StartOperation(nil, "test-nil-context")
	op1.SetAttribute("test", "value")
	op1.Finish()
	fmt.Println("✓ nil context 处理正常")

	// 测试空操作名
	op2, _ := obs.StartOperation(context.Background(), "")
	op2.SetAttribute("test", "value")
	op2.Finish()
	fmt.Println("✓ 空操作名处理正常")

	// 测试 nil 选项
	op3, _ := obs.StartOperation(context.Background(), "test-nil-options", nil)
	op3.SetAttribute("test", "value")
	op3.Finish()
	fmt.Println("✓ nil 选项处理正常")
}

func testEmptyParameters(obs observability.Observability) {
	fmt.Println("\n--- 测试空参数处理 ---")

	op, _ := obs.StartOperation(context.Background(), "test-empty-params")

	// 测试空属性键
	op.SetAttribute("", "value")
	fmt.Println("✓ 空属性键处理正常")

	// 测试空事件名
	op.AddEvent("")
	fmt.Println("✓ 空事件名处理正常")

	// 测试 nil 错误
	op.SetError(nil)
	fmt.Println("✓ nil 错误处理正常")

	// 测试空指标名
	op.IncrementCounter("", 1.0)
	op.SetGauge("", 1.0)
	op.RecordHistogram("", 1.0)
	op.RecordSummary("", 1.0)
	fmt.Println("✓ 空指标名处理正常")

	op.Finish()
}

func testFinishedOperations(obs observability.Observability) {
	fmt.Println("\n--- 测试已完成操作的处理 ---")

	op, _ := obs.StartOperation(context.Background(), "test-finished-operation")
	
	// 先完成操作
	op.Finish()
	fmt.Println("✓ 操作已完成")

	// 尝试在已完成的操作上调用方法
	op.SetAttribute("after-finish", "value")
	op.SetStatus(observability.OperationStatusRunning)
	op.AddEvent("after-finish-event")
	op.SetError(errors.New("after finish error"))
	op.IncrementCounter("after_finish_counter", 1.0)
	op.SetGauge("after_finish_gauge", 1.0)
	
	// 尝试创建子操作
	child := op.StartChild("child-after-finish")
	child.Finish()
	fmt.Println("✓ 已完成操作的子操作处理正常")

	fmt.Println("✓ 已完成操作的方法调用处理正常")
}

func testClosedObservability(obs observability.Observability) {
	fmt.Println("\n--- 测试关闭后的可观测性实例 ---")

	// 创建一个新的实例用于测试关闭
	config := observability.DefaultConfig()
	config.ServiceName = "test-closed"
	
	testObs, err := observability.NewObservability(config)
	if err != nil {
		fmt.Printf("创建测试实例失败: %v\n", err)
		return
	}

	// 关闭实例
	testObs.Close()

	// 尝试在关闭后创建操作
	op, _ := testObs.StartOperation(context.Background(), "after-close-operation")
	op.Finish()
	fmt.Println("✓ 关闭后的操作创建处理正常")
}

func testConcurrentAccess(obs observability.Observability) {
	fmt.Println("\n--- 测试并发访问 ---")

	op, _ := obs.StartOperation(context.Background(), "test-concurrent-access")

	// 启动多个goroutine并发访问操作
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// 并发设置属性
			op.SetAttribute(fmt.Sprintf("attr_%d", id), fmt.Sprintf("value_%d", id))
			
			// 并发添加事件
			op.AddEvent(fmt.Sprintf("event_%d", id))
			
			// 并发记录指标
			op.IncrementCounter(fmt.Sprintf("counter_%d", id), 1.0)
			op.SetGauge(fmt.Sprintf("gauge_%d", id), float64(id))
			
			// 并发创建子操作
			child := op.StartChild(fmt.Sprintf("child_%d", id))
			child.SetAttribute("child_attr", "child_value")
			child.Finish()
			
			time.Sleep(10 * time.Millisecond)
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	op.Finish()
	fmt.Println("✓ 并发访问处理正常")
}