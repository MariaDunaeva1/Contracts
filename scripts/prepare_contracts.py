import json

# Leer dataset (formato JSON Lines)
data = []
with open('data/contracts/train.json', 'r', encoding='utf-8') as f:
    for line in f:
        data.append(json.loads(line))

print(f"📊 Loaded {len(data)} contracts")

# Convertir a formato fine-tuning
finetune_data = []

for item in data[:500]:  # Primeros 500
    finetune_data.append({
        "messages": [
            {
                "role": "user", 
                "content": f"Classify this contract clause:\n\n{item['text'][:500]}"
            },
            {
                "role": "assistant",
                "content": item.get('provision', 'Unknown')
            }
        ]
    })

# Guardar
with open('data/contracts/finetune_train.json', 'w', encoding='utf-8') as f:
    json.dump(finetune_data, f, indent=2, ensure_ascii=False)

print(f"✅ Created {len(finetune_data)} training examples")
print(f"📁 Saved to: data/contracts/finetune_train.json")