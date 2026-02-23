import requests

# Test full dataset validation
print("Testing full ledgar_finetune_train.json")
url = "http://localhost:8080/api/v1/datasets"
files = {'file': open('data/contracts/ledgar_finetune_train.json', 'rb')}
response = requests.post(url, files=files)
print(f"Validation Response: {response.text}")
