import argparse
import json
import csv
import re
from datetime import datetime
from collections import Counter
from typing import List, Dict, Any, Optional
import statistics

class LogParser:
    APACHE_REGEX = r'(?P<ip>\S+) \S+ \S+ \[(?P<time>.*?)\] ".*?" (?P<status>\d{3}) (?P<size>\d+|-)(?: "(?P<ref>.*?)" "(?P<ua>.*?)")?'
    
    @staticmethod
    def parse_line(line: str) -> Optional[Dict[str, Any]]:
        try:
            if line.startswith('{'):
                return json.loads(line)
            match = re.match(LogParser.APACHE_REGEX, line)
            if match:
                return match.groupdict()
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

    def get_top_endpoints(self, n: int = 10) -> List[tuple]:
        return Counter(log.get('request', '').split()[1] for log in self.logs if 'request' in log).most_common(n)

    def get_response_time_percentiles(self) -> Dict[str, float]:
        times = [float(log['response_time']) for log in self.logs if 'response_time' in log]
        if not times: return {}
        return {
            "p50": statistics.median(times),
            "p95": sorted(times)[int(len(times) * 0.95)]
        }

def format_table(data: Dict[str, Any]) -> str:
    lines = ["+------------------+-------+", "| Metric           | Value |", "+------------------+-------+"]
    for k, v in data.items():
        lines.append(f"| {str(k):<16} | {str(v):<5} |")
    lines.append("+------------------+-------+")
    return "\n".join(lines)

def main():
    parser = argparse.ArgumentParser(description="Log Analyzer Tool")
    parser.add_argument("--file", required=True, help="Path to log file")
    parser.add_argument("--format", choices=["table", "json", "csv"], default="table")
    parser.add_argument("--limit", type=int, default=10)
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

    if args.format == "json":
        print(json.dumps(results, indent=2))
    elif args.format == "csv":
        writer = csv.writer(import sys; sys.stdout)
        for k, v in results.items():
            writer.writerow([k, v])
    else:
        print(format_table({k: len(v) if isinstance(v, list) else v for k, v in results.items()}))

if __name__ == "__main__":
    main()