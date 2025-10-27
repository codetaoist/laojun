# 批量修复UTF-8编码问题的PowerShell脚本

# 定义需要修复的文件路径
$files = @(
    "internal\handlers\developer_handler.go",
    "internal\handlers\plugin_review_handler.go", 
    "internal\handlers\extended_plugin_handler.go",
    "internal\handlers\community_handler.go"
)

# 定义替换映射
$replacements = @{
    "�?" = "求"
    "�" = "者"
    "�?" = "滤"
    "�?" = "务"
    "�?" = "态"
    "�?" = "值"
    "�?" = "天"
    "�?" = "境"
    "�?" = "件"
    "�?" = "名"
    "�?" = "息"
    "�?" = "录"
    "�?" = "查"
    "�?" = "点"
    "�?" = "现"
    "�?" = "据"
    "�?" = "期"
    "�?" = "员"
    "�?" = "理"
    "�?" = "诉"
    "�?" = "负"
    "�?" = "荷"
}

foreach ($file in $files) {
    $fullPath = Join-Path $PSScriptRoot $file
    if (Test-Path $fullPath) {
        Write-Host "Processing $file..."
        
        # 读取文件内容
        $content = Get-Content $fullPath -Raw -Encoding UTF8
        
        # 执行替换
        foreach ($key in $replacements.Keys) {
            $content = $content -replace [regex]::Escape($key), $replacements[$key]
        }
        
        # 写回文件
        $content | Out-File $fullPath -Encoding UTF8 -NoNewline
        
        Write-Host "Fixed $file"
    } else {
        Write-Host "File not found: $file"
    }
}

Write-Host "Encoding fix completed!"