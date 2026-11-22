# 專案目錄與檔案說明

本文件依 2026-11-22 的專案內容整理，說明各目錄、檔案與主要功能的對應關係。
環境變數只記錄名稱及用途，不記錄實際金鑰或 token；資料庫只依程式說明結構，不列出個人記帳內容。

## 完整目錄結構

```text
cash_carole_record/
├── .env
├── .env.example
├── .gitignore
├── .vscode/
│   └── settings.json
├── AGENTS.md
├── README.md
├── go.mod
├── go.sum
├── main.go
├── data/
│   └── records.db
├── docs/
│   └── README.md
├── migrations/
│   ├── 001_records_category_foreign_key.sql
│   └── 002_records_line_user_id.sql
├── logs/                         # 服務啟動時建立
│   ├── app-YYYY-MM-DD.log
│   ├── gin.log                   # 舊版留下的日誌，保留不再寫入
│   ├── gin-YYYY-MM-DD.log
│   └── webhook-YYYY-MM-DD.log
└── internal/
    ├── config/
    │   └── config.go
    ├── line/
    │   ├── mealReminder.go
    │   ├── sendMessages.go
    │   ├── types.go
    │   ├── webhookEventHandler.go
    │   └── webhookVerifier.go
    ├── logging/
    │   └── dailyWriter.go
    ├── record/
    │   ├── analysisQuery.go
    │   ├── analyzer.go
    │   ├── db.go
    │   ├── handler.go
    │   ├── mealReminder.go
    │   └── model.go
    └── web-frontend/
        ├── errorLogging.js
        ├── index.html
        ├── liffLogin.html
        ├── spendInput.html
        └── statistics.html
```

## 各目錄用途

| 目錄 | 存放內容 |
| --- | --- |
| 專案根目錄 | 主程式、Go 套件依賴、環境設定與使用說明。 |
| `.vscode/` | 此工作區的 VS Code 編輯器設定。 |
| `data/` | 執行時使用的 SQLite 資料庫。 |
| `docs/` | 專案結構與程式職責的說明文件。 |
| `migrations/` | 既有資料庫結構的一次性遷移 SQL。 |
| `logs/` | 一般應用程式、Gin 與 webhook 分開保存的每日日誌，以及保留的歷史檔案。 |
| `internal/` | 本專案內部使用的後端套件及前端靜態檔案。 |
| `internal/config/` | 環境變數載入與必要設定驗證。 |
| `internal/line/` | LINE webhook 驗證、事件處理、資料型別及 Messaging API 呼叫。 |
| `internal/logging/` | 共用的每日分檔日誌工具。 |
| `internal/record/` | 記帳資料模型、資料庫初始化、HTTP API 與 Gemini 記帳分析。 |
| `internal/web-frontend/` | 登入、首頁、記帳頁面，以及前端共用錯誤日誌工具。 |

## 根目錄檔案

| 檔案 | 放了什麼 |
| --- | --- |
| [AGENTS.md](../AGENTS.md) | 專案工作指引。要求每次新對話或工作階段首次修改前，完整閱讀本文件以了解架構。 |
| [main.go](../main.go) | 程式入口。先設定一般日誌同步輸出到每日檔案與終端機，再載入設定、建立 Gin 與 webhook 每日日誌、LINE client、資料庫與分析服務，啟動每日三餐提醒排程，註冊路由並啟動 HTTP server。 |
| [.env](../.env) | 本機環境設定，例如服務埠、LINE 憑證、LIFF ID 與 Gemini 金鑰。由 `config.Load()` 讀取；已存在的系統環境變數優先。 |
| [.env.example](../.env.example) | 設定範本，保留 `.env` 的參數順序與註解，值全部留空。包含 `PORT`、LINE secret／token、`LIFF_ID` 與 `GEMINI_API_KEY`。 |
| [.gitignore](../.gitignore) | 忽略 `.env`、`data/`、`logs/` 與 `.DS_Store`，保留 `.env.example` 供提交；不會自動移除已追蹤檔案。 |
| [go.mod](../go.mod) | Go module 名稱、Go 版本與依賴套件版本。使用 Gin、godotenv 與 modernc SQLite；Gemini 與 LINE API 透過標準函式庫 `net/http` 呼叫。 |
| [go.sum](../go.sum) | Go 依賴套件的校驗資訊，用來確認下載內容與預期一致。 |
| [README.md](../README.md) | 專案使用說明，包括環境設定、LINE 控制台設定、記帳問答、錯誤日誌與啟動方式。 |

