# tiny-proxy

Hacking around a simple http proxy.

Don't use in production!

# Example Config

```yaml
servers:
  - http:
      host: localhost
      port: 8000
      middlewares:
        - name: log
        - name: cors
          options:
            allowOrigins: ["*"]
            allowMethods: ["*"]
            allowHeaders: ["*"]
      routes:
        - path: "/api/(.+)"
          rewrite: "/api/{1}"
          backend:
            url: http://localhost:4242
        - path: "/frontend/*"
          backend:
            url: http://localhost:5173
```
