import urllib.request
import json

try:
    data = json.loads(urllib.request.urlopen('http://localhost:8080/api/v1/jobs').read())
    for j in data['data']:
        print(f"Job {j.get('ID', '?')}: {j.get('status', '?')} - {j.get('metrics', '')}")
except Exception as e:
    print(e)
