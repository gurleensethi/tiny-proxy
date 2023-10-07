# tiny-proxy

Hacking around a simple http proxy.

Don't use in production!

# Example Config

```yaml
servers:
  - http:
      host: localhost
      port: 8000
      log: true
      routes:
        - path: "/api/.+"
          backend:
            url: http://localhost:4242
        - path: "/frontend/*"
          backend:
            url: http://localhost:5173
```
