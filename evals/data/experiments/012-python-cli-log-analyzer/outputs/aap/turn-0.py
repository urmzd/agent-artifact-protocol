<aap:target id="log-analyzer-script">import argparse
import json
import re
import csv
from collections import Counter, defaultdict
from datetime import datetime
from typing import List, Dict, Any, Optional

<aap:target id="log-parser-logic">
def parse_line(line: str) -> Dict[str, Any]:
    # Basic regex for Combined Log Format (Apache/Nginx)
    regex = r'(?P<ip>\S+) \S+ \S+ \[(?P<timestamp>.*?)\] "(?P<method>\S+) (?P<path>\S+) \S+" (?P<status>\d+) (?P<size>\S+)(?: (?P<response_time>\d+))?'
    match = re.search(regex, line)
    if match:
        data = match.groupdict()
        data['status'] = int(data['status'])
        data['response_time'] = float(data['response_time'] or 0)
        return data
    try:
        return json.loads(line)
    except json.JSONDecodeError:
        return {}
</aap:target>

<aap:target id="analysis-functions">
def analyze_logs(logs: List[Dict[str, Any]], limit: int = 10) -> Dict[str, Any]:
    ips = Counter(log.get('ip') for log in logs if 'ip' in log)
    statuses = Counter(log.get('status') for log in logs if 'status' in log)
    paths = Counter(log.get('path') for log in logs if 'path' in log)
    times = sorted([log.get('response_time', 0) for log in logs])
    
    return {
        "top_ips": ips.most_common(limit),
        "status_dist": dict(statuses),
        "top_paths": paths.most_common(limit),
        "p95": times[int(len(times) * 0.95)] if times else 0
    }
</aap:target>

<aap:target id="formatters">
def format_output(data: Dict[str, Any], fmt: str):
    if fmt == 'json':
        print(json.dumps(data, indent=2))
    elif fmt == 'csv':
        writer = csv.writer(import sys).stdout
        for key, value in data.items():
            writer.writerow([key, value])
    else:
        for k, v in data.items():
            print(f"{k.upper()}: {v}")
</aap:target>

<aap:target id="main-execution">
def main():
    parser = argparse.ArgumentParser(description="Log Analyzer")
    parser.add_argument("file", help="Path to log file")
    parser.add_argument("--format", choices=['table', 'json', 'csv'], default='table')
    parser.add_argument("--limit", type=int, default=10)
    args = parser.parse_args()

    logs = []
    with open(args.file, 'r') as f:
        for line in f:
            parsed = parse_line(line)
            if parsed: logs.append(parsed)

    results = analyze_logs(logs, args.limit)
    format_output(results, args.format)

if __name__ == "__main__":
    main()
</aap:target>
</aap:target>