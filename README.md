# Purrtopia

Discord Bot for Maplestory game community.

## Features

### 冰雪季活動 `/冰雪季`
- **代購服務** - 登記想購買的冰雪市場道具
- **代售服務** - 登記想出售的冰雪市場道具
- **查看列表** - 查詢目前所有代購/代售資料
- 支援 **亞服** 與 **台港澳服** 分區資料

### 頭像收集（暫停使用）
- ~~`/抽頭` - 隨機抽取頭像（加權機率）~~
- ~~`/上傳頭顱` - 上傳遊戲頭像~~
- ~~`/展示` - 展示收集到的頭像~~

## Project Structure

```
purrtopia/
├── cmd/bot/
│   └── main.go                 # Entry point
├── internal/
│   ├── auth/                   # 權限檢查
│   ├── bot/                    # Bot 核心邏輯
│   ├── commands/               # 指令處理
│   │   ├── commands.go         # 指令註冊
│   │   ├── gacha/              # 頭像收集相關（暫停）
│   │   │   ├── draw_head.go
│   │   │   ├── show_collection.go
│   │   │   └── upload_avatar.go
│   │   └── snow/               # 冰雪季活動
│   │       └── snow_season.go
│   ├── component/              # UI 元件 (Button, Select, Modal)
│   ├── config/                 # 設定管理
│   ├── database/               # 資料庫連線與 Schema
│   └── embed/                  # Embed 訊息建構
├── docker-compose.yml
├── Dockerfile
└── Dockerfile.dev
```

## Quick Start

```bash
# 複製環境變數
cp .env.example .env
# 編輯 .env 填入 DISCORD_TOKEN

# 開發環境（hot reload）
docker compose --profile dev up

# 正式環境
docker compose --profile prod up

# 查看日誌
docker compose logs -f

# 停止
docker compose down
```

## Environment Variables

| 變數 | 必填 | 說明 |
|------|------|------|
| `DISCORD_TOKEN` | Yes | Discord Bot Token |
| `GUILD_IDS` | No | 伺服器 ID（逗號分隔，支援多伺服器） |
| `BOT_OWNER_IDS` | No | Bot 擁有者 ID（逗號分隔） |
| `BOT_ADMIN_IDS` | No | Bot 管理員 ID（逗號分隔） |
| `DB_HOST` | No | 資料庫主機（預設 localhost） |
| `DB_PORT` | No | 資料庫埠號（預設 3306） |
| `DB_USER` | No | 資料庫使用者 |
| `DB_PASSWORD` | No | 資料庫密碼 |
| `DB_NAME` | No | 資料庫名稱 |

## Permission System

```go
import "purrtopia/internal/auth"

// 檢查權限等級
perm := auth.CheckPermission(s, guildID, userID)

if perm == auth.PermissionBotOwner {
    // Bot 擁有者
}

// 檢查是否有指定等級以上權限
if auth.HasPermission(s, guildID, userID, auth.PermissionBotAdmin) {
    // BotAdmin 或 BotOwner
}
```

權限等級（高到低）：
- `PermissionBotOwner` - Bot 擁有者
- `PermissionBotAdmin` - Bot 管理員
- `PermissionServerAdmin` - 伺服器管理員
- `PermissionNone` - 無特殊權限
