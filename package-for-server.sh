#!/bin/bash
# 打包 Answer 雲端部署檔案

echo "正在打包 Answer 雲端部署檔案..."

# 創建臨時目錄
mkdir -p answer-server-package

# 複製可執行檔
cp answer-linux answer-server-package/

# 創建資料目錄結構
mkdir -p answer-server-package/data/{conf,i18n,cache,uploads}

# 複製配置檔案
cp config-production.yaml answer-server-package/data/conf/config.yaml

# 複製國際化檔案
cp -r data/i18n/* answer-server-package/data/i18n/

# 創建啟動腳本
cat > answer-server-package/start.sh << 'EOF'
#!/bin/bash
echo "啟動 Answer 服務..."
./answer-linux run -C ./data
EOF

chmod +x answer-server-package/start.sh

# 創建 systemd 服務檔案範本
cat > answer-server-package/answer.service << 'EOF'
[Unit]
Description=Answer Service
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

# 創建安裝說明
cat > answer-server-package/README.md << 'EOF'
# Answer 雲端部署包

## 檔案說明
- `answer-linux`: Answer 可執行檔 (Linux amd64)
- `data/`: 資料目錄
  - `conf/config.yaml`: 配置檔案
  - `i18n/`: 國際化檔案
  - `cache/`: 快取目錄 (自動創建)
  - `uploads/`: 上傳檔案目錄 (自動創建)
- `start.sh`: 啟動腳本
- `answer.service`: systemd 服務檔案範本

## 部署步驟
1. 上傳整個資料夾到伺服器 `/root/answer_server/`
2. 設定權限: `chmod +x answer-linux start.sh`
3. 測試啟動: `./start.sh`
4. 設定系統服務:
   ```bash
   sudo cp answer.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable answer
   sudo systemctl start answer
   ```

## 端口設定
- 預設端口: 80 (可在 config.yaml 中修改)
- 如使用 80 端口需要 root 權限或配置防火牆

## 重要修改
- 已移除 X-Frame-Options: DENY (允許嵌入)
- 已加入 Google Analytics (G-NRX9V3TDXX)
EOF

# 創建壓縮包
tar -czf answer-server-linux-amd64.tar.gz answer-server-package/

echo "打包完成！檔案: answer-server-linux-amd64.tar.gz"
echo "檔案大小:"
ls -lh answer-server-linux-amd64.tar.gz

echo ""
echo "📦 需要搬移到伺服器的檔案清單:"
echo "1. answer-server-linux-amd64.tar.gz (包含所有必要檔案)"
echo ""
echo "🚀 或者手動複製以下檔案到伺服器:"
echo "├── answer-linux (可執行檔, 76MB)"
echo "├── data/"
echo "│   ├── conf/config.yaml (配置檔案)"
echo "│   └── i18n/ (國際化檔案, ~2MB)"
echo "└── 其他輔助檔案 (start.sh, answer.service 等)"