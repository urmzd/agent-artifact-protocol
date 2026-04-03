import argparse
import json
import csv
import re
import statistics
from datetime import datetime
from collections import Counter
from typing import List, Dict, Any, Optional

class LogParser:
    APACHE_REGEX = r'(?P<ip>\S+) \S+ \S+ \[(?P<time>.*?)\] "(?P<request>.*?)" (?P<status>\d{3}) (?P<size>\d+|-)(?: (?P<response_time>\d+))?'
    
    @staticmethod
    def parse_line(line: str) -> Optional[Dict[str, Any]]:
        try:
            if line.startswith('{'):
                return json.loads(line)
            match = re.match(LogParser.APACHE_REGEX, line)
            if match:
                data = match.groupdict()
                if data['response_time']: data['response_time'] = float(data['response_time'])
                return data
        except Exception:
            return None
        return None

class LogAnalyzer:
    def __init__(self, logs: List[Dict[str, Any]]):
        self.logs = logs

    def get_top_ips(self, n: int = 10) -> List[tuple]:
        return Counter(log.get('ip') for log in self.logs if 'ip' in log).most_common(n)

    def get_status_distribution(self) -> Dict[str, int]:
        return dict(Counter(str(log.get('status')) for log in self.logs if 'status' in log))

    def detect_anomalies(self) -> List[Dict[str, Any]]:
        times = [float(log['response_time']) for log in self.logs if 'response_time' in log]
        if not times: return []
        
        p99 = sorted(times)[int(len(times) * 0.99)]
        anomalies = [log for log in self.logs if log.get('response_time', 0) > p99]
        return anomalies

    def get_top_endpoints(self, n: int = 10) -> List[tuple]:
        return Counter(log.get('request', '').split()[1] for log in self.logs if 'request' in log).most_common(n)

def format_table(data: Dict[str, Any]) -> str:
    lines = ["+------------------+-------+", "| Metric           | Value |", "+------------------+-------+"]
    for k, v in data.items():
        val = len(v) if isinstance(v, list) else v
        lines.append(f"| {str(k):<16} | {str(val):<5} |")
    lines.append("+------------------+-------+")
    return "\n".join(lines)

def main():
    parser = argparse.ArgumentParser(description="Log Analyzer Tool")
    parser.add_argument("--file", required=True, help="Path to log file")
    parser.add_argument("--format", choices=["table", "json", "csv"], default="table")
    parser.add_argument("--limit", type=int, default=10)
    parser.add_argument("--detect-anomalies", action="store_true", help="Flag p99 latency spikes")
    args = parser.parse_args()

    logs = []
    with open(args.file, 'r') as f:
        for line in f:
            parsed = LogParser.parse_line(line)
            if parsed: logs.append(parsed)

    analyzer = LogAnalyzer(logs)
    results = {
        "top_ips": analyzer.get_top_ips(args.limit),
        "status_codes": analyzer.get_status_distribution(),
        "top_endpoints": analyzer.get_top_endpoints(args.limit)
    }
    
    if args.detect_anomalies:
        results["anomalies"] = analyzer.detect_anomalies()

    if args.format == "json":
        print(json.dumps(results, indent=2))
    elif args.format == "csv":
        writer = csv.writer(import sys; sys.stdout)
        for k, v in results.items():
            writer.writerow([k, str(v)])
    else:
        print(format_table(results))

if __name__ == "__main__":
    main()