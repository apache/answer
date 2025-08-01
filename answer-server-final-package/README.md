# Answer 雲端部署最終版（包含所有自定義功能）

## ✨ 包含功能
✅ **Google Analytics 追蹤** (G-NRX9V3TDXX)  
✅ **移除 X-Frame-Options** - 允許網頁嵌入  
✅ **Footer 添加「民眾之窗」連結** - https://flash.justice-tw.org/grassway  
✅ **Linux amd64 靜態編譯版本** - 無依賴性問題  

## 📁 檔案說明
- `answer-linux`: Answer 可執行檔 (Linux amd64，包含所有自定義功能)
- `data/`: 資料目錄
  - `conf/config.yaml`: 生產環境配置檔案
  - `i18n/`: 國際化檔案 (40+ 語言)
  - `cache/`: 快取目錄 (自動創建)
  - `uploads/`: 上傳檔案目錄 (自動創建)
- `start.sh`: 啟動腳本
- `answer.service`: systemd 服務檔案範本

## 🚀 部署步驟

### 首次部署
1. 上傳檔案到雲端伺服器：`/root/answer_server/`
2. 解壓部署包：`tar -xzf answer-server-final-linux-amd64.tar.gz`
3. 進入目錄：`cd answer-server-final-package`
4. 設定權限：`chmod +x answer-linux start.sh`
5. 測試啟動：`./start.sh`

### 更新現有部署
1. 停止舊服務：`sudo systemctl stop answer`
2. 備份資料：`cp -r data data_backup_$(date +%Y%m%d)`
3. 替換可執行檔：`cp answer-linux-final /path/to/current/answer-linux`
4. 重啟服務：`sudo systemctl restart answer`

### 系統服務設定
```bash
sudo cp answer.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable answer
sudo systemctl start answer
sudo systemctl status answer
```

## 🔍 功能驗證

### 1. Google Analytics
檢查網站源碼包含：
```bash
curl -s http://your-domain.com/ | grep "googletagmanager"
```
應該看到：`<script async src="https://www.googletagmanager.com/gtag/js?id=G-NRX9V3TDXX"></script>`

### 2. 嵌入功能
檢查 HTTP 標頭：
```bash
curl -I http://your-domain.com/ | grep -i "x-frame-options"
```
應該沒有返回結果（表示已移除限制）

### 3. 民眾之窗連結
檢查網站底部應顯示：
`Terms of Service | Privacy Policy | 民眾之窗`

## ⚙️ 配置選項

### 端口設定
預設使用端口 80，如需修改請編輯 `data/conf/config.yaml`：
```yaml
server:
  http:
    addr: 0.0.0.0:3000  # 改為其他端口
```

### 域名設定
如有正式域名，請更新 `data/conf/config.yaml`：
```yaml
ui:
  base_url: "https://yourdomain.com"
  api_base_url: "https://yourdomain.com"
```

## 🛠️ 疑難排解

### 服務無法啟動
```bash
# 檢查日誌
journalctl -u answer -f

# 檢查端口佔用
netstat -tlnp | grep :80

# 測試直接啟動
./answer-linux run -C ./data
```

### 權限問題
```bash
# 80 端口需要 root 權限
sudo ./start.sh

# 或改用非特權端口（如 3000）
```

## 📞 技術支援

如遇問題請提供：
1. 錯誤日誌：`journalctl -u answer --no-pager`
2. 系統資訊：`uname -a && cat /etc/os-release`
3. 網路狀態：`netstat -tlnp | grep answer`

---
**版本資訊**  
- Answer 版本：1.6.0  
- 編譯時間：$(date)  
- 架構：linux/amd64  
