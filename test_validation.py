import requests
import json
import sys

# Test 1: Upload a generic invalid JSON to see error
url = "http://localhost:8080/api/v1/datasets"
with open("bad_dataset.json", "w") as f:
    f.write('[{"wrong_key": "value"}]')

files = {'file': open('bad_dataset.json', 'rb')}
response = requests.post(url, files=files)
print(f"Bad Dataset Response: {response.text}")

# Let's also check what the most recent log says
