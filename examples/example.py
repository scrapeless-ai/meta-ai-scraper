"""Submit the sample question, or read the answer for an existing task ID."""

import json
import os
import sys
from urllib.error import HTTPError
from urllib.parse import quote
from urllib.request import Request, urlopen

API = "https://api.scrapeless.com/api/v2/scraper"
PAYLOAD = {
    "actor": "scraper.meta",
    "input": {"prompt": "Which public museums in Chicago have architecture exhibits?", "country": "US"},
}


def main():
    token = os.environ.get("SCRAPELESS_API_TOKEN")
    if not token:
        sys.exit("Set SCRAPELESS_API_TOKEN before running this example.")
    if len(sys.argv) > 2:
        sys.exit("Usage: python3 examples/example.py [task_id]")

    create = len(sys.argv) == 1
    path = "/request" if create else "/result/" + quote(sys.argv[1], safe="")
    request = Request(
        API + path,
        data=json.dumps(PAYLOAD).encode() if create else None,
        headers={"Content-Type": "application/json", "x-api-token": token},
    )
    try:
        with urlopen(request, timeout=60) as response:
            data = json.load(response)
    except HTTPError as error:
        sys.exit(f"HTTP {error.code}: {error.read().decode(errors='replace')}")

    if create:
        print(data["task_id"])
        return
    status = data["status"]
    if status in ("pending", "running"):
        print(status)
        return
    if status != "success":
        sys.exit(data.get("message") or f"Task status: {status}")

    result = data["task_result"]
    print(result["full_response"])
    for source in result.get("references") or []:
        print(f"- {source['title']}: {source['url']}")


if __name__ == "__main__":
    main()
