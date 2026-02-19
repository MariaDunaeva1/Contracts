# Finetune Studio - Frontend

Modern, responsive web interface for Finetune Studio.

## Features

### ✅ Dashboard (index.html)
- Quick stats overview (models trained, datasets uploaded, training time, success rate)
- Recent jobs list with real-time status
- Quick action buttons
- Auto-refresh every 10 seconds

### ✅ Dataset Upload (dataset-upload.html)
- Drag & drop file upload
- Real-time validation
- Dataset preview (first 10 samples)
- Label distribution chart
- Progress bar during upload
- Supports JSON and JSONL formats (max 500MB)

### ✅ New Training (training-new.html)
- 3-step wizard:
  1. Select dataset from dropdown
  2. Choose base model (Llama 3.2 1B/3B, Llama 3.1 8B)
  3. Configure training parameters
- Interactive sliders for LoRA rank and epochs
- Learning rate selection
- Real-time training estimation (time & VRAM)

### ✅ Training View (training-view.html)
- Real-time progress bar
- Current epoch indicator
- Live metrics cards (loss, accuracy, time, ETA)
- Training loss chart (Chart.js)
- Real-time logs streaming (SSE)
- Auto-scroll logs with toggle
- Action buttons:
  - Cancel (if running)
  - Download model (if completed)
  - Evaluate model (if completed)

### ✅ Evaluation View (evaluation-view.html)
- Performance comparison chart
- Side-by-side metrics comparison
- Improvement deltas
- Examples table with predictions
- Winner badges
- Download report as JSON

## Tech Stack

- **Pure JavaScript** (no frameworks)
- **Chart.js** for visualizations
- **Server-Sent Events (SSE)** for real-time logs
- **Fetch API** for HTTP requests
- **Custom CSS** (Tailwind-inspired utilities)
- **Nginx** for serving static files and proxying API

## File Structure

```
frontend/
├── index.html                  # Dashboard
├── dataset-upload.html         # Upload dataset
├── training-new.html           # Create new training job
├── training-view.html          # View training progress
├── evaluation-view.html        # View evaluation results
├── css/
│   └── styles.css             # Main stylesheet
├── js/
│   ├── api.js                 # API client
│   ├── utils.js               # Utility functions
│   └── charts.js              # Chart.js wrappers
├── nginx.conf                 # Nginx configuration
├── Dockerfile                 # Docker image
└── README.md                  # This file
```

## Development

### Local Development (without Docker)

1. **Start backend server:**
```bash
cd backend
go run ./cmd/server
```

2. **Serve frontend with any static server:**
```bash
cd frontend

# Option 1: Python
python -m http.server 3000

# Option 2: Node.js
npx serve -p 3000

# Option 3: PHP
php -S localhost:3000
```

3. **Update API URL in js/api.js:**
```javascript
const API_BASE_URL = 'http://localhost:8080/api/v1';
```

4. **Open browser:**
```
http://localhost:3000
```

### Docker Development

```bash
# Build and start all services
docker-compose up --build

# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
# MinIO Console: http://localhost:9001
```

## API Integration

The frontend communicates with the backend through the `api.js` module:

```javascript
// Example: Upload dataset
const result = await api.uploadDataset(file, name, description);

// Example: Create training job
const job = await api.createJob(datasetId, configuration);

// Example: Stream logs (SSE)
const eventSource = api.streamLogs(jobId, 
    (data) => console.log('New logs:', data),
    (error) => console.error('Error:', error)
);

// Example: Download model
api.downloadModel(modelId);
```

## Responsive Design

The UI is fully responsive and works on:
- Desktop (1200px+)
- Tablet (768px - 1199px)
- Mobile (< 768px)

Key responsive features:
- Grid layouts collapse to single column on mobile
- Navigation menu adapts to smaller screens
- Tables scroll horizontally on mobile
- Touch-friendly buttons and controls

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+
- Opera 76+

Required features:
- ES6+ JavaScript
- Fetch API
- EventSource (SSE)
- CSS Grid
- CSS Flexbox

## Customization

### Colors

Edit `css/styles.css`:

```css
:root {
    --primary: #0066cc;
    --success: #28a745;
    --danger: #dc3545;
    --warning: #ffc107;
    /* ... */
}
```

### API Endpoint

Edit `js/api.js`:

```javascript
const API_BASE_URL = 'http://your-api-url/api/v1';
```

### Auto-refresh Intervals

Edit respective HTML files:

```javascript
// Dashboard: 10 seconds
setInterval(loadDashboard, 10000);

// Training view: 5 seconds
setInterval(loadJob, 5000);

// Evaluation view: 10 seconds (only if pending/running)
setInterval(() => {
    if (evaluation.status === 'pending' || evaluation.status === 'running') {
        loadEvaluation();
    }
}, 10000);
```

## Performance

- Gzip compression enabled
- Static assets cached for 1 year
- API responses not cached
- SSE connections kept alive
- Charts update efficiently (no full redraw)

## Security

- XSS protection via `escapeHtml()` utility
- CORS headers configured in backend
- No sensitive data in frontend
- Presigned URLs for secure downloads
- Security headers in Nginx config

## Testing

### Manual Testing Checklist

- [ ] Dashboard loads and displays stats
- [ ] Can upload dataset with drag & drop
- [ ] Dataset validation works
- [ ] Can create new training job
- [ ] Training progress updates in real-time
- [ ] Logs stream without lag
- [ ] Can cancel running job
- [ ] Can download completed model
- [ ] Can start evaluation
- [ ] Evaluation results display correctly
- [ ] All charts render properly
- [ ] Mobile responsive works
- [ ] Error handling shows user-friendly messages

### Automated Testing

```bash
# Run backend tests
cd backend
go test ./...

# Test API endpoints
bash scripts/test_logs_and_models.sh
```

## Troubleshooting

### Issue: API requests fail with CORS error

**Solution:** Ensure backend has CORS middleware enabled:
```go
r.Use(func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
    c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
    if c.Request.Method == "OPTIONS" {
        c.AbortWithStatus(204)
        return
    }
    c.Next()
})
```

### Issue: SSE logs not streaming

**Solution:** 
1. Check backend is running
2. Verify job ID is correct
3. Check browser console for errors
4. Ensure SSE endpoint returns correct headers:
   - `Content-Type: text/event-stream`
   - `Cache-Control: no-cache`
   - `Connection: keep-alive`

### Issue: Charts not rendering

**Solution:**
1. Ensure Chart.js is loaded: `<script src="https://cdn.jsdelivr.net/npm/chart.js"></script>`
2. Check canvas element exists
3. Verify data format is correct
4. Check browser console for errors

### Issue: File upload fails

**Solution:**
1. Check file size < 500MB
2. Verify file format is JSON or JSONL
3. Ensure backend accepts multipart/form-data
4. Check backend logs for errors

## Production Deployment

### Build Docker Image

```bash
cd frontend
docker build -t finetune-studio-frontend .
```

### Run Container

```bash
docker run -d \
  -p 3000:80 \
  --name frontend \
  -e API_URL=http://your-backend-url \
  finetune-studio-frontend
```

### Nginx Configuration

For production, update `nginx.conf`:

```nginx
# Enable SSL
listen 443 ssl http2;
ssl_certificate /path/to/cert.pem;
ssl_certificate_key /path/to/key.pem;

# Update backend proxy
location /api/ {
    proxy_pass https://your-backend-url;
    # ... other settings
}
```

## Contributing

1. Follow existing code style
2. Test on multiple browsers
3. Ensure mobile responsive
4. Add comments for complex logic
5. Update this README if adding features

## License

Same as main project.
