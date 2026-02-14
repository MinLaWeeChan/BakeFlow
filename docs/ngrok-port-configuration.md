# ngrok Port Configuration Guide

## Quick Reference

| Branch | ngrok Port | Target Service | Architecture |
|--------|-----------|----------------|--------------|
| `main` | **8080** | Go Backend | Backend-first (recommended) |
| `feature/multiple-orders` | **3000** | Next.js Frontend | Frontend-first |

## Current Setup (main branch)

### Port Configuration
- **Backend (Go)**: Port 8080
- **Frontend (Next.js)**: Port 3000
- **ngrok**: Port 8080 (exposing backend)

### Command
```bash
ngrok http 8080
```

### Architecture Flow
```
External Services (Facebook, Payments)
    ↓
ngrok (public URL)
    ↓
Backend :8080 (Go API)
    ↑
Frontend :3000 (Next.js) ← User Browser
```

### Why Port 8080?
The backend needs to be publicly accessible for:
- **Facebook Messenger webhooks** - Receives messages, postbacks, delivery confirmations
- **Payment callbacks** - Payment provider notifications
- **External API integrations** - Any third-party service callbacks

### Next.js Proxy Configuration
The frontend proxies backend requests via `next.config.mjs`:

```javascript
async rewrites() {
  return [
    { source: '/api/:path*', destination: 'http://localhost:8080/api/:path*' },
    { source: '/uploads/:path*', destination: 'http://localhost:8080/uploads/:path*' },
    { source: '/webhook', destination: 'http://localhost:8080/webhook' },
    { source: '/qr_codes/:path*', destination: 'http://localhost:8080/qr_codes/:path*' },
    { source: '/promotions/:path*', destination: 'http://localhost:8080/promotions/:path*' },
    { source: '/checkout', destination: 'http://localhost:8080/checkout' },
  ];
}
```

## Previous Setup (feature/multiple-orders branch)

### Port Configuration
- **Backend (Go)**: Port 8080
- **Frontend (Next.js)**: Port 3000
- **ngrok**: Port 3000 (exposing frontend)

### Command
```bash
ngrok http 3000
```

### Why This Changed
The `feature/multiple-orders` branch didn't have the API proxy rewrites in `next.config.mjs`, so the frontend was exposed directly. This approach has limitations:
- ❌ Webhooks go through Next.js proxy (extra overhead)
- ❌ Less clear separation of concerns
- ❌ Frontend becomes a bottleneck for backend operations

## Backend Port Configuration

The backend port is set in [`backend/main.go`](file:///home/keys/Desktop/Bakeflow/BakeFlow/backend/main.go):

```go
// Get port from environment or use default
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
```

To change the backend port:
```bash
PORT=8081 go run main.go
```

## Frontend Port Configuration

Next.js defaults to port 3000. To change it:

```bash
# In package.json
"dev": "next dev -p 3001"

# Or run directly
npm run dev -- -p 3001
```

## Troubleshooting

### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080

# Or
netstat -tulpn | grep :8080

# Kill the process
kill -9 <PID>
```

### ngrok Not Working
1. Check if backend is running: `curl http://localhost:8080/api/health`
2. Verify ngrok is pointing to correct port: `ps aux | grep ngrok`
3. Check ngrok config: `ngrok config check`
4. Restart ngrok: `ngrok http 8080`

### Webhook Configuration
When setting up Facebook Messenger webhooks, use:
```
https://<your-ngrok-url>/webhook
```

The ngrok URL changes each time you restart (unless using a paid plan with reserved domains).

## Best Practices

✅ **DO**:
- Use port 8080 for backend (main branch approach)
- Expose backend via ngrok for webhooks
- Use Next.js rewrites to proxy API calls
- Document your ngrok URL when testing

❌ **DON'T**:
- Expose frontend via ngrok (unless specifically needed)
- Hardcode ngrok URLs in code
- Forget to update webhook URLs after ngrok restart
- Run multiple ngrok instances on same port

## Related Files
- [`backend/main.go`](file:///home/keys/Desktop/Bakeflow/BakeFlow/backend/main.go) - Backend port configuration
- [`frontend/next.config.mjs`](file:///home/keys/Desktop/Bakeflow/BakeFlow/frontend/next.config.mjs) - API proxy rewrites
- [`frontend/package.json`](file:///home/keys/Desktop/Bakeflow/BakeFlow/frontend/package.json) - Frontend dev scripts
- [`backend/controllers/webhook.go`](file:///home/keys/Desktop/Bakeflow/BakeFlow/backend/controllers/webhook.go) - Webhook handler (expects port 8080)
