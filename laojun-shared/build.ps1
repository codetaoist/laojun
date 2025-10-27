# PowerShell build script for laojun-shared library

param(
    [Parameter(Position=0)]
    [string]$Command = "help",
    
    [Parameter()]
    [string]$Package = ""
)

function Show-Help {
    Write-Host "可用的命令:" -ForegroundColor Green
    Write-Host "  build         - 构建所有工具" -ForegroundColor Yellow
    Write-Host "  test          - 运行所有测试" -ForegroundColor Yellow
    Write-Host "  check-api     - 检查API规范" -ForegroundColor Yellow
    Write-Host "  examples      - 运行所有示例" -ForegroundColor Yellow
    Write-Host "  codegen       - 生成新模块代码模板 (需要 -Package 参数)" -ForegroundColor Yellow
    Write-Host "  clean         - 清理构建文件" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "示例:" -ForegroundColor Green
    Write-Host "  .\build.ps1 build" -ForegroundColor Cyan
    Write-Host "  .\build.ps1 codegen -Package mypackage" -ForegroundColor Cyan
}

function Build-Tools {
    Write-Host "🔨 构建代码生成工具..." -ForegroundColor Blue
    if (!(Test-Path "bin")) {
        New-Item -ItemType Directory -Path "bin" | Out-Null
    }
    
    go build -o bin/codegen.exe ./tools/codegen
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 构建代码生成工具失败" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "🔨 构建API检查工具..." -ForegroundColor Blue
    go build -o bin/linter.exe ./tools/linter
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 构建API检查工具失败" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ 构建完成" -ForegroundColor Green
}

function Run-Tests {
    Write-Host "🧪 运行单元测试..." -ForegroundColor Blue
    go test ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 单元测试失败" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "🧪 运行集成测试..." -ForegroundColor Blue
    go test ./test/...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 集成测试失败" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ 测试完成" -ForegroundColor Green
}

function Check-API {
    if (!(Test-Path "bin/linter.exe")) {
        Write-Host "🔨 构建API检查工具..." -ForegroundColor Blue
        Build-Tools
    }
    
    Write-Host "📋 检查API规范..." -ForegroundColor Blue
    & .\bin\linter.exe -dir .
    Write-Host "✅ API规范检查完成" -ForegroundColor Green
}

function Run-Examples {
    Write-Host "🚀 运行缓存示例..." -ForegroundColor Blue
    go run examples/cache_example.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 缓存示例运行失败" -ForegroundColor Red
        return
    }
    
    Write-Host ""
    Write-Host "🚀 运行工具示例..." -ForegroundColor Blue
    go run examples/utils_example.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 工具示例运行失败" -ForegroundColor Red
        return
    }
    
    Write-Host ""
    Write-Host "🚀 运行健康检查示例..." -ForegroundColor Blue
    go run examples/health_example.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 健康检查示例运行失败" -ForegroundColor Red
        return
    }
    
    Write-Host ""
    Write-Host "🚀 运行日志示例..." -ForegroundColor Blue
    go run examples/logger_example.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ 日志示例运行失败" -ForegroundColor Red
        return
    }
    
    Write-Host "✅ 所有示例运行完成" -ForegroundColor Green
}

function Generate-Code {
    if ([string]::IsNullOrEmpty($Package)) {
        Write-Host "❌ 请指定包名: .\build.ps1 codegen -Package mypackage" -ForegroundColor Red
        exit 1
    }
    
    if (!(Test-Path "bin/codegen.exe")) {
        Write-Host "🔨 构建代码生成工具..." -ForegroundColor Blue
        Build-Tools
    }
    
    Write-Host "📝 生成 $Package 模块代码模板..." -ForegroundColor Blue
    & .\bin\codegen.exe -package $Package -output .
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ 代码模板生成完成" -ForegroundColor Green
    } else {
        Write-Host "❌ 代码模板生成失败" -ForegroundColor Red
        exit 1
    }
}

function Clean-Build {
    Write-Host "🧹 清理构建文件..." -ForegroundColor Blue
    if (Test-Path "bin") {
        Remove-Item -Recurse -Force "bin"
    }
    go clean -cache
    Write-Host "✅ 清理完成" -ForegroundColor Green
}

# 主逻辑
switch ($Command.ToLower()) {
    "help" { Show-Help }
    "build" { Build-Tools }
    "test" { Run-Tests }
    "check-api" { Check-API }
    "examples" { Run-Examples }
    "codegen" { Generate-Code }
    "clean" { Clean-Build }
    default {
        Write-Host "❌ 未知命令: $Command" -ForegroundColor Red
        Write-Host ""
        Show-Help
        exit 1
    }
}