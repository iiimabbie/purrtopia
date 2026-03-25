# Purrtopia

心動小鎮 (Heartopia) Discord Bot

## Features

### 潮流季 `/潮流季`
- **代售服務** - 登記本週可代售的物品
- **搜尋代售** - 查詢有誰可以幫你代售
- **代售總覽** - 查看所有代售登記
- 支援 **亞服** 與 **台港澳服** 分區
- 物品清單從資料庫讀取，換季只需更新 `season_items` 表

### 揪團 `/揪團`
- 建立揪團活動（彩虹花束、釣魚等）
- 自動開啟討論串
- 報名/取消報名

### AI 功能
- Gemini AI 整合（主要 + Fallback API）

## Tech Stack

- **語言**: Go 1.23
- **Discord**: discordgo v0.29.0
- **資料庫**: PostgreSQL + GORM
- **部署**: Docker Compose

## Project Structure

```
purrtopia/
├── cmd/bot/
│   └── main.go                 # Entry point
├── internal/
│   ├── ai/                     # Gemini AI 整合
│   ├── auth/                   # 權限檢查
│   ├── bot/                    # Bot 核心邏輯
│   ├── commands/               # 指令處理
│   │   ├── commands.go         # 指令註冊
│   │   ├── group/              # 揪團功能
│   │   └── season/             # 潮流季功能
│   ├── component/              # UI 元件 (Button, Select, Modal)
│   ├── config/                 # 設定管理
│   ├── database/               # 資料庫連線與 Repository
│   └── embed/                  # Embed 訊息建構
├── docker-compose.yml
└── Dockerfile
```

## Quick Start

```bash
# 複製環境變數
cp .env.example .env
# 編輯 .env 填入 DISCORD_TOKEN 和資料庫連線資訊

# 啟動
docker compose up -d --build
```

## Environment Variables

| 變數 | 必填 | 說明 |
|------|------|------|
| `DISCORD_TOKEN` | Yes | Discord Bot Token |
| `GUILD_IDS` | No | 伺服器 ID（逗號分隔） |
| `BOT_OWNER_IDS` | No | Bot 擁有者 ID |
| `BOT_ADMIN_IDS` | No | Bot 管理員 ID |
| `DB_HOST` | No | PostgreSQL 主機（預設 localhost） |
| `DB_PORT` | No | PostgreSQL 埠號（預設 5432） |
| `DB_USER` | No | 資料庫使用者 |
| `DB_PASSWORD` | No | 資料庫密碼 |
| `DB_NAME` | No | 資料庫名稱 |
| `GEMINI_API_KEY` | No | Gemini AI API Key |
| `GEMINI_MODEL` | No | Gemini 模型（預設 gemini-2.0-flash） |

## Season Items 管理

物品清單存在 `season_items` 表，換季用 SQL 更新：

```sql
-- 清空舊季
TRUNCATE season_items RESTART IDENTITY;

-- 新增物品（category: 1=第一組選單, 2=第二組選單）
-- 單一選單上限 25 個，超過就分組
INSERT INTO season_items (name, category) VALUES
('物品A', 1),
('物品B', 1),
('物品C', 2);
```
