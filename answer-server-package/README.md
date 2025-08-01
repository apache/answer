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
