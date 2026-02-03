#!/bin/bash

# 重新关联远程地址并推送到指定分支的脚本
# 使用方法：./push_to_remote.sh <远程地址> <分支名>
# 例如：./push_to_remote.sh https://github.com/username/repo.git main

if [ $# -lt 2 ]; then
    echo "使用方法: $0 <远程地址> <分支名>"
    echo "例如: $0 https://github.com/username/repo.git main"
    exit 1
fi

REMOTE_URL=$1
BRANCH_NAME=$2

echo "=== 重新关联远程仓库 ==="
echo "远程地址: $REMOTE_URL"
echo "目标分支: $BRANCH_NAME"
echo ""

# 1. 移除旧的远程地址（如果存在）
echo "1. 移除旧的 origin 远程..."
git remote remove origin 2>/dev/null || true

# 2. 添加新的远程地址
echo "2. 添加新的远程地址..."
git remote add origin "$REMOTE_URL"

# 3. 验证远程地址
echo "3. 验证远程地址..."
git remote -v

# 4. 检查当前分支，如果不存在目标分支则创建
echo "4. 检查/创建分支 $BRANCH_NAME..."
current_branch=$(git branch --show-current)
if [ "$current_branch" != "$BRANCH_NAME" ]; then
    echo "   当前分支: $current_branch"
    echo "   创建并切换到分支: $BRANCH_NAME"
    git checkout -b "$BRANCH_NAME" 2>/dev/null || git checkout "$BRANCH_NAME"
else
    echo "   已在目标分支 $BRANCH_NAME"
fi

# 5. 推送代码到指定分支
echo "5. 推送代码到远程分支 $BRANCH_NAME..."
git push -u origin "$BRANCH_NAME"

echo ""
echo "=== 完成 ==="
echo "代码已推送到: $REMOTE_URL (分支: $BRANCH_NAME)"
