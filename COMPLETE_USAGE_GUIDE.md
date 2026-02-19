# 📖 Complete Usage Guide - Finetune Studio

End-to-end guide for using Finetune Studio from UI.

## 🚀 Quick Start

### 1. Start All Services

```bash
# Clone repository
git clone <repo-url>
cd finetune-studio

# Start with Docker Compose
docker-compose up --build

# Wait for services to start...
# ✓ PostgreSQL: localhost:5432
# ✓ MinIO: localhost:9000 (console: 9001)
# ✓ Backend API: localhost:8080
# ✓ Frontend UI: localhost:3000
```

### 2. Open Frontend

Open your browser and navigate to:
```
http://localhost:3000
```

You should see the Finetune Studio dashboard.

## 📋 Complete Workflow

### Step 1: Upload Dataset

1. **Navigate to Upload Page**
   - Click "Upload Dataset" button on dashboard
   - Or go to: http://localhost:3000/dataset-upload.html

2. **Fill Dataset Information**
   - Enter dataset name (e.g., "Sentiment Analysis Dataset")
   - Add description (optional)

3. **Upload File**
   - Drag & drop your JSON/JSONL file
   - Or click the dropzone to browse
   - File must be < 500MB

4. **Review Validation**
   - Check total samples count
   - Review dataset preview (first 10 samples)
   - View label distribution chart

5. **Upload**
   - Click "Upload Dataset" button
   - Wait for progress bar to complete
   - You'll be redirected to dashboard

**Expected Result:** Dataset appears in datasets list

### Step 2: Create Training Job

1. **Navigate to New Training**
   - Click "Start Training" button on dashboard
   - Or go to: http://localhost:3000/training-new.html

2. **Step 1: Select Dataset**
   - Choose your dataset from dropdown
   - Review dataset information
   - Click "Next →"

3. **Step 2: Select Base Model**
   - Choose a model:
     - Llama 3.2 1B (fast, 4GB VRAM)
     - Llama 3.2 3B (balanced, 8GB VRAM)
     - Llama 3.1 8B (best quality, 16GB VRAM)
   - Click "Next →"

4. **Step 3: Configure Training**
   - Adjust LoRA Rank (8/16/32)
     - Higher = better quality, slower training
   - Set Epochs (1-5)
     - More = better learning, longer time
   - Choose Learning Rate
     - Low: Stable but slow
     - Medium: Recommended
     - High: Fast but risky
   - Review estimation (time & VRAM)
   - Click "🚀 Start Training"

**Expected Result:** Job created, redirected to training view

### Step 3: Monitor Training

1. **Training View Opens Automatically**
   - URL: http://localhost:3000/training-view.html?id=X

2. **Monitor Progress**
   - Watch progress bar update
   - Check current epoch
   - View metrics cards:
     - Training loss (should decrease)
     - Accuracy (should increase)
     - Time elapsed
     - ETA remaining

3. **View Loss Chart**
   - Real-time line chart
   - Shows loss over epochs
   - Updates automatically

4. **Watch Logs Stream**
   - Real-time logs via SSE
   - Color-coded by level:
     - Green: INFO
     - Yellow: WARN
     - Red: ERROR
   - Auto-scrolls to bottom
   - Toggle auto-scroll if needed

5. **Available Actions**
   - **Cancel**: Stop training (if running)
   - **Refresh**: Update all data
   - **Back**: Return to dashboard

**Expected Result:** Training completes successfully

### Step 4: Download Model

1. **Wait for Completion**
   - Status changes to "Completed"
   - Progress bar shows 100%
   - "Download Model" button appears

2. **Download**
   - Click "📥 Download Model" button
   - ZIP file downloads automatically
   - Contains:
     - lora_adapters/ (adapter files)
     - gguf/ (quantized model)
     - metrics.json
     - README.md

3. **Verify Download**
   ```bash
   unzip model-*.zip
   ls -lh
   ```

**Expected Result:** Model files downloaded successfully

### Step 5: Evaluate Model

1. **Start Evaluation**
   - Click "🎯 Evaluate Model" button
   - Evaluation job created
   - Redirected to evaluation view

2. **Wait for Evaluation**
   - Status shows "Pending" or "Running"
   - Page auto-refreshes every 10 seconds
   - Evaluation typically takes 2-5 minutes

3. **View Results**
   - Status changes to "Completed"
   - Results appear automatically

**Expected Result:** Evaluation completes with results

### Step 6: Review Evaluation Results

