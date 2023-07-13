import sys
import json
from pprint import pprint
from datetime import datetime
from dateutil.parser import parse
from collections import defaultdict


def calculate_utilization(data):
  utilization = 0
  for item in data:
    utilization += (item['returnedAt'] - item['leasedAt']).total_seconds()
  return utilization


def main():
  if len(sys.argv) < 2:
    print("Usage: python time_utilization.py <filename>")
    sys.exit(1)

  filename = sys.argv[1]

  with open(filename, 'r') as f:
    data = [json.loads(line) for line in f.readlines() if line.strip()]

  for item in data:
    item['leasedAt'] = parse(item['leasedAt'])
    item['returnedAt'] = parse(item['returnedAt'])

  data.sort(key=lambda x: x['leasedAt'])

  logs_per_pod = defaultdict(list)
  for item in data:
    logs_per_pod[item['podId']].append(item)

  time_window = (data[-1]['returnedAt'] - data[0]['leasedAt']).total_seconds()

  utilization_per_pod = {}
  for pod, logs in logs_per_pod.items():
    utilization_per_pod[pod] = calculate_utilization(logs) / time_window

  pprint(utilization_per_pod)

  total_utilization = calculate_utilization(data)
  print("Total utilization: ", total_utilization / time_window)


if __name__ == '__main__':
  main()
