# go-file-server



基于Gin + Vue + Element UI 的前后端分离文件管理系统



## ✨ 特性
- Casbin的 RBAC 访问控制模型
- JWT 认证
- GORM 的数据库存储
- time/rate 令牌桶限速
- bleve 文件索引

## 🎁 内置
1. 文件管理：文件的增删改查
1. 用户管理：用户是系统操作者，该功能主要完成系统用户配置。
2. 部门管理：配置系统组织机构（公司、部门、小组），树结构展现支持数据权限。
3. 角色管理：角色菜单权限分配、设置角色按机构进行数据范围权限划分。
4. 操作日志：系统正常操作日志记录和查询；系统异常信息日志记录和查询。
5. 登录日志：系统登录日志记录查询包含登录异常。
6. 服务监控：查看一些服务器的基本信息。


## 📦 本地开发

### 环境要求

go 1.21

mysql 8.2

redis 5.6(可选，默认使用内存)




### 获取代码


```bash
git clone https://github.com/ctxgo/go-file-server.git
```

### 启动说明


```bash
# 进入项目目录
cd go-file-server

# 修改配置 
vi ./config/config.yml
# 更新整理依赖
go mod tidy

# 启动服务
go run main.go server -c ./config/config.yml
```

#### 构建docker镜像

```shell
# 编译镜像
docker build -t go-file-server .
```
<br>

## 初始用户
> 用户 admin

> 密码 123456

<br>

## 部署
### docker 部署
> 注意：修改 config.yaml，密码部分都是弱密码
```shell
cd deploy/docker
# 启动
docker-compose up -d
```


### helm 部署
> 注意：修改 config.yaml，密码部分都是弱密码

> 前提要求
- Helm 3

- Kubernetes 1.20+

#### 部署中间件
```shell
cd deploy/helm

# 添加中间件helm仓库
helm repo add bitnami https://charts.bitnami.com/bitnami

# 安装mysql
helm install mysql bitnami/mysql --version 9.5.1 --values mysql-9.5.1-values.yaml

#安装redis(可选)
helm install redis bitnami/redis --version 17.15.4 --values redis-17.15.4-values.yaml
```

#### 部署app
>[点击前往app helm仓库](https://github.com/ctxgo/helm-charts/tree/master/go-file-server)
```shell
# 添加app helm仓库
helm repo add go-file-server https://ctxgo.github.io/helm-charts/

# 如果修改了上述中间values配置,对于的配置文件也需要修改
# 创建configMap
kubectl create configmap go-file-server --from-file=config.yaml=config.yaml


#app vuales配置中支持ingress和ingressRoute,需要创建对对应额度 tls secrets
kubectl create secret tls example.com --key example.com.key --cert example.com.crt

#安装前后端app
#安装之前请修改app-1.0.0-values.yaml，如existingConfigMap、persistence、ingress部分
helm install go-file-server go-file-server/go-file-server --version 1.0.0 --values app-1.0.0-values.yaml
```

## 预览
<img width="1440" alt="go-file-server" src="https://github.com/user-attachments/assets/0b3865ea-7e92-426c-ab4f-e68427399df8">


<br>

## 🤝 特别感谢
[go-admin](https://github.com/go-admin-team/go-admin)
