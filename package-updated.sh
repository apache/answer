#!/bin/bash
# 打包包含 Google Analytics 的 Answer 雲端部署檔案

echo "正在打包更新版 Answer 雲端部署檔案（包含 Google Analytics）..."

# 刪除舊的打包目錄
rm -rf answer-server-package-updated

# 創建新的打包目錄
mkdir -p answer-server-package-updated

# 複製新的可執行檔
cp answer-linux-with-ga answer-server-package-updated/answer-linux

# 創建資料目錄結構
mkdir -p answer-server-package-updated/data/{conf,i18n,cache,uploads}

# 複製配置檔案
cp config-production.yaml answer-server-package-updated/data/conf/config.yaml

# 複製國際化檔案
cp -r data/i18n/* answer-server-package-updated/data/i18n/

# 創建啟動腳本
cat > answer-server-package-updated/start.sh << 'EOF'
#!/bin/bash
echo "啟動 Answer 服務（包含 Google Analytics）..."
./answer-linux run -C ./data
EOF

chmod +x answer-server-package-updated/start.sh

# 創建 systemd 服務檔案範本
cat > answer-server-package-updated/answer.service << 'EOF'
[Unit]
Description=Answer Service with Google Analytics
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/answer_server
ExecStart=/root/answer_server/answer-linux run -C /root/answer_server/data
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# 創建更新說明
cat > answer-server-package-updated/README.md << 'EOF'
# Answer 雲端部署包（包含 Google Analytics）

## 更新內容
✅ **已加入 Google Analytics 追蹤 (G-NRX9V3TDXX)**
✅ **已移除 X-Frame-Options: DENY（允許嵌入）**
✅ **Linux amd64 靜態編譯版本**

## 檔案說明
- `answer-linux`: Answer 可執行檔 (Linux amd64，包含 GA)
- `data/`: 資料目錄
  - `conf/config.yaml`: 配置檔案
  - `i18n/`: 國際化檔案
  - `cache/`: 快取目錄 (自動創建)
  - `uploads/`: 上傳檔案目錄 (自動創建)
- `start.sh`: 啟動腳本
- `answer.service`: systemd 服務檔案範本

## 部署步驟
1. 停止舊版服務: `sudo systemctl stop answer`
2. 備份舊版資料: `cp -r /root/answer_server/data /root/answer_server_backup/`
3. 上傳並解壓到 `/root/answer_server/`
4. 設定權限: `chmod +x answer-linux start.sh`
5. 測試啟動: `./start.sh`
6. 更新服務: `sudo systemctl restart answer`

## 驗證 Google Analytics
檢查網站源碼應包含：
```html
<script async src="https://www.googletagmanager.com/gtag/js?id=G-NRX9V3TDXX"></script>
```

## 驗證嵌入功能
網站不應包含 `X-Frame-Options: DENY` 標頭
EOF

# 創建壓縮包
tar -czf answer-server-with-ga-linux-amd64.tar.gz answer-server-package-updated/

echo "✅ 更新版打包完成！"
echo "📦 檔案: answer-server-with-ga-linux-amd64.tar.gz"
echo "📊 檔案大小:"
ls -lh answer-server-with-ga-linux-amd64.tar.gz

echo ""
echo "🎯 主要更新："
echo "- ✅ Google Analytics (G-NRX9V3TDXX) 已整合"
echo "- ✅ X-Frame-Options 已移除（可嵌入）"
echo "- ✅ 靜態編譯，無依賴性問題"
echo ""
echo "🚀 上傳指令："
echo "scp -P 2323 answer-server-with-ga-linux-amd64.tar.gz oliver0804@dev.bashcat.net:~/"