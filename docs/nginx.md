# Nginx TLS 与反向代理

UniSub 建议只监听本地地址，例如 `127.0.0.1:8080`，由 Nginx 提供 TLS 终止和反向代理。

## Nginx 配置

```nginx
server {
    listen 443 ssl http2;
    server_name sub.example.com;

    ssl_certificate     /etc/letsencrypt/live/sub.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/sub.example.com/privkey.pem;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:10m;
    ssl_protocols TLSv1.2 TLSv1.3;

    access_log /var/log/nginx/unisub.access.log;
    error_log  /var/log/nginx/unisub.error.log;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 60s;
        proxy_connect_timeout 10s;
    }
}

server {
    listen 80;
    server_name sub.example.com;
    return 301 https://$host$request_uri;
}
```

## 启用步骤

```bash
sudo nginx -t
sudo systemctl reload nginx
```

如果证书还未签发，可先用 `certbot --nginx -d sub.example.com` 生成证书后再重载。
