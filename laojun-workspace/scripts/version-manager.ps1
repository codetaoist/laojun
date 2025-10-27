# Laojun 版本管理脚本
# 用于管理各个仓库的版本标签和发布

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("list", "tag", "release", "sync")]
    [string]$Action,
    
    [string]$Repository = "all",
    [string]$Version,
    [string]$Message,
    [switch]$DryRun,
    [switch]$Force,
    [switch]$Verbose
)

$ErrorActionPreference = "Stop"

Write-Host "=== Laojun 版本管理 ===" -ForegroundColor Green

# 工作区根目录
$workspaceRoot = Split-Path -Parent $PSScriptRoot
$parentDir = Split-Path -Parent $workspaceRoot

# 所有仓库配置
$repositories = @(
    @{
        Name = "laojun-shared"
        Path = "laojun-shared"
        Type = "library"
        Description = "共享组件库"
        VersionPrefix = "v"
    },
    @{
        Name = "laojun-plugins"
        Path = "laojun-plugins"
        Type = "library"
        Description = "插件系统"
        VersionPrefix = "v"
    },
    @{
        Name = "laojun-config-center"
        Path = "laojun-config-center"
        Type = "service"
        Description = "配置中心"
        VersionPrefix = "v"
    },
    @{
        Name = "laojun-admin-api"
        Path = "laojun-admin-api"
        Type = "service"
        Description = "管理后端 API"
        VersionPrefix = "v"
    },
    @{
        Name = "laojun-marketplace-api"
        Path = "laojun-marketplace-api"
        Type = "service"
        Description = "市场后端 API"
        VersionPrefix = "v"
    },
    @{
        Name = "laojun-admin-web"
        Path = "laojun-admin-web"
        Type = "frontend"
        Description = "管理前端"
        VersionPrefix = "v"
    },
    @{
        Name = "laojun-marketplace-web"
        Path = "laojun-marketplace-web"
        Type = "frontend"
        Description = "市场前端"
        VersionPrefix = "v"
    }
)

function Get-RepositoryVersions {
    param($repo)
    
    $repoPath = Join-Path $parentDir $repo.Path
    if (-not (Test-Path $repoPath)) {
        return @()
    }
    
    Set-Location $repoPath
    
    try {
        # 获取所有标签
        $tags = git tag -l "$($repo.VersionPrefix)*" --sort=-version:refname 2>$null
        if ($LASTEXITCODE -ne 0) {
            return @()
        }
        
        $versions = @()
        foreach ($tag in $tags) {
            # 获取标签信息
            $tagInfo = git show --format="%H|%an|%ad|%s" --no-patch $tag 2>$null
            if ($tagInfo) {
                $parts = $tagInfo -split '\|'
                $versions += @{
                    Tag = $tag
                    Version = $tag -replace "^$($repo.VersionPrefix)", ""
                    Commit = $parts[0]
                    Author = $parts[1]
                    Date = $parts[2]
                    Message = $parts[3]
                }
            }
        }
        
        return $versions
    } catch {
        Write-Warning "获取 $($repo.Name) 版本信息失败: $_"
        return @()
    }
}

function Get-CurrentBranch {
    param($repoPath)
    
    Set-Location $repoPath
    try {
        $branch = git rev-parse --abbrev-ref HEAD 2>$null
        return if ($LASTEXITCODE -eq 0) { $branch } else { "unknown" }
    } catch {
        return "unknown"
    }
}

function Get-CommitsSinceTag {
    param($repoPath, $tag)
    
    Set-Location $repoPath
    try {
        if ($tag) {
            $commits = git rev-list "$tag..HEAD" --count 2>$null
            return if ($LASTEXITCODE -eq 0) { [int]$commits } else { 0 }
        } else {
            $commits = git rev-list HEAD --count 2>$null
            return if ($LASTEXITCODE -eq 0) { [int]$commits } else { 0 }
        }
    } catch {
        return 0
    }
}

