# Docker 常用指令

## 啟動/停止

```bash
# 啟動（開發環境，有 hot reload）
docker compose --profile dev up -d

# 啟動（正式環境）
docker compose --profile prod up -d

# 停止（保留資料）
docker compose down

# 停止並刪除資料（危險！）
docker compose down -v
```

## 查看狀態

```bash
# 查看運行中的容器
docker ps

# 查看所有容器（包含已停止）
docker ps -a

# 查看 logs
docker compose logs -f              # 全部
docker compose logs discord-bot -f  # 只看 bot
docker compose logs mysql -f        # 只看 mysql

# 查看 logs 最後 N 行
docker compose logs --tail 50
```

## 重啟/重建

```bash
# 重啟容器
docker compose restart

# 重建並啟動（程式碼有改動時）
docker compose --profile prod up -d --build

# 只重建某個服務
docker compose --profile prod up -d --build discord-bot
```

## 進入容器

```bash
# 進入 bot 容器
docker exec -it purrtopia-discord-bot sh

# 進入 mysql 容器
docker exec -it purrtopia-mysql bash

# 直接執行 mysql 指令
docker exec -it purrtopia-mysql mysql -upurrtopia -pchangeme purrtopia
```

## 清理（釋放空間）

```bash
# 安全清理（不刪資料）
docker system prune -a

# 只清理未使用的映像
docker image prune -a

# 只清理停止的容器
docker container prune

# 查看佔用空間
docker system df
```

## 危險指令（會刪資料）

```bash
# 刪除未使用的 volume（資料會不見！）
docker volume prune

# 全部清理包含 volume（資料會不見！）
docker system prune -a --volumes
```

## Volume（資料儲存）

```bash
# 查看所有 volume
docker volume ls

# 查看特定 volume 詳情
docker volume inspect purrtopia_mysql_data
```

## 備份與還原

### 自動備份

`mysql-backup` 服務會自動：
- 每天凌晨 3 點備份
- 備份檔存放在 `./backups/` 資料夾
- 使用 gzip 壓縮

```bash
# 查看備份檔
ls -la ./backups/
```

### 手動備份

```bash
# 備份到指定檔案
docker exec purrtopia-mysql mysqldump -upurrtopia -pchangeme purrtopia > backup.sql

# 備份並壓縮
docker exec purrtopia-mysql mysqldump -upurrtopia -pchangeme purrtopia | gzip > backup_$(date +%Y%m%d).sql.gz
```

### 還原資料

```bash
# 從 .sql 檔還原
docker exec -i purrtopia-mysql mysql -upurrtopia -pchangeme purrtopia < backup.sql

# 從 .sql.gz 壓縮檔還原
gunzip < ./backups/db_backup_xxxxx.sql.gz | docker exec -i purrtopia-mysql mysql -upurrtopia -pchangeme purrtopia
```

### 資料不見時

1. 先確認有沒有備份檔：`ls ./backups/`
2. 找到最新的備份檔
3. 執行上面的還原指令

## 常見問題

### 容器一直重啟
```bash
# 查看錯誤訊息
docker compose logs --tail 100
```

### Port 被佔用
```bash
# 查看誰佔用 3306
lsof -i :3306
# 或
netstat -tlnp | grep 3306
```

### 映像太舊
```bash
# 強制重新拉取並建構
docker compose --profile prod build --no-cache
docker compose --profile prod up -d
```
