# Answer 插件整合解決方案

## 問題描述

使用 Answer 官方的 `build --with` 命令添加插件時，會從官方代碼庫重新構建，導致自定義修改（Google Analytics、民眾之窗連結、X-Frame-Options 移除）丟失。

## 解決方案

使用 `ANSWER_MODULE` 環境變量指定本地已修改的源碼作為基礎來構建插件版本。

### 步驟

1. 確保本地源碼包含所有自定義修改
2. 使用以下命令構建插件版本：

```bash
ANSWER_MODULE=$(pwd) ./answer build --with github.com/apache/answer-plugins/embed-basic@latest --output ./answer-final-complete
```

### 關鍵要點

- `ANSWER_MODULE=$(pwd)` 告訴 Answer build 命令使用當前目錄（包含我們修改的源碼）作為基礎
- 構建過程會自動執行 `go mod edit -replace github.com/apache/answer=/path/to/local/source`
- 這樣既保留了自定義修改，又成功集成了官方插件

### 驗證方法

1. 檢查插件是否成功集成：
```bash
./answer-final-complete plugin
```

2. 啟動服務測試：
```bash
./answer-final-complete run -C ./data/
```

3. 檢查網頁是否包含：
   - Google Analytics 代碼 (G-NRX9V3TDXX)
   - 民眾之窗連結
   - embed-basic 插件功能
   - 可嵌入 iframe（X-Frame-Options 已移除）

## 自定義修改內容

### 1. Google Analytics 集成
- 檔案：`ui/template/header.html` 和 `ui/public/index.html`
- 追蹤代碼：G-NRX9V3TDXX

### 2. 民眾之窗連結
- 檔案：`ui/src/components/Footer/index.tsx` 和 `ui/template/footer.html`
- 連結：https://flash.justice-tw.org/grassway

### 3. 移除 X-Frame-Options
- 檔案：
  - `internal/router/ui.go:134`
  - `internal/controller/template_controller.go:626` 
  - `internal/base/middleware/auth.go:216`

### 4. 插件集成
- embed-basic@latest（支援多種嵌入格式：YouTube、Twitter、GitHub Gist 等）

## 雲端部署編譯

對於雲端部署（Ubuntu 22.04.5 LTS），使用以下命令進行交叉編譯：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ANSWER_MODULE=$(pwd) ./answer build --with github.com/apache/answer-plugins/embed-basic@latest --output ./answer-linux-final
```

## 構建成功確認

執行以下命令確認構建成功：
- `./answer-final-complete plugin` - 確認插件已集成
- 啟動服務後檢查網頁源碼包含 Google Analytics
- 檢查頁腳是否有民眾之窗連結
- 測試是否可嵌入 iframe

## 注意事項

1. 必須在包含自定義修改的源碼目錄中執行構建命令
2. `ANSWER_MODULE` 環境變量必須指向當前修改的源碼路徑
3. 構建過程會創建臨時目錄，完成後自動清理
4. 此方法適用於所有官方插件的集成

日期：2025-08-02