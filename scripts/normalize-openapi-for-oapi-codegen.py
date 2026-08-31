#!/usr/bin/env python3
"""Normalize OpenAPI 3.1 nullable unions for oapi-codegen."""

import re
import sys

spec = sys.stdin.read()
spec = re.sub(r'type:\n(\s*)- "null"\n\1- ([A-Za-z]+)', r'type: \2', spec)
spec = re.sub(r'type:\n(\s*)- ([A-Za-z]+)\n\1- "null"', r'type: \2', spec)
spec = spec.replace('type: "null"', 'type: string')
sys.stdout.write(spec)
