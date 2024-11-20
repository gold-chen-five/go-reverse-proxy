# Go Reverse Proxy

### 概述
一個用 Go 語言實現的反向代理，支援SSL/TLS 自動更換憑證以及多種負載平衡策略。

### 特性
- 多伺服器配置
- SSL/TLS 憑證自動管理
- 多種負載平衡策略
  - 輪詢
  - 加權輪詢
  - 最少連接
  - IP 雜湊
- 基於路徑的路由
- 多上游伺服器支援
- 動態伺服器健康檢查
- 支援websocket
### 負載平衡策略
代理支援四種不同的負載平衡策略：

1. **輪詢** (`round-robin`)
   - 按順序將請求分配給伺服器池中的伺服器
   - 每個伺服器輪流接收請求
   - 適用於性能相近的伺服器
   ```yaml
   strategy:
     type: "round-robin"
   ```

2. **加權輪詢** (`weighted-round-robin`)
   - 類似輪詢，但考慮伺服器權重
   - 權重較高的伺服器收到更多請求
   - 適用於不同性能的伺服器配置
   ```yaml
   strategy:
     type: "weighted-round-robin"
     config:
       weights:
         "http://localhost:8081": 5
         "http://localhost:8082": 3
   ```

3. **最少連接** (`least-connections`)
   - 將流量導向活動連接數最少的伺服器
   - 根據伺服器響應時間自動平衡負載
   - 適用於處理時間不同的請求
   ```yaml
   strategy:
     type: "least-connections"
   ```

4. **IP 雜湊** (`ip-hash`)
   - 將客戶端 IP 固定映射到特定伺服器
   - 相同的客戶端 IP 始終指向相同的伺服器（如果可用）
   - 適用於需要會話保持的場景
   ```yaml
   strategy:
     type: "ip-hash"
   ```

### 配置指南
代理伺服器透過 `settings.yaml` 檔案進行配置。以下是配置結構的詳細說明：

```yaml
servers:
  - listen: ":443"              # 監聽連接埠
    ssl: true                   # 啟用 SSL/TLS
    host: "your.domain.com"     # 網域名稱
    routes:                     # 路由配置
      - match:                  # 路由匹配規則
          path: "/"             # 匹配路徑 ex: "/path"
        proxy:                  # 代理配置
          upstream:             # 上游伺服器列表
            - "http://localhost:8081"
            - "http://localhost:8082"
          strategy:             # 負載平衡策略
            type: "weighted-round-robin"  # 策略類型 
            config:             # 策略配置
              weights:          # 伺服器權重
                "http://localhost:8081": 5
                "http://localhost:8082": 3
  - listen: ":443"
    ssl: true
    host: "your2.domain.com"
    routes:
      - match:
          path: "/"
        proxy:
          upstream:
            - "http://localhost:8083"
          strategy:
            type: "round-robin"
```


### 啟動
所有配置在Makefile。

- 建立 exe，build 資料夾會有一個exe與setting.yaml，修改配置符合你的專案需求。
    ```
    make build
    ```
- 啟動exe

### 測試專案
- 下載golang
- 下載makefile
- 啟動測試伺服器 8081 8082 8083
    ```
    make test-server
    ```
- 根據setting.yaml啟動反向代理
    ```
    make dev
    make start (gcp/linux)
    ```    
- 傳送request和websocket到proxy，改變domain name 到example_server/client/test_client.go修改
    ```
    make test-client
    make test-client-ssl
    ```