### `.env` 設定由哪裡使用

以下為程式支援的名稱，不代表目前 `.env` 一定有設定每一項。

| 變數 | 用途與預設行為 |
| --- | --- |
| `PORT` | HTTP server 監聽埠，未設定時使用 `8080`。 |
| `LINE_CHANNEL_SECRET` | 驗證 LINE webhook 簽章，啟動時檢查必填。 |
| `LINE_CHANNEL_ACCESS_TOKEN` | 呼叫 LINE Messaging API，啟動時檢查必填。 |
| `LIFF_ID` | 前端登入時透過 `/api/liff-config` 取得，用於 LIFF 初始化。 |
| `GEMINI_API_KEY` | 呼叫 Gemini API；未設定時，文字問答會回覆設定提示。 |
| `GEMINI_MODEL` | Gemini 模型名稱，未設定時使用 `gemini-3.5-flash-lite`。 |

## `.vscode/`

| 檔案 | 放了什麼 |
| --- | --- |
| [settings.json](../.vscode/settings.json) | 工作區介面配色與 Peacock 顏色設定，調整 VS Code 的活動列、標題列與狀態列等外觀。 |

## `data/`

| 檔案 | 放了什麼 |
| --- | --- |
| [records.db](../data/records.db) | SQLite 資料庫，存放記帳紀錄、分類及 LINE 使用者白名單；結構與預設分類定義於 `internal/record/db.go`。 |

`main.go` 使用相對路徑 `data/records.db`。實際開啟哪一份資料庫取決於服務啟動時的工作目錄，
初始化日誌會輸出資料庫的絕對路徑與分類筆數。
`InitDB()` 會在開啟 SQLite 前自動建立資料庫父目錄；`data/` 不存在時不需手動建立。
建立目錄失敗會回傳錯誤並停止啟動；既有資料庫仍會沿用，不清空。

### 資料表對應

| 資料表 | 用途 | 主要欄位 |
| --- | --- | --- |
| `records` | 每筆記帳資料與原始記帳者。 | `id`、`category_id`、`tag`、`note`、`amount`、`date`、`created_at`、`line_user_id`。 |
| `spendCategories` | 記帳分類與顯示圖示。 | `id`、`name`、`icon`。 |
| `lineUserWhitelist` | LINE 使用者授權名單。 | `user_id`、`created_at`。 |

`records.category_id` 以外鍵參照 `spendCategories.id`，並建立 `idx_records_category_id` 索引。
`ON UPDATE CASCADE` 會在分類 ID 更新時同步更新記帳；`ON DELETE RESTRICT` 禁止刪除仍被使用的分類。
`db.go` 透過 DSN 為連線池中的每個新連線啟用 `foreign_keys(1)`；舊服務需重啟才會使用新連線設定。

`records.line_user_id` 儲存原始記帳者的 LINE userId，API 欄位名稱為 `lineUserId`。
新增時由前端讀取登入後保存的 `lineUserId`，後端要求非空並檢查 LINE 使用者白名單。
目前沿用前端傳送 ID 的方式，白名單檢查不等於驗證 LINE 登入憑證。
編輯只更新記帳內容，不覆寫原作者。查詢、統計與 Gemini 問答仍使用共用帳本，不依作者篩選。
`LIFF_ID` 是應用程式 ID，不作為個人識別。
`income` 分類視為收入，其餘分類視為支出。

## `docs/`

| 檔案 | 放了什麼 |
| --- | --- |
| [README.md](README.md) | 本文件，整理所有目錄及檔案用途，並提供路由、主要流程與修改位置的索引。 |

## `migrations/`

