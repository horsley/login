# 一个简单的统一登录

结合 nginx 的 ngx_http_auth_request_module 模块做的统一登录

对于要保护的后端，进行如下配置

```
# 需要保护的路径添加下面这两行
auth_request /auth;
error_page 401 = /login;
```


然后把登录和鉴权的location 配置加入
```
# 需要保护的站点添加如下 location 配置 分别用于nginx鉴权和登录 
# 可以通过include方式引入避免复制粘贴 include auth.inc
# 
# auth.inc
location = /auth {
    internal;
    proxy_pass http://127.0.0.1:23514/auth;
    proxy_pass_request_body off;
    proxy_set_header Content-Length "";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Scheme $scheme;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Original-URI $request_uri;
}

location /login {
    proxy_pass http://127.0.0.1:23514;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Scheme $scheme;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Original-URI $request_uri;
}
```

使用 docker 可以拉镜像 https://hub.docker.com/r/horsley/login/tags

通过 volume 可以挂配置文件进去，配置文件参考 config.example.yaml，挂到容器内 /config.yaml 即可