function Test-WorkingDirectoryClean {
    param($repoPath)
    
    Set-Location $repoPath
    try {
        $status = git status --porcelain 2>$null
        return ($LASTEXITCODE -eq 0 -and -not $status)
    } catch {
        return $false
    }
}

function New-RepositoryTag {
    param($repo, $version, $message)
    
    $repoPath = Join-Path $parentDir $repo.Path
    if (-not (Test-Path $repoPath)) {
        Write-Error "仓库路径不存在: $($repo.Path)"
        return $false
    }
    
    Set-Location $repoPath
    
    # 检查工作目录是否干净
    if (-not (Test-WorkingDirectoryClean $repoPath)) {
        Write-Error "$($repo.Name): 工作目录不干净，请先提交或暂存更改"
        return $false
    }
    
    # 检查是否在主分支
    $currentBranch = Get-CurrentBranch $repoPath
    if ($currentBranch -notin @("main", "master", "develop")) {
        Write-Warning "$($repo.Name): 当前分支 '$currentBranch' 不是主分支"
        if (-not $Force) {
            Write-Error "使用 -Force 参数强制在当前分支创建标签"
            return $false
        }
    }
    
    $fullTag = "$($repo.VersionPrefix)$version"
    
    # 检查标签是否已存在
    $existingTag = git tag -l $fullTag 2>$null
    if ($existingTag -and -not $Force) {
        Write-Error "$($repo.Name): 标签 '$fullTag' 已存在，使用 -Force 参数覆盖"
        return $false
    }
    
    try {
        if ($DryRun) {
            Write-Host "$($repo.Name): [DRY RUN] 将创建标签 '$fullTag'" -ForegroundColor Yellow
            if ($message) {
                Write-Host "  消息: $message" -ForegroundColor Gray
            }
            return $true
        }
        
        # 创建标签
        if ($message) {
            git tag -a $fullTag -m $message
        } else {
            git tag $fullTag
        }
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "$($repo.Name): ✓ 创建标签 '$fullTag'" -ForegroundColor Green
            
            # 推送标签到远程（如果有远程仓库）
            $remotes = git remote 2>$null
            if ($remotes -and $remotes.Contains("origin")) {
                if ($DryRun) {
                    Write-Host "  [DRY RUN] 将推送标签到远程" -ForegroundColor Yellow
                } else {
                    git push origin $fullTag 2>$null
                    if ($LASTEXITCODE -eq 0) {
                        Write-Host "  ✓ 推送标签到远程" -ForegroundColor Green
                    } else {
                        Write-Warning "  推送标签到远程失败"
                    }
                }
            }
            
            return $true
        } else {
            Write-Error "$($repo.Name): 创建标签失败"
            return $false
        }
    } catch {
        Write-Error "$($repo.Name): 创建标签异常: $_"
        return $false
    }
}

function Show-RepositoryVersions {
    param($repos)
    
    Write-Host "`n版本信息:" -ForegroundColor Cyan
    Write-Host ("{0,-25} {1,-15} {2,-10} {3,-15} {4}" -f "仓库", "最新版本", "提交数", "分支", "状态") -ForegroundColor White
    Write-Host ("-" * 80) -ForegroundColor Gray
    
    foreach ($repo in $repos) {
        $repoPath = Join-Path $parentDir $repo.Path
        if (-not (Test-Path $repoPath)) {
            Write-Host ("{0,-25} {1,-15} {2,-10} {3,-15} {4}" -f $repo.Name, "N/A", "N/A", "N/A", "不存在") -ForegroundColor Red
            continue
        }
        
        $versions = Get-RepositoryVersions $repo
        $latestVersion = if ($versions.Count -gt 0) { $versions[0].Version } else { "无标签" }
        $latestTag = if ($versions.Count -gt 0) { $versions[0].Tag } else { $null }
        
        $currentBranch = Get-CurrentBranch $repoPath
        $commitsSince = Get-CommitsSinceTag $repoPath $latestTag
        $isClean = Test-WorkingDirectoryClean $repoPath
        
        $status = if ($isClean) { "干净" } else { "有更改" }
        if ($commitsSince -gt 0) {
            $status += " (+$commitsSince)"
        }
        
        $color = if ($isClean -and $commitsSince -eq 0) { "Green" } elseif ($isClean) { "Yellow" } else { "Red" }
        
        Write-Host ("{0,-25} {1,-15} {2,-10} {3,-15} {4}" -f $repo.Name, $latestVersion, $commitsSince, $currentBranch, $status) -ForegroundColor $color
        
        if ($Verbose -and $versions.Count -gt 0) {
            Write-Host "  最近版本:" -ForegroundColor Gray
            foreach ($version in $versions | Select-Object -First 3) {
                Write-Host ("    {0} - {1} ({2})" -f $version.Tag, $version.Message, $version.Date) -ForegroundColor Gray
            }
        }
    }
}

