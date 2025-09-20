"""
Copyright (c) 2024 OPNsense
All rights reserved.

DHCP AdGuard Sync log format
"""

import re
import time
from systemhealth.logformats import BaseLogFormat


class DhcpsyncLogFormat(BaseLogFormat):
    def __init__(self, filename):
        super(DhcpsyncLogFormat, self).__init__(filename)
        self._startup = time.time()

    def match(self, line):
        # Match lines that start with a timestamp and contain log level
        # Format: 2025/09/20 01:07:30 [INFO] Starting service...
        if 'dhcpsync' not in self._filename:
            return False

        # Check for timestamp pattern YYYY/MM/DD HH:MM:SS [LEVEL]
        return re.match(r'^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} \[[A-Z]+\]', line) is not None

    def timestamp(self, line):
        # Extract timestamp from line format: 2025/09/20 01:07:30 [INFO] message
        ts_match = re.match(r'^(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2})', line)
        if ts_match:
            try:
                # Parse the timestamp
                ts_str = ts_match.group(1)
                return time.mktime(time.strptime(ts_str, '%Y/%m/%d %H:%M:%S'))
            except ValueError:
                pass

        # Fallback to current time if parsing fails
        return self._startup

    def line(self, line):
        # Extract the log message part, removing timestamp and level
        # From: 2025/09/20 01:07:30 [INFO] Starting service...
        # To: Starting service...
        match = re.match(r'^\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} \[[A-Z]+\] (.*)$', line)
        if match:
            return match.group(1)
        return line

    def process_name(self, line):
        # Extract log level as process name for filtering
        # [INFO], [ERROR], [WARN], [DEBUG]
        level_match = re.search(r'\[([A-Z]+)\]', line)
        if level_match:
            return f"dhcpsync[{level_match.group(1)}]"
        return "dhcpsync"
