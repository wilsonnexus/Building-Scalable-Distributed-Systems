import json, re
from collections import Counter

# reducer output
with open("final.json","r",encoding="utf-8") as f:
    reduced = json.load(f)

# local ground truth
txt = open("shakespeare-hamlet.txt","r",encoding="utf-8",errors="ignore").read().lower()
words = re.findall(r"[a-z']+", txt)
truth = Counter(words)

# detect format (dict vs list)
if isinstance(reduced, list):
    # if it’s [{"word": "...", "count": ...}, ...]
    tmp = {}
    for x in reduced:
        if isinstance(x, dict) and "word" in x and "count" in x:
            tmp[x["word"]] = int(x["count"])
    reduced = tmp

# compare
mismatches = 0
for w,c in truth.items():
    rc = int(reduced.get(w, 0))
    if rc != c:
        mismatches += 1
        if mismatches <= 20:
            print("Mismatch:", w, "truth", c, "reduced", rc)

print("Total unique words truth:", len(truth))
print("Total unique words reduced:", len(reduced))
print("Total mismatched words:", mismatches)
print("PASS" if mismatches == 0 else "NOT EXACT MATCH")
