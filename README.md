# 钓鱼游戏后端服务

基于 Golang + Gin + Redis 实现的钓鱼游戏后端API服务，支持抽奖和榜单功能。

## 🚀 快速启动

### 使用 Docker Compose（推荐）

1. 克隆项目
```bash
git clone <your-repo>
cd fishing-game
```

2. 启动所有服务
```bash
docker-compose up -d
```

3. 查看服务状态
```bash
docker-compose ps
```

4. 查看日志
```bash
docker-compose logs -f backend
```

### 本地开发

1. 启动 Redis
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

2. 启动后端服务
```bash
cd backend
go run main.go
```

## 📡 API 接口

### 健康检查
```bash
curl http://localhost:8080/fishing/health
```

### 抽奖接口
```bash
# 执行抽奖
curl -X POST http://localhost:8080/fishing/lotteries/draw \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123"}'

# 查看抽奖历史
curl http://localhost:8080/fishing/lotteries/history/user123?limit=10
```

### 榜单接口
```bash
# 手动增加积分
curl -X POST http://localhost:8080/fishing/leaderboards/global_ranklist/scores/increment \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "delta": 100, "reason": "test"}'

# 查看排行榜
curl "http://localhost:8080/fishing/leaderboards/global_ranklist/top?page=1&page_size=10"

# 查看用户排名
curl http://localhost:8080/fishing/leaderboards/global_ranklist/users/user123
```

## 🎯 功能特性

- ✅ 权重抽奖系统（空军、小鱼、中鱼、大鱼、稀有鱼）
- ✅ 实时排行榜（基于Redis ZSET）
- ✅ 抽奖历史记录
- ✅ 容器化部署
- ✅ 健康检查

## 🏗️ 项目结构

```
fishing-game/
├── backend/                 # 后端服务
│   ├── main.go             # 入口文件
│   ├── config/             # 配置模块
│   ├── model/              # 数据模型
│   ├── service/            # 业务逻辑
│   ├── handler/            # HTTP处理器
│   ├── configs/            # 配置文件
│   └── Dockerfile          # 后端容器配置
├── docker-compose.yml      # 容器编排配置
└── README.md              # 项目说明
```

## 🛠️ 技术栈

- **后端**: Go 1.20 + Gin Framework
- **数据库**: Redis 7
- **容器化**: Docker + Docker Compose
- **架构**: 模块化设计，服务层分离

## 🔧 配置说明

### 奖池配置 (`backend/configs/lottery_pool.json`)
```json
{
  "items": [
    {"name": "空军", "weight": 5000, "points": 0},
    {"name": "小鱼", "weight": 3000, "points": 5},
    {"name": "中鱼", "weight": 1500, "points": 20},
    {"name": "大鱼", "weight": 400, "points": 100},
    {"name": "稀有鱼", "weight": 100, "points": 500}
  ]
}
```

### 环境变量
- `REDIS_ADDR`: Redis连接地址（默认: localhost:6379）

## 📊 数据存储

### Redis 数据结构
- `leaderboard:global_ranklist`: 全局排行榜 (ZSET)
- `lottery:draws:{user_id}`: 用户抽奖历史 (LIST)

## 🚦 服务管理

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f

# 进入容器
docker-compose exec backend sh
docker-compose exec redis redis-cli
```

## 📈 监控和调试

- 后端服务: http://localhost:8080/fishing/health
- Redis: `docker-compose exec redis redis-cli`
