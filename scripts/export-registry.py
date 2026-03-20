#!/usr/bin/env python3
"""
Reads the clinic registry JSON files and outputs a single combined JSON file
suitable for the marketing site (or anything else that wants a flat snapshot).

Usage:
    python3 scripts/export-registry.py > ../clinic-marketing-site/client/public/registry.json
    python3 scripts/export-registry.py -o ../clinic-marketing-site/client/public/registry.json
"""

import json
import sys
import argparse
from pathlib import Path

# Category display names — controls ordering and how they appear on the site
CATEGORY_LABELS = {
    "cloud": "Cloud Providers",
    "deploy": "Deployment",
    "iac": "Infrastructure",
    "k8s": "Kubernetes",
    "payments": "Payments",
    "observability": "Observability",
    "backend": "Backend",
    "utility": "Utilities",
    "secrets": "Secrets",
    "social": "Social Media",
    "productivity": "Productivity",
    "media": "Media",
    "finance": "Finance",
    "news": "News",
    "ecommerce": "E-Commerce",
}

def main():
    parser = argparse.ArgumentParser(description="Export clinic registry to a single JSON file")
    parser.add_argument("-o", "--output", help="Output file path (default: stdout)")
    args = parser.parse_args()

    registry_dir = Path(__file__).resolve().parent.parent / "registry"
    index_path = registry_dir / "index.json"

    if not index_path.exists():
        print(f"Error: {index_path} not found", file=sys.stderr)
        sys.exit(1)

    index = json.loads(index_path.read_text())

    tools = []
    for name in sorted(index["tools"]):
        tool_path = registry_dir / "tools" / f"{name}.json"
        if not tool_path.exists():
            print(f"Warning: {tool_path} not found, skipping", file=sys.stderr)
            continue
        tool = json.loads(tool_path.read_text())
        tools.append({
            "cmd": tool["command"],
            "name": tool["name"],
            "desc": tool["description"],
            "category": CATEGORY_LABELS.get(tool["category"], tool["category"]),
        })

    stacks = []
    for name in sorted(index["stacks"]):
        stack_path = registry_dir / "stacks" / f"{name}.json"
        if not stack_path.exists():
            print(f"Warning: {stack_path} not found, skipping", file=sys.stderr)
            continue
        stack = json.loads(stack_path.read_text())
        stacks.append({
            "name": stack["name"],
            "description": stack["description"],
            "tools": stack["tools"],
        })

    output = {
        "schema_version": index["schema_version"],
        "tools": tools,
        "stacks": stacks,
    }

    result = json.dumps(output, indent=2) + "\n"

    if args.output:
        Path(args.output).write_text(result)
        print(f"Wrote {len(tools)} tools and {len(stacks)} stacks to {args.output}", file=sys.stderr)
    else:
        sys.stdout.write(result)


if __name__ == "__main__":
    main()
