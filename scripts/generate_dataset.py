import json
import random

def generate_dataset(output_file="sentiment_500.json", num_examples=500):
    sentiments = ["positive", "negative", "neutral"]
    data = []

    positive_texts = [
        "I love this product!", "Amazing experience.", "Great service.", "Highly recommended.",
        "Best purchase ever.", "Very satisfied.", "Exceeded expectations.", "Fantastic quality.",
        "Will buy again.", "Five stars!"
    ]
    negative_texts = [
        "Terrible service.", "Waste of money.", "Very disappointed.", "Not worth it.",
        "Poor quality.", "Arrived damaged.", "Customer support was rude.", "Never again.",
        "Faulty item.", "One star."
    ]
    neutral_texts = [
        "It was okay.", "Average performance.", "Nothing special.", "Just as described.",
        "Met requirements.", "Standard delivery.", "Decent for the price.", "Mixed feelings.",
        "Not bad, not great.", "It works."
    ]

    for i in range(num_examples):
        sentiment = random.choice(sentiments)
        if sentiment == "positive":
            text = random.choice(positive_texts) + f" (Sample {i})"
        elif sentiment == "negative":
            text = random.choice(negative_texts) + f" (Sample {i})"
        else:
            text = random.choice(neutral_texts) + f" (Sample {i})"
        
        data.append({
            "text": text,
            "label": sentiment
        })

    with open(output_file, "w") as f:
        json.dump(data, f, indent=2)
    
    print(f"Generated {num_examples} examples in {output_file}")

if __name__ == "__main__":
    generate_dataset()