1. **Summary Stats**
   - Test samples count
   - Base model name
   - Fine-tuned model name
   - Accuracy improvement

2. **Performance Comparison Chart**
   - Bar chart comparing:
     - Accuracy
     - F1 Score
     - Precision
     - Recall
   - Base model (orange)
   - Fine-tuned model (green)

3. **Metrics Comparison**
   - Left card: Base model metrics
   - Right card: Fine-tuned metrics (highlighted)
   - Improvement deltas shown in green

4. **Example Predictions**
   - Table with test examples
   - Shows:
     - Input text
     - Expected label
     - Base model prediction (✅/❌)
     - Fine-tuned prediction (✅/❌)
     - Winner badge

5. **Download Report**
   - Click "📥 Download Report (JSON)"
   - Complete evaluation data saved

**Expected Result:** Clear comparison showing improvement

## 🎯 Common Tasks

### View All Jobs

1. Go to Dashboard (http://localhost:3000)
2. Scroll to "Recent Training Jobs"
3. Use status filter to filter by:
   - All Status
   - Running
   - Completed
   - Failed
   - Pending

### Cancel Running Job

1. Go to Dashboard
2. Find running job
3. Click "Cancel" button
4. Confirm cancellation
5. Status changes to "Cancelled"

### View Job Details

1. Go to Dashboard
2. Find job in table
3. Click "View" button
4. Training view opens with full details

### Re-run Training

1. Go to "New Training"
2. Select same dataset
3. Adjust configuration if needed
4. Start new training job

### Compare Multiple Models

1. Complete multiple training jobs
2. Evaluate each model
3. Download evaluation reports
4. Compare JSON files manually

## 🔍 Troubleshooting

### Issue: Dashboard shows no data

**Solution:**
1. Check backend is running: http://localhost:8080/api/v1/health
2. Check browser console for errors (F12)
3. Verify CORS is enabled in backend
4. Refresh page (Ctrl+R)

### Issue: Dataset upload fails

**Possible causes:**
- File too large (> 500MB)
- Invalid JSON format
- Backend not running
- MinIO not accessible

**Solution:**
1. Check file size: `ls -lh your-file.json`
2. Validate JSON: `jq . your-file.json`
3. Check backend logs
4. Verify MinIO is running: http://localhost:9001

### Issue: Training job stuck in "Pending"

**Possible causes:**
- Worker pool full
- Kaggle not configured
- Dataset not found

**Solution:**
1. Check backend logs
2. Verify worker pool is running
3. Check job queue size
4. Cancel and retry

### Issue: Logs not streaming

**Possible causes:**
- SSE connection failed
- Job not running
- Backend not responding

**Solution:**
1. Check browser console for SSE errors
2. Verify job is running
3. Check backend SSE endpoint: `curl -N http://localhost:8080/api/v1/jobs/1/logs`
4. Refresh page

### Issue: Model download fails

**Possible causes:**
- Model not ready
- MinIO not accessible
- Presigned URL expired

**Solution:**
1. Verify job status is "Completed"
2. Check MinIO is running
3. Try again (new presigned URL generated)
4. Check backend logs

### Issue: Evaluation stuck

**Possible causes:**
- Ollama not running
- Model not loaded
- Test set not found

**Solution:**
1. Check Ollama: `curl http://localhost:11434/api/tags`
2. Verify models loaded: `docker exec ollama ollama list`
3. Check test set path
4. Review backend logs

## 📊 Understanding Metrics

### Training Metrics

**Loss:**
- Measures prediction error
- Lower is better
- Should decrease over epochs
- Typical range: 0.1 - 2.0

**Accuracy:**
- Percentage of correct predictions
- Higher is better
- Range: 0.0 - 1.0 (0% - 100%)
- Target: > 0.8 (80%)

**Epochs:**
- One complete pass through dataset
- More epochs = more learning
- Too many = overfitting
- Typical: 3-5 epochs

### Evaluation Metrics

**Accuracy:**
- Overall correctness
- (Correct predictions) / (Total predictions)

**Precision:**
- Of predicted positives, how many are correct
- High precision = few false positives

**Recall:**
- Of actual positives, how many were found
- High recall = few false negatives

**F1 Score:**
- Harmonic mean of precision and recall
- Balanced metric
- Range: 0.0 - 1.0

**Response Time:**
- Average time per prediction
- Lower is better
- Measured in milliseconds

## 🎨 UI Tips

### Keyboard Shortcuts

- `Ctrl+R` / `F5`: Refresh page
- `Ctrl+Click`: Open link in new tab
- `Esc`: Close modals (if any)

### Navigation

- Use browser back button to go back
- Or use "← Back to Dashboard" buttons
- Breadcrumb navigation in header

### Auto-refresh

- Dashboard: Every 10 seconds
- Training view: Every 5 seconds
- Evaluation view: Every 10 seconds (if pending)

### Mobile Usage

- All pages are mobile-responsive
- Tables scroll horizontally
- Touch-friendly buttons
- Optimized for small screens

## 🔐 Best Practices

### Dataset Preparation

1. **Clean your data:**
   - Remove duplicates
   - Fix formatting issues
   - Validate JSON structure

2. **Balance your dataset:**
   - Equal samples per class
   - Minimum 100 samples per class
   - Recommended: 1000+ total samples

3. **Split your data:**
   - Training: 80%
   - Validation: 10%
   - Test: 10%

### Training Configuration

1. **Start small:**
   - Use smaller model first (1B)
   - Fewer epochs (2-3)
   - Lower LoRA rank (8-16)

2. **Monitor closely:**
   - Watch loss curve
   - Check for overfitting
   - Review logs for errors

3. **Iterate:**
   - Adjust based on results
   - Increase epochs if underfitting
   - Decrease if overfitting

### Model Evaluation

1. **Use separate test set:**
   - Never use training data
   - Representative of real data
   - Sufficient size (100+ samples)

2. **Compare fairly:**
   - Same test set for all models
   - Same evaluation metrics
   - Document differences

3. **Analyze examples:**
   - Review incorrect predictions
   - Identify patterns
   - Improve dataset if needed

## 📈 Performance Expectations

### Training Time

- **1B model, 3 epochs:** ~30 minutes
- **3B model, 3 epochs:** ~1-2 hours
- **8B model, 3 epochs:** ~3-4 hours

*Times vary based on:*
- Dataset size
- Hardware (GPU/CPU)
- LoRA rank
- Batch size

### Accuracy Improvement

- **Good:** +10-20% over base model
- **Great:** +20-30% over base model
- **Excellent:** +30%+ over base model

*Depends on:*
- Dataset quality
- Task difficulty
- Base model capability
- Training configuration

### Resource Usage

- **VRAM:** 4-16GB (depends on model)
- **RAM:** 8-16GB recommended
- **Disk:** 10-50GB per model
- **Network:** Minimal (except upload/download)

## 🎓 Learning Resources

### Understanding Fine-tuning

- [LoRA Paper](https://arxiv.org/abs/2106.09685)
- [Llama 3 Documentation](https://ai.meta.com/llama/)
- [Fine-tuning Best Practices](https://huggingface.co/docs/transformers/training)

### Dataset Preparation

- [Data Cleaning Guide](https://www.kaggle.com/learn/data-cleaning)
- [JSON Format Specification](https://www.json.org/)
- [JSONL Format](http://jsonlines.org/)

### Model Evaluation

- [Evaluation Metrics Explained](https://scikit-learn.org/stable/modules/model_evaluation.html)
- [Precision vs Recall](https://en.wikipedia.org/wiki/Precision_and_recall)
- [F1 Score](https://en.wikipedia.org/wiki/F-score)

## 🆘 Getting Help

### Check Logs

**Backend logs:**
```bash
docker-compose logs backend
```

**Frontend logs:**
- Open browser console (F12)
- Check Network tab for API calls
- Look for errors in Console tab

### Report Issues

Include:
1. Steps to reproduce
2. Expected behavior
3. Actual behavior
4. Screenshots
5. Browser console errors
6. Backend logs

### Community Support

- GitHub Issues
- Discord Server
- Stack Overflow (tag: finetune-studio)

## 🎉 Success!

You've completed the full workflow:
1. ✅ Uploaded dataset
2. ✅ Created training job
3. ✅ Monitored training progress
4. ✅ Downloaded trained model
5. ✅ Evaluated model performance
6. ✅ Reviewed results

Your fine-tuned model is ready to use!

## 🚀 Next Steps

1. **Deploy your model:**
   - Use with Ollama
   - Integrate into your app
   - Serve via API

2. **Improve further:**
   - Try different configurations
   - Use more training data
   - Experiment with base models

3. **Share your results:**
   - Document your findings
   - Share with community
   - Contribute improvements

Happy fine-tuning! 🎯
