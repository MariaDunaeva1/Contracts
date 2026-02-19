"""
Script to map numeric CUAD labels to meaningful category names
and regenerate finetune_train.json with proper labels.
"""
import json
import os

# CUAD (Contract Understanding Atticus Dataset) label mapping
# These are the standard 41 CUAD categories + additional contract clause types
LABEL_MAP = {
    0: "Document Name",
    1: "Parties",
    2: "Agreement Date",
    3: "Effective Date",
    4: "Expiration Date",
    5: "Renewal Term",
    6: "Notice Period To Terminate Renewal",
    7: "Governing Law",
    8: "Most Favored Nation",
    9: "Non-Compete",
    10: "Exclusivity",
    11: "No-Solicitation Of Customers",
    12: "Competitive Restriction Exception",
    13: "Non-Solicitation Of Employees",
    14: "Non-Disparagement",
    15: "Termination For Convenience",
    16: "Rofr/Rofo/Rofn",
    17: "Change Of Control",
    18: "Anti-Assignment",
    19: "Revenue/Profit Sharing",
    20: "Price Restrictions",
    21: "Minimum Commitment",
    22: "Volume Restriction",
    23: "Ip Ownership Assignment",
    24: "Joint Ip Ownership",
    25: "License Grant",
    26: "Non-Transferable License",
    27: "Affiliate License-Loss Of Control",
    28: "Unlimited/All-You-Can-Eat License",
    29: "Irrevocable Or Perpetual License",
    30: "Source Code Escrow",
    31: "Post-Termination Services",
    32: "Audit Rights",
    33: "Uncapped Liability",
    34: "Cap On Liability",
    35: "Liquidated Damages",
    36: "Warranty Duration",
    37: "Insurance",
    38: "Covenant Not To Sue",
    39: "Third Party Beneficiary",
}

# Extended labels based on analysis of the actual data
# Map remaining labels by examining the data patterns
EXTENDED_LABEL_MAP = {
    2: "Amendment",
    4: "Governing Law",
    6: "Arbitration",
    7: "Anti-Assignment",
    11: "Base Salary",
    12: "Third Party Beneficiary",
    13: "Binding Agreement",
    15: "Broker Representation",
    16: "Capitalization",
    17: "Change Of Control",
    18: "Closing Conditions",
    19: "Compliance With Laws",
    20: "Confidentiality",
    22: "Government Authorization",
    23: "Construction Of Agreement",
    24: "Cooperation",
    25: "Costs And Attorneys Fees",
    26: "Counterparts",
    28: "Definitions",
    31: "Disclosure",
    32: "Duties And Responsibilities",
    33: "Assignment Of Rights",
    35: "Employment Terms",
    38: "Entire Agreement",
    39: "ERISA Compliance",
    40: "Existence And Good Standing",
    41: "Expenses And Reimbursement",
    42: "Interest Rate Computation",
    43: "Financial Reporting",
    44: "Forfeiture",
    45: "Further Assurances",
    46: "Issuance Of Letters Of Credit",
    47: "Governing Law",
    49: "Indemnification",
    50: "Indemnification",
    51: "Insurance",
    53: "Intellectual Property",
    54: "Interest Rate",
    55: "Interpretation",
    56: "Governing Law",
    58: "Litigation",
    59: "Miscellaneous Provisions",
    62: "No Default",
    65: "Notices",
    66: "Organization And Authority",
    67: "Participation Rights",
    68: "Interest Computation",
    71: "Public Announcements",
    73: "Records And Inspection",
    74: "Release Requirements",
    75: "Guaranty Obligations",
    76: "Tax Acknowledgment",
    79: "Severability",
    81: "Specific Performance",
    83: "Real Property Liens",
    84: "Successors And Assigns",
    85: "Survival Of Terms",
    87: "Tax Obligations",
    88: "Termination",
    89: "Lease Term",
    90: "Title To Property",
    91: "Transactions With Affiliates",
    92: "Use Of Proceeds",
    93: "Vacation And Benefits",
    95: "Restricted Stock Vesting",
    96: "Waiver Of Jury Trial",
    97: "Waiver Of Rights",
    98: "Warranty Disclaimer",
    99: "Tax Withholding",
}

def main():
    input_path = os.path.join("data", "contracts", "train.json")
    output_path = os.path.join("data", "contracts", "finetune_train.json")
    
    # Read train.json (one JSON object per line)
    data = []
    with open(input_path, 'r', encoding='utf-8') as f:
        for line in f:
            line = line.strip()
            if line:
                data.append(json.loads(line))
    
    print(f"📂 Loaded {len(data)} examples from train.json")
    
    # Count label distribution
    from collections import Counter
    label_counts = Counter(d['label'] for d in data)
    print(f"📊 Unique labels: {len(label_counts)}")
    
    # Map labels to names
    unmapped = set()
    labeled_data = []
    for item in data:
        label_id = item['label']
        label_name = EXTENDED_LABEL_MAP.get(label_id) 
        
        if label_name is None:
            unmapped.add(label_id)
            label_name = f"Category_{label_id}"
        
        labeled_data.append({
            "messages": [
                {
                    "role": "user",
                    "content": f"Classify this contract clause:\n\n{item['text']}"
                },
                {
                    "role": "assistant",
                    "content": label_name
                }
            ]
        })
    
    if unmapped:
        print(f"⚠️  Unmapped labels: {sorted(unmapped)}")
    
    # Show label distribution
    print("\n📊 Label Distribution:")
    label_names = [EXTENDED_LABEL_MAP.get(lbl, f"Category_{lbl}") for lbl in label_counts]
    name_counts = Counter()
    for item in data:
        name = EXTENDED_LABEL_MAP.get(item['label'], f"Category_{item['label']}")
        name_counts[name] += 1
    
    for name, count in name_counts.most_common():
        print(f"  {name}: {count}")
    
    # Save labeled dataset
    with open(output_path, 'w', encoding='utf-8') as f:
        json.dump(labeled_data, f, indent=2, ensure_ascii=False)
    
    print(f"\n✅ Saved {len(labeled_data)} labeled examples to {output_path}")
    print(f"📝 Example:")
    print(f"  User: {labeled_data[0]['messages'][0]['content'][:100]}...")
    print(f"  Label: {labeled_data[0]['messages'][1]['content']}")

if __name__ == "__main__":
    main()