# 主逻辑
Set-Location $workspaceRoot

# 过滤仓库
$targetRepos = if ($Repository -eq "all") {
    $repositories
} else {
    $repositories | Where-Object { $_.Name -eq $Repository }
}

if (-not $targetRepos) {
    Write-Error "未找到仓库: $Repository"
    Write-Host "可用仓库:" -ForegroundColor Yellow
    foreach ($repo in $repositories) {
        Write-Host "  - $($repo.Name)" -ForegroundColor White
    }
    exit 1
}

switch ($Action) {
    "list" {
        Show-RepositoryVersions $targetRepos
    }
    
    "tag" {
        if (-not $Version) {
            Write-Error "创建标签需要指定版本号 (-Version)"
            exit 1
        }
        
        # 验证版本号格式（简单的语义版本检查）
        if ($Version -notmatch '^\d+\.\d+\.\d+(-[\w\.-]+)?$') {
            Write-Error "版本号格式无效，应为 x.y.z 或 x.y.z-suffix"
            exit 1
        }
        
        Write-Host "创建版本标签: $Version" -ForegroundColor Cyan
        if ($DryRun) {
            Write-Host "[DRY RUN 模式]" -ForegroundColor Yellow
        }
        Write-Host ""
        
        $successCount = 0
        foreach ($repo in $targetRepos) {
            if (New-RepositoryTag $repo $Version $Message) {
                $successCount++
            }
        }
        
        Write-Host "`n标签创建完成: $successCount/$($targetRepos.Count)" -ForegroundColor Green
    }
    
    "release" {
        Write-Host "发布功能开发中..." -ForegroundColor Yellow
        Write-Host "将来会支持:" -ForegroundColor Gray
        Write-Host "  - 自动生成 CHANGELOG" -ForegroundColor Gray
        Write-Host "  - 创建 GitHub Release" -ForegroundColor Gray
        Write-Host "  - 构建和上传制品" -ForegroundColor Gray
    }
    
    "sync" {
        Write-Host "同步版本信息..." -ForegroundColor Cyan
        
        # 检查所有仓库的版本一致性
        $versionMap = @{}
        foreach ($repo in $targetRepos) {
            $versions = Get-RepositoryVersions $repo
            if ($versions.Count -gt 0) {
                $latestVersion = $versions[0].Version
                if (-not $versionMap.ContainsKey($latestVersion)) {
                    $versionMap[$latestVersion] = @()
                }
                $versionMap[$latestVersion] += $repo.Name
            }
        }
        
        Write-Host "`n版本分布:" -ForegroundColor White
        foreach ($version in $versionMap.Keys | Sort-Object -Descending) {
            $repos = $versionMap[$version] -join ", "
            Write-Host "  $version : $repos" -ForegroundColor Gray
        }
        
        if ($versionMap.Keys.Count -gt 1) {
            Write-Host "`n⚠ 检测到版本不一致，建议统一版本号" -ForegroundColor Yellow
        } else {
            Write-Host "`n✓ 所有仓库版本一致" -ForegroundColor Green
        }
    }
}

Write-Host "`n版本管理完成！" -ForegroundColor Green