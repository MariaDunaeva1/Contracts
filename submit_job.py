import urllib.request
import json
import time

job_payload = {
    "dataset_id": 1,
    "configuration": {"epochs": 3, "batch_size": 4, "learning_rate": "2e-5"}
}

req = urllib.request.Request(
    'http://localhost:8080/api/v1/jobs', 
    data=json.dumps(job_payload).encode('utf-8'),
    headers={'Content-Type': 'application/json'},
    method='POST'
)

print("Submitting training job...")
try:
    with urllib.request.urlopen(req) as response:
        res = json.loads(response.read())
        print(f"Job created: {res}")
        job_id = res['ID']
        
        print(f"Polling job {job_id} status...")
        for _ in range(60): # Poll for up to 5 minutes
            time.sleep(5)
            status_req = urllib.request.Request(f'http://localhost:8080/api/v1/jobs')
            with urllib.request.urlopen(status_req) as s_res:
                jobs = json.loads(s_res.read())['data']
                job = next(j for j in jobs if j['ID'] == job_id)
                print(f"Status: {job['status']} | Metrics: {job.get('metrics', '')}")
                if job['status'] in ['completed', 'failed']:
                    break
except urllib.error.HTTPError as e:
    print(f"Error: {e.code} {e.reason}")
    print(e.read().decode('utf-8'))
except Exception as e:
    print(f"Error: {e}")