| 檔案 | 放了什麼 |
| --- | --- |
| [001_records_category_foreign_key.sql](../migrations/001_records_category_foreign_key.sql) | 為既有 `records` 新增分類外鍵的交易式遷移。建立新表、複製全部資料、替換舊表、保留遞增序號，最後建立分類索引並檢查外鍵。 |
| [002_records_line_user_id.sql](../migrations/002_records_line_user_id.sql) | 為既有 `records` 新增 `line_user_id TEXT NOT NULL DEFAULT ''`，並依使用者指定將既有紀錄作者填為 `Ud4af9c2ce7cca29234bfeb60b9da4198`。之後新增紀錄使用當時登入者的 ID。 |

遷移不在服務啟動時自動執行，也不應重複套用到已有此外鍵的資料庫。
執行前先備份，確認沒有孤立紀錄，並檢查原表的額外欄位、索引、trigger、view 與其他表參照。
由執行端控制遇錯停止與回滾，提交前核對資料、ID、序號及完整性。

`002` 僅適用於尚未有 `line_user_id` 的既有資料庫，不可重複執行。
若兩份遷移皆需套用，必須先執行 `001` 再執行 `002`，避免舊版重建資料表時移除新增欄位。
新資料庫的建表 SQL 已包含記帳者欄位，不需執行 `002`；既有資料庫須先遷移才能使用新版 API。
遷移不在服務啟動時自動執行。此帳本已依使用者授權備份並套用 `002`。

## `logs/`

此目錄由服務啟動時自動建立，不需要預先新增日誌檔案。

| 檔案 | 放了什麼 |
| --- | --- |
| `app-YYYY-MM-DD.log` | 設定載入、資料庫、記帳 API、Gemini 診斷及前端上報等一般日誌，每日分檔並同步輸出至終端機。 |
| `gin-YYYY-MM-DD.log` | Gin 的 HTTP 請求、路由除錯及 panic recovery 日誌，一天一個檔案。 |
| `webhook-YYYY-MM-DD.log` | Webhook 事件處理、失敗原因、LINE 回覆結果與請求耗時，一天一個檔案。 |
| `gin.log` | 舊版留下的歷史日誌，更新並重啟後不再寫入；保留原檔。 |

`main.go` 將 `gin.DefaultWriter` 與 `gin.DefaultErrorWriter` 接到 Gin 每日輸出，
並把另一個獨立 logger 傳給 webhook handler。檔名日期以台灣時間（UTC+8）計算，
同一天追加寫入，跨日後的第一筆寫入自動開啟新檔案，保留所有歷史檔案。
路徑相對於服務啟動時的工作目錄；啟動時無法建立日誌則停止啟動，執行中寫入失敗則回退到終端機。
標準 `log` 經由 `io.MultiWriter` 同步輸出至終端機與 app 每日檔案，初始化位於讀取設定之前。
日誌工具本身使用獨立的終端機 logger 記錄錯誤，避免再寫入自身造成死鎖。

## `internal/config/`

| 檔案 | 放了什麼 |
| --- | --- |
| [config.go](../internal/config/config.go) | `Config` 結構、`Load()`、`Validate()` 與 `getEnv()`。載入 `.env`、取得環境變數及預設值，驗證必要的 LINE 設定，並記錄設定檔載入錯誤。 |

## `internal/line/`

| 檔案 | 放了什麼 |
| --- | --- |
| [types.go](../internal/line/types.go) | LINE webhook 請求、事件來源、文字訊息、postback 及其他事件的結構；也定義回覆用的 `Message` 與 `NewText()`。有定義型別的事件不一定有實作處理邏輯。 |
| [webhookVerifier.go](../internal/line/webhookVerifier.go) | `VerifySignature()`。以 channel secret 對原始請求內容計算 HMAC-SHA256，轉成 Base64 後與 `X-Line-Signature` 比對。 |
| [webhookEventHandler.go](../internal/line/webhookEventHandler.go) | `WebhookHandler()` 與 `handleEvent()`。讀取、驗證及解析 webhook，檢查白名單，將文字問題交給記帳分析服務，處理 postback，再透過 LINE 回覆。使用注入的 logger 寫入獨立每日日誌，記錄事件數、失敗原因、回覆結果及處理耗時。 |
| [sendMessages.go](../internal/line/sendMessages.go) | LINE API client。`NewClient()` 建立 client，`Reply()` 回覆訊息，`Push()` 主動推送，`post()` 共用 JSON 編碼、HTTP 請求與回應檢查。 |
| [mealReminder.go](../internal/line/mealReminder.go) | 台灣時間每天 20:00 檢查三餐紀錄。缺少任一餐時，透過既有 `Push()` 私訊每位 LINE 白名單使用者，列出缺少餐別。排程由 `main.go` 啟動，結束時取消並等待正在處理的提醒結束。 |

