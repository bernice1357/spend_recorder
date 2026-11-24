# cash_carole_record

使用 Go、Gin 與 SQLite 的共用記帳工具，提供 LINE 登入、網頁記帳、每月統計、Gemini 記帳問答，以及每日三餐記帳提醒。
詳細架構、檔案職責與資料流程請見 [專案工作文件](docs/README.md)。

## 功能

| 功能 | 行為 |
| --- | --- |
| 網頁記帳 | 依日期新增、查詢、編輯與刪除紀錄，可填金額、備註與 tag。 |
| 分類 | 可新增分類，名稱與圖示皆必填；圖示由使用者輸入或貼上 emoji。 |
| 最近使用分類 | 成功新增或更新後，該分類移到最上方，其餘已使用分類保留最近使用順序；同一瀏覽器依帳號保存。 |
| 每月統計 | 切換月份與支出／收入／結餘，提供環形圓餅圖、長條圖、分類占比及金額排序明細。 |
| LINE 記帳問答 | 白名單使用者可詢問共用帳本，由 Gemini 解讀條件並根據程式查詢結果回答。 |
| 每日提醒 | 台灣時間每天 20:00 檢查共用帳本，缺少早餐、午餐或晚餐任一類紀錄時，私訊白名單使用者。 |

金額初始為 `0`，點入欄位時清除預設 `0`，空白離開則恢復 `0`；儲存成功後也重設為 `0`。
分類選取時，四邊框線與底色會變色。tag 表示這筆由誰出錢，提供「尼」、「拿」（該人代墊）與房屋圖示（家裡出錢），單選，再次點選可取消；未選表示記帳者自己出錢。舊資料的多個 tag 以逗號分隔存於同一欄位（例如 `尼,午餐`），編輯時帶入原 tag。三餐改用「早餐」、「午餐」、「晚餐」分類記錄；舊紀錄上的三餐 tag 仍會顯示並保留。
點選一筆紀錄即進入修改狀態，可更新、取消或刪除。點日期可開啟月曆跳到指定日期。

## 目錄結構

```text
cash_carole_record/
├── main.go                         # 初始化服務、API 與每日提醒
├── go.mod
├── go.sum
├── .env.example                    # 空白設定範本
├── .env                            # 本機設定，不提交 Git
├── .gitignore                      # 忽略本機設定、資料庫、備份與日誌
├── docs/
│   └── README.md                   # 架構、檔案職責與流程索引
├── migrations/
│   ├── 001_records_category_foreign_key.sql
│   ├── 002_records_line_user_id.sql
│   └── 003_whitelist_display_name.sql
├── scripts/
│   └── sql/backfill-display-name/  # 可選的一次性 SQL：補填記帳者名稱
├── data/
│   ├── records.db                  # SQLite 資料庫
│   └── backups/                    # 資料庫備份
├── logs/                           # 每日一般、Gin 與 webhook 日誌
└── internal/
    ├── config/
    │   └── config.go               # 環境變數載入與必要欄位驗證
    ├── line/
    │   ├── webhookEventHandler.go  # Webhook 與事件處理
    │   ├── webhookVerifier.go      # LINE 簽章驗證
    │   ├── sendMessages.go         # Reply / Push client
    │   ├── mealReminder.go         # 每日提醒排程與私訊
    │   └── types.go
    ├── logging/
    │   └── dailyWriter.go          # 台灣日期分檔日誌
    ├── record/
    │   ├── db.go                   # SQLite 建表與預設分類
    │   ├── model.go                # 資料模型
    │   ├── handler.go              # 記帳、分類與白名單 API
    │   ├── mealReminder.go         # 共用帳本三餐完整性查詢
    │   ├── analyzer.go             # Gemini 問答
    │   └── analysisQuery.go        # 條件驗證與唯讀查詢
    └── web-frontend/
        ├── index.html             # 記帳與統計入口
        ├── liffLogin.html         # LINE 登入與返回頁面
        ├── spendInput.html        # 記帳頁
        ├── statistics.html        # 每月統計
        └── errorLogging.js        # API 檢查與前端錯誤上報
```

## 從零開始安裝

### 1. 安裝 Go

