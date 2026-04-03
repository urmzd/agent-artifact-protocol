import argparse
import json
import csv
import re
import sys
from datetime import datetime
from collections import Counter, defaultdict
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
                if data.get('response_time'): data['response_time'] = float(data['response_time'])
                time_str = data.get('time', '').split(' ')[0]
                try:
                    data['dt'] = datetime.strptime(time_str, '%d/%b/%Y:%H:%M:%S')
                except:
                    data['dt'] = datetime.now()
                return data
        except Exception:
            return None
        return None

class LogAnalyzer:
    def __init__(self, logs: List[Dict[str, Any]]):
        self.logs = logs

    def get_group_counts(self, group_by: str) -> Dict[str, int]:
        counts = defaultdict(int)
        for log in self.logs:
            dt = log.get('dt', datetime.now())
            if group_by == 'hour': key = dt.strftime('%Y-%m-%d %H:00')
            elif group_by == 'day': key = dt.strftime('%Y-%m-%d')
            elif group_by == 'week': key = dt.strftime('%Y-W%W')
            else: key = 'total'
            counts[key] += 1
        return dict(counts)

    def detect_anomalies(self) -> List[Dict[str, Any]]:
        times = [log['response_time'] for log in self.logs if isinstance(log.get('response_time'), (int, float))]
        if not times: return []
        p99 = sorted(times)[int(len(times) * 0.99)]
        return [log for log in self.logs if log.get('response_time', 0) > p99]

def format_box_table(data: Dict[str, Any]) -> str:
    lines = []
    # Box drawing characters
    top = "┌──────────────────┬──────────┐"
    mid = "├──────────────────┼──────────┤"
    bot = "└──────────────────┴──────────┘"
    
    lines.append(top)
    lines.append("│ Metric           │ Value    │")
    lines.append(mid)
    
    for k, v in data.items():
        val = str(len(v) if isinstance(v, list) else v)
        lines.append(f"│ {str(k):<16} │ {val:<8} │")
    
    lines.append(bot)
    return "\n".join(lines)

def main():
    parser = argparse.ArgumentParser(description="Log Analyzer Tool")
    parser.add_argument("--file", required=True)
    parser.add_argument("--format", choices=["table", "json", "csv"], default="table")
    parser.add_argument("--group-by", choices=["hour", "day", "week"])
    parser.add_argument("--detect-anomalies", action="store_true")
    args = parser.parse_args()

    logs = []
    with open(args.file, 'r') as f:
        for line in f:
            parsed = LogParser.parse_line(line)
            if parsed: logs.append(parsed)

    analyzer = LogAnalyzer(logs)
    results = {}
    
    if args.group_by:
        results[f"counts_{args.group_by}"] = analyzer.get_group_counts(args.group_by)
    if args.detect_anomalies:
        results["anomalies"] = analyzer.detect_anomalies()

    if args.format == "json":
        print(json.dumps(results, indent=2, default=str))
    elif args.format == "csv":
        writer = csv.writer(sys.stdout)
        for k, v in results.items():
            writer.writerow([k, str(v)])
    else:
        print(format_box_table({k: v for k, v in results.items()}))

if __name__ == "__main__":
    main()