目前文字分析脫離 webhook 請求的取消訊號，但仍有 40 秒總上限。
Gemini 單次 HTTP 呼叫上限為 20 秒；LINE API client 的 HTTP 呼叫上限為 15 秒。

## `internal/logging/`

| 檔案 | 放了什麼 |
| --- | --- |
| [dailyWriter.go](../internal/logging/dailyWriter.go) | `DailyWriter` 與 `NewDailyWriter()`。建立以台灣日期命名的追加日誌，寫入時檢查是否需要跨日切檔，以 mutex 保護同時寫入，並處理關閉與寫入失敗時的終端機備援。 |

## `internal/record/`

| 檔案 | 放了什麼 |
| --- | --- |
| [model.go](../internal/record/model.go) | `Record`、`SpendCategory` 與 `LineUserWhitelist` 資料結構，以及 API JSON 欄位名稱。 |
| [db.go](../internal/record/db.go) | `InitDB()` 先自動建立資料庫父目錄，再開啟 SQLite 連線；包含三張資料表的建表 SQL、記帳分類外鍵與索引，以及分類表為空時寫入的 13 個預設分類。透過 DSN 為每個連線啟用外鍵檢查；建立目錄或初始化失敗會回傳錯誤，由主程式記錄。 |
| [handler.go](../internal/record/handler.go) | 記帳新增、查詢、修改、刪除，分類查詢與新增，以及 LINE 白名單驗證的 Gin handler。包含參數解析、SQL 執行、結果讀取與錯誤日誌。 |
| [mealReminder.go](../internal/record/mealReminder.go) | 查詢指定日期的共用帳本，依分類名稱完全符合「早餐」、「午餐」、「晚餐」判斷各餐是否至少有一筆，回傳缺少的餐別；不依作者、tag 或金額篩選。 |
| [analyzer.go](../internal/record/analyzer.go) | Gemini 問答服務。`Answer()` 處理問題與使用者提示；`analyze()` 規劃查詢並整理答案；`generate()` 呼叫 Gemini；`analysisPlanSchema()` 定義模型須回傳的查詢格式。 |
| [analysisQuery.go](../internal/record/analysisQuery.go) | 問答用的查詢條件、統計與明細結果結構。驗證日期、分類、排序與筆數，組合參數化唯讀 SQL，查詢完整總計、分組統計與明細。 |

### 記帳分析的分工

1. `analyzer.go` 取得分類，讓 Gemini 將使用者問題轉成查詢條件。
2. `analysisQuery.go` 驗證條件，由程式組合 SQL 並計算查詢結果。
3. `analyzer.go` 把結果交給 Gemini，產生繁體中文回答。
4. `webhookEventHandler.go` 將回答或失敗提示送回原 LINE 聊天室。

查詢最多四組條件，每組分組結果最多 100 組，明細最多 50 筆；總計不因明細上限而截斷。
金額轉為原始單位百分之一的整數後計算。相對日期由模型依台灣當日日期解讀。
目前沒有專門逐日補齊日期、列出未記帳日的查詢流程，也沒有保存對話歷史。

## 每日記帳提醒

後端服務運行期間，每天台灣時間（UTC+8）20:00 檢查 `records.date` 為當日的共用帳本。
透過 `records.category_id` 對應 `spendCategories.name`，早餐、午餐、晚餐各至少一筆才不提醒。
缺少任一餐時，私訊 `lineUserWhitelist` 中的每位使用者，例如「今天（2026-11-19）還沒記錄：早餐、晚餐，記得記帳。」。
查詢失敗不發送提醒；個別推播失敗會記錄日誌並繼續處理其他收件者。
日誌寫入 `app-YYYY-MM-DD.log`，以 `[meal-reminder]` 標記。
20:00 服務未運行就不會執行；20:00 後啟動則從隔天開始，不補發。
排程在每個服務程序各自執行，應只運行一個實例，避免同一時間重複提醒。

