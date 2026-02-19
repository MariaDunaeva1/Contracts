# 🔧 Troubleshooting Frontend Connection

## Problem: "Failed to load dashboard data"

This error occurs when the frontend cannot connect to the backend API.

## Quick Fix

### Option 1: Use Python Server (Recommended)

1. **Start Backend** (if not running):
```bash
cd backend
go run ./cmd/server
```

2. **Start Frontend**:
```bash
# Windows
start-frontend.bat

# Or manually:
cd frontend
python serve.py
```

3. **Open Browser**:
```
http://localhost:3000
```

### Option 2: Test Connection First

1. **Open test page**:
```
http://localhost:3000/test-connection.html
```

2. **Click "Test Health"** button
   - If successful: Backend is working
   - If failed: Backend is not running or not accessible

3. **Click other test buttons** to verify all endpoints

## Common Issues

### Issue 1: Backend not running

**Symptoms:**
- "Failed to connect to backend"
- Network errors in browser console
- Test page shows all tests failed

**Solution:**
```bash
# Check if backend is running
curl http://localhost:8080/api/v1/health

# If not running, start it:
cd backend
go run ./cmd/server
```

### Issue 2: CORS errors

**Symptoms:**
- "CORS policy" errors in browser console
- Backend is running but requests fail

**Solution:**
Backend already has CORS enabled. If still having issues:

1. Check backend logs for CORS errors
2. Verify CORS middleware in `backend/cmd/server/main.go`
3. Try clearing browser cache (Ctrl+Shift+Delete)

### Issue 3: Wrong API URL

**Symptoms:**
- 404 errors
- "Cannot GET /api/v1/..." errors

**Solution:**
Check `frontend/js/api.js` line 7:
```javascript
const API_BASE_URL = 'http://localhost:8080/api/v1';
```

Make sure it matches your backend URL.

### Issue 4: Port already in use

**Symptoms:**
- "Address already in use" error
- Cannot start frontend server

**Solution:**
```bash
# Windows - Find process using port 3000
netstat -ano | findstr :3000

# Kill process
taskkill /PID <PID> /F

# Or use different port
cd frontend
python -m http.server 3001
```

## Detailed Diagnostics

### Step 1: Check Backend

```bash
# Test health endpoint
curl http://localhost:8080/api/v1/health

# Expected response:
# {"status":"ok","services":{"db":"up","storage":"up"}}
```

### Step 2: Check Frontend Files

```bash
# Verify files exist
ls frontend/index.html
ls frontend/js/api.js
ls frontend/css/styles.css
```

### Step 3: Check Browser Console

1. Open browser (Chrome/Edge/Firefox)
2. Press F12 to open Developer Tools
3. Go to Console tab
4. Look for errors (red text)
5. Common errors:
   - `Failed to fetch`: Backend not accessible
   - `CORS error`: CORS not configured
   - `404 Not Found`: Wrong URL
   - `Network error`: Backend not running

### Step 4: Check Network Tab

1. Open Developer Tools (F12)
2. Go to Network tab
3. Refresh page (F5)
4. Look for failed requests (red)
5. Click on failed request to see details

## Manual Testing

### Test 1: Backend Health

```bash
curl http://localhost:8080/api/v1/health
```

Expected: `{"status":"ok",...}`

### Test 2: List Jobs

```bash
curl http://localhost:8080/api/v1/jobs
```

Expected: `{"data":[...],"total":...}`

### Test 3: List Datasets

```bash
curl http://localhost:8080/api/v1/datasets
```

Expected: `{"data":[...],"total":...}`

### Test 4: Frontend Access

```bash
curl http://localhost:3000
```

Expected: HTML content with "Finetune Studio"

## Alternative: Direct File Access

If you can't run a server, you can open files directly:

1. **Open in browser**:
```
file:///C:/path/to/frontend/index.html
```

2. **Note**: Some features won't work:
   - API calls will fail (CORS)
   - Relative paths may break
   - SSE won't work

**Not recommended for development.**

## Docker Setup (Alternative)

If you prefer Docker:

```bash
# Build and start all services
docker-compose up --build

# Frontend: http://localhost:3000
# Backend: http://localhost:8080
```

## Verify Setup

### Checklist:

- [ ] Backend running on port 8080
- [ ] Frontend running on port 3000
- [ ] Can access http://localhost:8080/api/v1/health
- [ ] Can access http://localhost:3000
- [ ] No CORS errors in browser console
- [ ] Dashboard loads without errors

## Still Not Working?

### Get Detailed Logs

1. **Backend logs**:
```bash
# If running directly
# Check terminal where you ran: go run ./cmd/server

# If using Docker
docker-compose logs backend
```

2. **Browser console logs**:
   - Open DevTools (F12)
   - Console tab
   - Copy all errors
   - Look for red text

3. **Network logs**:
   - DevTools (F12)
   - Network tab
   - Filter: XHR
   - Check failed requests

### Report Issue

Include:
1. Backend logs
2. Browser console errors
3. Network tab screenshot
4. Steps to reproduce
5. Operating system
6. Browser version

## Quick Commands Reference

```bash
# Start backend
cd backend
go run ./cmd/server

# Start frontend (Python)
cd frontend
python serve.py

# Start frontend (Node.js)
cd frontend
npx serve -p 3000

# Test backend
curl http://localhost:8080/api/v1/health

# Test frontend
curl http://localhost:3000

# Check ports
netstat -ano | findstr :8080
netstat -ano | findstr :3000
```

## Success Indicators

When everything works:
- ✅ Backend responds to health check
- ✅ Frontend loads without errors
- ✅ Dashboard shows stats (even if 0)
- ✅ No red errors in console
- ✅ Network tab shows successful requests (200 OK)

## Next Steps

Once connection is working:
1. Upload a dataset
2. Create a training job
3. Monitor progress
4. Download model
5. Run evaluation

Happy debugging! 🐛
