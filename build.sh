#!/bin/bash

# Answer 自動建置腳本
# 支援本地測試和雲端部署兩種模式

set -e

echo "========================================"
echo "     Answer 自動建置工具"
echo "========================================"
echo ""

# 檢查是否在正確的目錄
if [ ! -f "go.mod" ] || [ ! -d "ui" ] || [ ! -f "PLUGIN_INTEGRATION_SOLUTION.md" ]; then
    echo "❌ 錯誤：請在 Answer 專案根目錄執行此腳本"
    exit 1
fi

# 檢查基本工具
command -v go >/dev/null 2>&1 || { echo "❌ 錯誤：需要安裝 Go"; exit 1; }
command -v pnpm >/dev/null 2>&1 || { echo "❌ 錯誤：需要安裝 pnpm"; exit 1; }

echo "✅ 環境檢查通過"
echo ""

# 選擇建置類型
echo "請選擇建置類型："
echo "1) 本地測試版本 (macOS/當前平台)"
echo "2) 雲端部署版本 (Linux AMD64)"
echo ""
read -p "請輸入選項 (1 或 2): " BUILD_TYPE

case $BUILD_TYPE in
    1)
        echo ""
        echo "🔨 開始建置本地測試版本..."
        echo ""
        
        # 建置本地版本
        echo "步驟 1/3: 使用 ANSWER_MODULE 建置本地版本..."
        ANSWER_MODULE=$(pwd) ./answer build --with github.com/apache/answer-plugins/embed-basic@latest --output ./answer-local-test
        
        echo ""
        echo "步驟 2/3: 驗證建置結果..."
        if [ -f "./answer-local-test" ]; then
            echo "✅ 建置成功！檔案位置: ./answer-local-test"
            echo "✅ 檔案大小: $(du -h ./answer-local-test | cut -f1)"
        else
            echo "❌ 建置失敗！"
            exit 1
        fi
        
        echo ""
        echo "步驟 3/3: 檢查插件..."
        ./answer-local-test plugin
        
        echo ""
        echo "🎉 本地測試版本建置完成！"
        echo ""
        echo "使用方法："
        echo "  ./answer-local-test run -C ./data/"
        echo ""
        echo "功能包含："
        echo "  ✅ Google Analytics (G-NRX9V3TDXX)"
        echo "  ✅ 民眾之窗連結"
        echo "  ✅ embed-basic 插件"
        echo "  ✅ 移除 X-Frame-Options (支援 iframe)"
        ;;
        
    2)
        echo ""
        echo "🔨 開始建置雲端部署版本..."
        echo ""
        
        # 建置 Linux 版本
        echo "步驟 1/3: 使用交叉編譯建置 Linux 版本..."
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ANSWER_MODULE=$(pwd) ./answer build --with github.com/apache/answer-plugins/embed-basic@latest --output ./answer-linux-deploy
        
        echo ""
        echo "步驟 2/3: 驗證建置結果..."
        if [ -f "./answer-linux-deploy" ]; then
            echo "✅ 建置成功！檔案位置: ./answer-linux-deploy"
            echo "✅ 檔案大小: $(du -h ./answer-linux-deploy | cut -f1)"
            echo "✅ 檔案類型: $(file ./answer-linux-deploy)"
        else
            echo "❌ 建置失敗！"
            exit 1
        fi
        
        echo ""
        echo "步驟 3/3: 創建部署包..."
        
        # 創建部署資料夾
        DEPLOY_DIR="answer-cloud-deploy-$(date +%Y%m%d_%H%M%S)"
        mkdir -p "$DEPLOY_DIR"
        
        # 複製必要檔案
        cp ./answer-linux-deploy "$DEPLOY_DIR/answer"
        cp -r ./data "$DEPLOY_DIR/" 2>/dev/null || echo "⚠️  data 目錄不存在，將在首次執行時創建"
        
        # 創建啟動腳本
        cat > "$DEPLOY_DIR/start.sh" << 'EOF'
#!/bin/bash

# Answer 雲端啟動腳本
echo "正在啟動 Answer 服務..."

# 檢查 data 目錄
if [ ! -d "./data" ]; then
    echo "創建 data 目錄..."
    mkdir -p ./data
fi

# 設置權限
chmod +x ./answer

# 啟動服務
echo "服務啟動中..."
./answer run -C ./data/

EOF
        
        chmod +x "$DEPLOY_DIR/start.sh"
        
        # 創建 README
        cat > "$DEPLOY_DIR/README.md" << EOF
# Answer 雲端部署包

## 功能特色
- ✅ Google Analytics 追蹤 (G-NRX9V3TDXX)
- ✅ 民眾之窗連結整合
- ✅ embed-basic 插件 (支援 YouTube、Twitter、GitHub Gist 等嵌入)
- ✅ 移除 X-Frame-Options (完全支援 iframe 嵌入)
- ✅ Ubuntu 22.04.5 LTS 相容

## 檔案說明
- \`answer\`: 主程式 (Linux AMD64 靜態編譯)
- \`start.sh\`: 啟動腳本
- \`data/\`: 數據目錄 (配置、數據庫、上傳檔案等)

## 使用方法

### 1. 上傳到伺服器
\`\`\`bash
scp -r $DEPLOY_DIR user@your-server:/opt/answer/
\`\`\`

### 2. 在伺服器上執行
\`\`\`bash
cd /opt/answer
./start.sh
\`\`\`

### 3. 瀏覽器訪問
http://your-server-ip:80

## 注意事項
1. 確保防火牆開放 80 端口
2. data 目錄需要寫入權限
3. 首次執行會進入初始化設置

## 技術細節
- 編譯時間: $(date)
- 插件版本: embed-basic@latest
- 建置方法: ANSWER_MODULE 本地源碼覆蓋
EOF

        echo "✅ 部署包創建完成: $DEPLOY_DIR/"
        echo ""
        echo "🎉 雲端部署版本建置完成！"
        echo ""
        echo "部署包內容："
        ls -la "$DEPLOY_DIR/"
        echo ""
        echo "上傳到伺服器："
        echo "  scp -r $DEPLOY_DIR user@your-server:/opt/answer/"
        echo ""
        echo "在伺服器執行："
        echo "  cd /opt/answer && ./start.sh"
        ;;
        
    *)
        echo "❌ 無效選項，請輸入 1 或 2"
        exit 1
        ;;
esac

echo ""
echo "========================================"
echo "         建置完成！"
echo "========================================"