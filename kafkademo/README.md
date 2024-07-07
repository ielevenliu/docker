# Kafka Docker 部署

- 执行 docker-compose up 或 docker-compose up -d 部署 docker compose 中的内容
- 打开 127.0.0.1:8080 kafka-ui，配置 192.168.1.6:9092(kafka外网访问IP) 关联 kafka 进程
- 登录 docker 容器，通过命令行创建 Topic 或在 kafka-ui 页面操作创建 Topic