本專案的 `go.mod` 宣告 Go `1.26.0`，請安裝 Go `1.26.0` 或更新版本。
到 [Go 官方下載頁](https://go.dev/dl/) 選擇符合電腦系統與處理器架構的安裝檔：

| 系統 | 安裝方式 |
| --- | --- |
| macOS | Apple Silicon（M 系列）選 `darwin-arm64.pkg`，Intel 選 `darwin-amd64.pkg`；開啟 `.pkg` 並依安裝精靈完成安裝。 |
| Windows | 一般 Intel／AMD 電腦選 `windows-amd64.msi`，ARM 電腦選 `windows-arm64.msi`；開啟 `.msi` 並依安裝精靈完成安裝。 |

安裝完成後，關閉並重新開啟終端；若使用 VS Code 的整合終端，也先重新開啟 VS Code。
macOS 使用 Terminal，Windows 以下使用 PowerShell。兩個系統都執行：

```text
go version
```

確認有顯示 `go version go1.26.0` 或更新版本，尾端的系統與架構會依電腦不同。
若顯示找不到 `go` 指令，確認安裝已完成並重新開啟終端；仍無法執行時，
依 [Go 官方安裝說明](https://go.dev/doc/install) 確認安裝與 PATH 設定。

### 2. 取得專案並進入根目錄

下載並解壓縮專案原始碼，或使用已取得的專案資料夾。
在含有 `go.mod`、`main.go`、`.env.example` 的 `cash_carole_record` 根目錄開啟終端。
可以在 VS Code 開啟此資料夾後，選擇「終端機 → 新增終端機」。
後續指令都在這個根目錄執行；不要在 `internal/` 或 `docs/` 內執行。

### 3. 下載專案依賴套件

macOS 與 Windows 都執行：

```text
go mod download
```

此指令需要網路連線，會下載 `go.mod` 指定的依賴套件，第一次可能需要一些時間。
不需要逐個安裝 Gin、godotenv 或 SQLite 套件，也不需要重新執行 `go mod init`。
完成且沒有錯誤後，繼續設定環境變數。

## 4. 環境設定

從範本建立 `.env`；如果已有 `.env`，保留原檔並直接編輯，不要覆蓋。
`data/` 目錄會在啟動時自動建立，不需手動建立。

macOS：

```bash
cp .env.example .env
```

Windows PowerShell：

```powershell
Copy-Item .env.example .env
```

填入 `.env` 中的設定值。`.env.example` 的值全部留空；程式已存在的系統環境變數優先於 `.env`。

| 變數 | 必填情況 | 說明 |
| --- | --- | --- |
| `PORT` | 否 | HTTP 監聽埠；未設定或為空時使用 `8080`。 |
| `LINE_CHANNEL_SECRET` | 啟動必填 | Messaging API channel 的 secret，用於驗證 webhook 簽章。 |
| `LINE_CHANNEL_ACCESS_TOKEN` | 啟動必填 | Messaging API channel 的 access token，用於回覆與私訊推播。 |
| `LIFF_ID` | 前端登入必填 | LIFF 應用程式 ID，前端從 `/api/liff-config` 取得；未設定時登入頁顯示錯誤。 |
| `GEMINI_API_KEY` | 記帳問答需要 | 未設定時，LINE 問答回覆設定提示。 |
| `GEMINI_MODEL` | 否 | 程式預設為 `gemini-3.5-flash-lite`；此變數可自行加入 `.env`。 |
| `PUBLIC_BASE_URL` | 否 | 對外 HTTPS 網址。設定後，三餐提醒會附「去記帳」按鈕，開啟當天記帳頁並預選缺少餐別的分類；未設定時只傳文字。 |

`.gitignore` 忽略 `.env`、`data/`、`logs/` 與 `.DS_Store`，`.env.example` 保留供提交。
建表程式與 `migrations/` 內的 SQL 應提交 Git；實際資料庫與備份不提交。
已被 Git 追蹤的檔案不會因新增忽略規則而自動移除。

## 5. 啟動與部署

完成上述步驟並填妥啟動必填設定後，macOS 與 Windows 都在專案根目錄啟動後端：

```text
go run .
```

第一次啟動會編譯程式，需等待服務顯示監聽訊息；終端保持開啟，按 `Ctrl+C` 可停止服務。
服務監聽 `PORT` 指定的埠；未設定才使用 `8080`。
在瀏覽器開啟 `http://localhost:8080/`；若 `.env` 設定其他 `PORT`，請將網址中的 `8080` 換成該值。
前端仍需完成下方的 LINE／LIFF 設定與白名單授權，才可登入使用。

本地需要公開 HTTPS 網址時，可在已安裝並完成設定的 ngrok 中，另開終端執行：

```bash
ngrok http <PORT>
```

請將 `<PORT>` 換成後端實際使用的埠，並讓前端與 `/api/` 請求連到同一個 Go 服務。
修改 Go 程式後需重新啟動服務。

`main.go` 以相對路徑 `data/records.db` 開啟資料庫：

- `data/` 已存在、資料庫不存在：SQLite 建立檔案，程式建立資料表與預設分類。
- 資料庫已存在：沿用既有紀錄，不清空；不會自動套用遷移。
- `data/` 不存在：`InitDB()` 先自動建立目錄，再建立資料庫、資料表與預設分類。
- 目錄因權限不足等原因無法建立：回傳錯誤並停止啟動。

資料庫與日誌路徑都相對於服務啟動時的工作目錄。初始化日誌會列出資料庫絕對路徑與分類筆數。
新資料庫的白名單為空，須先加入允許使用的 LINE userId。

## LINE 設定與登入

在 LINE Developers 設定 Messaging API webhook 為 `https://<你的網域>/webhook`，開啟 webhook，
並將該 channel 的 secret 與 access token 填入環境設定。服務需能回應 LINE 的 webhook 驗證請求。
LIFF 的 `LIFF_ID` 則用於前端 LINE 登入；它是應用程式 ID，不是個人的使用者 ID。

圖文選單可開啟 `/spendInput.html` 或 `/statistics.html`。入口頁未保存登入狀態時，
將原路徑、查詢參數與 fragment 編碼到登入頁的 `returnTo`。
登入頁保存同分頁目的地並傳入 LINE 登入回呼網址；取得 profile 後，以 `profile.userId` 檢查白名單，
成功後保存 `lineUserId` 並返回原頁面。已保存 ID 時，登入頁會先重新驗證白名單。

允許返回的路徑為 `/`、`/index.html`、`/spendInput.html` 與 `/statistics.html`。
外部網址、登入頁本身及未知路徑不接受。沒有有效目的地時會顯示提示，請從入口連結重新開啟。
新增前端頁面時，需同步更新 `liffLogin.html` 的 `allowedReturnPaths`，並帶上 `returnTo`。
首頁提供記帳與統計入口，已登入者不會自動跳到記帳頁。

## 資料庫與記帳者

| 資料表 | 用途 |
| --- | --- |
| `records` | 記帳內容與原始記帳者 `line_user_id`。 |
| `spendCategories` | 分類 ID、名稱及圖示。 |
| `lineUserWhitelist` | 可使用服務的 LINE userId 與加入時間。 |

新增記帳時，前端將登入後保存的 `lineUserId` 傳入 API，後端要求非空並檢查白名單，
儲存為 `records.line_user_id`。編輯只修改記帳內容，保留原作者。
目前白名單檢查依前端傳送的 ID 進行，不等於驗證 LINE 登入憑證。
查詢、統計、Gemini 問答與每日三餐檢查仍使用共用帳本，不依作者分開。
`income` 分類視為收入，其餘分類視為支出。
記帳頁每筆紀錄會標示記帳者，名稱來自 `lineUserWhitelist.display_name`，於使用者從 LINE 登入時自動更新。

`records.category_id` 參照 `spendCategories.id`，使用 `ON UPDATE CASCADE ON DELETE RESTRICT`：
分類 ID 更新會同步更新記帳，仍被使用的分類不能刪除。分類欄位有 `idx_records_category_id` 索引。
`InitDB()` 透過 DSN 的 `_pragma=foreign_keys(1)` 為每個新 SQLite 連線啟用外鍵檢查。

### 既有資料庫遷移

新資料庫由 `db.go` 直接建立完整結構，不需套用下列遷移。
既有資料庫須先備份，再依目前結構選擇尚未執行的遷移；服務啟動時不自動執行。

| SQL | 適用情況與行為 |
| --- | --- |
| `001_records_category_foreign_key.sql` | 尚未有分類外鍵時，重建記帳表、保留原始 ID 與遞增序號，加入分類外鍵與索引。 |
| `002_records_line_user_id.sql` | 尚未有 `line_user_id` 時新增欄位，並將既有紀錄作者填為此帳本擁有者先前指定的 ID；此值已寫在 SQL 內。 |
| `003_whitelist_display_name.sql` | 尚未有 `lineUserWhitelist.display_name` 時新增欄位，用於標示記帳者。新版程式需要此欄位，部署前必須先執行。 |

兩份都需要時，先執行 `001` 再執行 `002`。不要重複執行；`001` 的舊版重建結構不包含作者欄位，
不可在已新增作者欄位後直接套用。使用其他帳本時，先確認 `002` 的作者值符合資料歸屬。
執行前檢查孤立分類紀錄、額外欄位、索引、trigger、view 與其他表參照；遇錯應停止並回滾，
提交前核對資料與結構。更完整的檔案職責見 [專案工作文件](docs/README.md)。

### 補填記帳者名稱（可選）

`scripts/sql/backfill-display-name/` 內的 SQL 不是遷移，不需每個環境執行。
名稱同步修正部署後，使用者重新開啟頁面會自動補上名稱；
若想在部署前先讓正式機顯示記帳者名稱，可手動執行：

1. `sqlite3 data/records.db ".read scripts/sql/backfill-display-name/1_find_users.sql"`，
   依每個 userId 的紀錄判斷對應的使用者（只讀）。
2. 在伺服器上將 `2_set_names.sql` 的 `A_ID`／`B_ID` 與名稱換成實際值後，以同樣方式執行。
3. 執行後以 `git checkout -- scripts/sql/backfill-display-name/2_set_names.sql` 還原，
   不要提交含實際 userId 的檔案。

手動填入的名稱會在使用者下次完整登入時被 LINE 顯示名稱覆蓋。

## 每月統計

從首頁進入 `/statistics.html`，選擇月份及「支出／收入／結餘」，可切換環形圓餅圖與長條圖。
頁面透過 `/api/records?month=YYYY-MM` 讀取當月紀錄，以分為單位加總，列出分類占比及金額明細，
明細可依金額升冪或降冪排序。

結餘總額為收入減去支出，圖表比較收入與支出，占比以收支合計計算。
若分類合計含負數，改以長條圖呈現絕對值長度，並標示實際正負金額。
頁面提供無資料、載入失敗與重新載入狀態；快速切換月份時不讓舊回應覆蓋新結果。

## 每日記帳提醒

後端服務每天台灣時間（UTC+8）20:00，檢查當天共用帳本：

- 以 `records.date` 判斷日期，透過分類 ID 取得分類名稱。
- 當天「早餐」、「午餐」、「晚餐」分類（或舊資料的同名 tag）各至少一筆才不提醒；不依記帳者或金額篩選。
- 啟動時若資料庫缺少這三個分類會自動補建；已有同名分類則沿用，不重複建立。
- 缺少任一餐時，透過 LINE Push 私訊每位白名單使用者，列出缺少的餐別。
- 查詢失敗不發送；個別推播失敗則記錄日誌並繼續其他收件者。

20:00 服務未運行則不提醒；20:00 後才啟動，從隔天開始，不補發。
排程隨服務程序運作，應只運行一個實例，避免重複提醒。日誌以 `[meal-reminder]` 標記。

## Gemini 記帳問答

白名單使用者可在 LINE 詢問：

- 這個月總共花了多少錢？
- 這個月餐費比上個月多多少？
- 上個月支出最多的是哪個分類？
- 列出本月金額最大的三筆支出。
- 找出備註有「咖啡」的記帳紀錄。
- 目前有哪些記帳分類？

Gemini 將問題轉為結構化查詢條件，程式驗證條件並使用參數化唯讀 SQL 查詢，
再由 Gemini 根據結果整理回答。模型不能產生或執行 SQL。
帳本查詢使用 `records` 與 `spendCategories`，不提供模型查詢白名單。

相對日期使用台灣時間，日期起訖均包含當日，金額以小數點後兩位計算。
一次最多四組條件，每組統計涵蓋完整範圍，分組最多 100 組、明細最多 50 筆，
省略資料時要求模型說明。每次訊息獨立分析，不保留對話歷史。

一般查詢最多使用兩次 Gemini 呼叫。收到 HTTP 429 或 `RESOURCE_EXHAUSTED` 時，回覆額度或呼叫上限提示；
不定時查詢剩餘額度、不自動切換模型或重試。未設定金鑰或服務失敗時也會回覆提示。
使用者問題、分類與查詢結果會傳送至 Gemini。模型價格、額度與資料使用條款請查閱
[Google 官方定價](https://ai.google.dev/gemini-api/docs/pricing)，本專案不保證免費使用。

## API 與頁面

| 方法 | 路徑 | 用途 |
| --- | --- | --- |
| GET | `/`、`/index.html` | 記帳與統計入口。 |
| GET | `/liffLogin.html` | LINE 登入及白名單驗證。 |
| GET | `/spendInput.html` | 記帳頁。 |
| GET | `/statistics.html` | 每月統計頁。 |
| GET | `/errorLogging.js` | 前端共用 API 與錯誤上報工具。 |
| POST | `/webhook` | LINE webhook，檢查簽章與來源使用者白名單。 |
| GET | `/api/records` | 查詢紀錄；支援 `date=YYYY-MM-DD` 或 `month=YYYY-MM`，同時提供時優先使用 `date`。 |
| POST | `/api/records` | 新增紀錄；必須提供白名單內的 `lineUserId`。 |
| PUT | `/api/records/:id` | 修改紀錄，保留原作者。 |
| DELETE | `/api/records/:id` | 刪除紀錄。 |
| GET | `/api/spend-categories` | 查詢分類。 |
| POST | `/api/spend-categories` | 新增分類。 |
| POST | `/api/validate-line-userId` | 檢查 `userId` 是否在白名單。 |
| GET | `/api/liff-config` | 取得前端所需的 `liffId`。 |
| POST | `/api/log` | 接收前端日誌。 |

## 錯誤日誌

前端透過 `errorLogging.js` 將錯誤送至 `POST /api/log`，後端以 `[frontend-log]` 寫入一般日誌與終端機。
涵蓋 API 連線、HTTP 狀態、JSON 格式、載入與記帳操作、LINE 登入、資源載入及未處理的 JavaScript 錯誤。
日誌包含錯誤訊息與頁面路徑，API 錯誤另包含方法、API 路徑與 HTTP 狀態，不傳送 API 請求本文。
相同錯誤物件只上報一次；日誌 API 自己失敗時只輸出 Console，避免循環上報。

日誌 API 接收 `msg`（必填，最多 4096 字）及 `level`（`info`、`warn`、`error`，預設 `info`）。
前端請求必須連到此 Go 服務，才能送達後端。

| 檔案 | 內容 |
| --- | --- |
| `logs/app-YYYY-MM-DD.log` | 設定、資料庫、API、Gemini、每日提醒及前端上報日誌，同步輸出至終端機。 |
| `logs/gin-YYYY-MM-DD.log` | HTTP 請求、路由除錯與 panic recovery。 |
| `logs/webhook-YYYY-MM-DD.log` | Webhook 事件、簽章、白名單檢查、分析與 LINE 回覆結果及請求耗時。 |

檔名使用台灣日期，同一天追加寫入，跨日後第一筆寫入切換新檔；保留歷史檔案及舊 `gin.log`。
啟動時無法建立日誌會停止啟動，執行中寫入失敗則回退至終端機。
Webhook 日誌不記錄原始訊息本文、userId 或 token；SQLite 初始化會輸出資料庫路徑與分類筆數。
日誌寫檔工具本身的錯誤直接輸出至終端機，避免再次寫入失敗的日誌檔。

## LINE 事件處理

```text
LINE → POST /webhook → 驗證簽章 → 檢查使用者白名單
  → message(text)：Gemini 解讀問題 → 程式唯讀查詢 → Gemini 整理 → LINE Reply
  → postback：回覆 postback data

每天台灣時間 20:00 → 查詢共用帳本三餐 → 缺少任一餐 → LINE Push 私訊白名單使用者
```
