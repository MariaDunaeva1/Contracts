import requests
import json
import time
import threading
from generate_dataset import generate_dataset

BASE_URL = "http://localhost:8080/api/v1"

def test_full_flow():
    print("--- Starting System Verification ---")
    
    # 1. Generate Dataset
    print("[1] Generating dataset...")
    generate_dataset("test_data.json", 100)

    # 2. Upload Dataset
    print("[2] Uploading dataset...")
    files = {'file': ('test_data.json', open('test_data.json', 'rb'), 'application/json')}
    data = {'name': 'Verification Dataset', 'description': 'Automated test dataset'}
    
    resp = requests.post(f"{BASE_URL}/datasets", files=files, data=data)
    if resp.status_code != 201:
        print(f"FAILED to upload dataset: {resp.text}")
        return
    
    dataset_id = resp.json()['ID']
    print(f"✅ Dataset Uploaded! ID: {dataset_id}")

    # 3. Create 5 Concurrent Jobs
    print("[3] Creating 5 concurrent jobs...")
    job_ids = []
    
    for i in range(5):
        payload = {
            "dataset_id": dataset_id,
            "configuration": {
                "epochs": 3,
                "learning_rate": 0.001
            }
        }
        resp = requests.post(f"{BASE_URL}/jobs", json=payload)
        if resp.status_code == 201:
            jid = resp.json()['ID']
            job_ids.append(jid)
            print(f"   -> Job created: {jid}")
        else:
            print(f"FAILED to create job: {resp.text}")

    # 4. Monitor Jobs
    print("[4] Monitoring jobs (concurrency check)...")
    
    active_jobs = set(job_ids)
    
    while active_jobs:
        print(f"   Active jobs: {active_jobs}")
        
        for jid in list(active_jobs):
            resp = requests.get(f"{BASE_URL}/jobs/{jid}")
            data = resp.json()
            status = data.get('status')
            
            if status in ["completed", "failed", "cancelled"]:
                print(f"   ✅ Job {jid} finished with status: {status}")
                if "metrics" in data:
                     print(f"      Metrics: {data['metrics']}")
                active_jobs.remove(jid)
        
        time.sleep(2)

    print("--- Verification Completed Successfully ---")

if __name__ == "__main__":
    # Ensure requests library is installed or use standard lib? 
    # Assuming user has requests or I can install it. 
    # For safety, I'll use standard library if requests is not available, but requests is standard enough.
    # To be safe for the user environment, I should probably check/install requests?
    # Or just use the user's environment.
    try:
        test_full_flow()
    except ImportError:
        print("Please install requests: pip install requests")
    except Exception as e:
        print(f"Verification failed: {e}")
