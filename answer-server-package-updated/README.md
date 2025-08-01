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