## `internal/web-frontend/`

前端是由 Go 直接提供的 HTML、CSS 與 JavaScript，頁面的樣式與大部分互動邏輯放在各 HTML 內。

| 檔案 | 放了什麼 |
| --- | --- |
| [index.html](../internal/web-frontend/index.html) | 首頁，提供記帳工具與每月統計連結。沒有登入狀態時，將原網址編碼成 `returnTo` 並導向登入頁；已登入時留在首頁。 |
| [liffLogin.html](../internal/web-frontend/liffLogin.html) | LINE 登入與白名單驗證頁。取得 LIFF 設定、初始化 SDK、取得 LINE profile，再向後端驗證白名單；成功後保存 `lineUserId`，並返回已驗證的原始目的地。包含一次自動登入、手動重試與失敗階段日誌。 |
| [spendInput.html](../internal/web-frontend/spendInput.html) | 記帳操作頁。沒有登入狀態時帶上 `returnTo` 導向登入頁。包含日期切換、當日紀錄、分類選取與新增、金額／備註／tag 輸入，以及紀錄新增、編輯、刪除和操作提示。分類名稱與圖示必填，圖示自行輸入；分類選取時四邊框與底色變色。tag 有尼、拿與房屋圖示，編輯帶入原 tag。金額初始及儲存後為 0，點入清除預設 0，空白離開恢復 0。成功新增或更新後依最近使用順序排列分類，同一瀏覽器依帳號保存。新增時傳送登入者 `lineUserId`，編輯保留原作者；分類載入失敗時提供重試按鈕。 |
| [statistics.html](../internal/web-frontend/statistics.html) | 每月統計頁，由首頁進入。提供月份、支出／收入／結餘與圓餅圖／長條圖切換、分類占比及金額排序明細。透過 `/api/records?month=YYYY-MM` 讀取指定月份；`income` 為收入，其餘分類為支出。結餘圖比較收支，總額顯示收入減支出。金額以分加總；負數改以長條圖呈現，標示實際正負值。 |
| [errorLogging.js](../internal/web-frontend/errorLogging.js) | 前端頁面共用的 API 檢查與日誌工具。`requestJSON()` 檢查 HTTP、JSON 與資料格式；`logError()` 去除重複上報；`sendFrontendLog()` 統一送至後端，並監聽未處理的 JavaScript、Promise 與資源載入錯誤。 |

### 登入後如何回到原本頁面

1. 入口頁將原本的路徑、查詢參數與 fragment 編碼為登入網址的 `returnTo`。
2. 登入頁確認目的地屬於本站已知頁面，並保存至 sessionStorage 的 `liffReturnTo`。
3. 呼叫 `liff.login()` 時將 `returnTo` 放進 `redirectUri`，同分頁保存的資料也可作為回呼備援。
4. 新登入或已保存登入狀態的白名單驗證通過後，返回原頁面，清除這次登入的目的地與嘗試標記。

目的地白名單在 `liffLogin.html` 的 `allowedReturnPaths`，目前包含 `/`、`/index.html`、`/spendInput.html` 與 `/statistics.html`。
外部網址、登入頁及未知路徑不接受；缺少有效目的地時顯示錯誤，不自行選擇另一個頁面。
新增前端頁面時，也要加入白名單並在入口傳遞 `returnTo`。

### 前端錯誤如何到後端

```text
頁面操作／API 請求／全域錯誤事件
  -> errorLogging.js 的 logError()
  -> sendFrontendLog()
  -> POST /api/log
  -> main.go 以 [frontend-log] 同步寫入 app 每日日誌與終端機
```

日誌包含錯誤訊息與頁面路徑；API 錯誤另包含請求方法、路徑及 HTTP 狀態，不上報請求本文。
日誌 API 自己失敗時只留在瀏覽器 Console，避免循環上報。前端上報內容寫入 app 每日日誌；Gin 與 webhook 各自每日分檔。

## HTTP 路由索引

所有路由都在 `main.go` 註冊。

| 方法 | 路徑 | 處理位置與用途 |
| --- | --- | --- |
| `POST` | `/webhook` | `line/webhookEventHandler.go`：接收 LINE 事件。 |
| `GET` | `/api/records` | `record/handler.go` 的 `List()`：查詢紀錄，可用 `date=YYYY-MM-DD` 或 `month=YYYY-MM` 篩選；兩者同時提供時以 `date` 為優先。月份格式錯誤回傳 400。 |
| `POST` | `/api/records` | `record/handler.go` 的 `Create()`：新增紀錄，必須提供白名單內的 `lineUserId`，儲存為原作者；缺少 ID 回傳 400，未授權回傳 403。 |
| `PUT` | `/api/records/:id` | `record/handler.go` 的 `Update()`：修改紀錄，保留原始 `lineUserId`，忽略請求中的作者欄位。 |
| `DELETE` | `/api/records/:id` | `record/handler.go` 的 `Delete()`：刪除紀錄。 |
| `GET` | `/api/spend-categories` | `record/handler.go` 的 `ListSpendCategories()`：取得分類。 |
| `POST` | `/api/spend-categories` | `record/handler.go` 的 `CreateSpendCategory()`：新增分類。 |
| `POST` | `/api/validate-line-userId` | `record/handler.go` 的 `ValidateLineUserID()`：查詢是否在白名單。 |
| `POST` | `/api/log` | `main.go`：接收前端日誌，同步寫入 app 每日日誌與終端機。 |
| `GET` | `/api/liff-config` | `main.go`：回傳前端 LIFF ID。 |
| `GET` | `/`、`/index.html` | `main.go` 的靜態檔案服務：提供首頁。 |
| `GET` | `/liffLogin.html`、`/spendInput.html`、`/errorLogging.js` | `main.go` 的靜態檔案服務：提供登入、記帳頁面與共用 JavaScript。 |
| `GET` | `/statistics.html` | `main.go` 的靜態檔案服務：提供每月統計頁。 |

## 修改功能時可先找哪裡

| 想修改的功能 | 優先查看 |
| --- | --- |
| 環境變數或預設模型 | `internal/config/config.go`。 |
| HTTP 路由與服務埠的使用 | `main.go`。 |
| 日誌檔名、每日切檔與寫入方式 | `internal/logging/dailyWriter.go`；分類與接線位置在 `main.go`。 |
| LINE 訊息收到後要做什麼 | `internal/line/webhookEventHandler.go`。 |
| LINE 簽章驗證或回覆 API | `internal/line/webhookVerifier.go`、`internal/line/sendMessages.go`。 |
| Gemini 提示詞、回覆方式或額度處理 | `internal/record/analyzer.go`。 |
| 問答可使用的查詢條件與統計方式 | `internal/record/analysisQuery.go`。 |
| 資料表與初始分類 | `internal/record/db.go`，資料結構另見 `model.go`。 |
| 既有記帳表的分類外鍵遷移 | `migrations/001_records_category_foreign_key.sql`。 |
| 記帳與分類 HTTP API | `internal/record/handler.go`。 |
| 記帳頁面的版面或操作 | `internal/web-frontend/spendInput.html`。 |
| 每月統計圖表與明細 | `internal/web-frontend/statistics.html`；月份查詢位於 `internal/record/handler.go`。 |
| 每日三餐提醒時間與推播 | `internal/line/mealReminder.go`；啟動位置為 `main.go`。 |
| 三餐完整性判斷 | `internal/record/mealReminder.go`。 |
| 記帳者欄位與既有作者遷移 | `internal/record/model.go`、`db.go`、`handler.go` 與 `migrations/002_records_line_user_id.sql`。 |
| LINE 前端登入 | `internal/web-frontend/liffLogin.html`。 |
| 前端錯誤收集與上報 | `internal/web-frontend/errorLogging.js`，後端接收位置在 `main.go`。 |
