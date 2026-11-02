"""Validate the upload destination and existing iOS build versions in App Store Connect."""
import json
from pathlib import Path
import re
import subprocess
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode, urljoin, urlsplit
from urllib.request import build_opener, HTTPRedirectHandler, Request


API = "https://api.appstoreconnect.apple.com"


class NoRedirect(HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def api_token(config):
    key = Path(config["key_directory"]).expanduser() / f"AuthKey_{config['api_key_id']}.p8"
    result = subprocess.run(
        ["xcrun", "swift", str(Path(__file__).with_name("ios-asc-token.swift")),
         str(key), config["api_key_id"], config["api_issuer_id"]],
        capture_output=True, text=True, timeout=120,
    )
    token = result.stdout.strip()
    if result.returncode or not re.fullmatch(r"[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+", token):
        raise ValueError("Cannot create App Store Connect token; check the configured team API key")
    return token


def request_json(url, token):
    parsed = urlsplit(url)
    if parsed.scheme != "https" or parsed.netloc != "api.appstoreconnect.apple.com":
        raise ValueError("Unexpected App Store Connect pagination URL")
    request = Request(url, headers={"Authorization": "Bearer " + token, "Accept": "application/json"})
    try:
        with build_opener(NoRedirect()).open(request, timeout=30) as response:
            value = json.load(response)
    except HTTPError as exc:
        raise ValueError(f"App Store Connect HTTP {exc.code}; check app access and API credentials") from None
    except (URLError, TimeoutError):
        raise ValueError("Cannot reach App Store Connect") from None
    if not isinstance(value, dict):
        raise ValueError("Invalid App Store Connect response")
    return value


def latest_build(config, bundle, version):
    """Return the latest CFBundleVersion as three numeric components for upload checks."""
    token = api_token(config)
    app_id = str(config["app_id"])
    app = request_json(API + f"/v1/apps/{app_id}?" + urlencode({"fields[apps]": "bundleId"}), token)
    data = app.get("data")
    attributes = data.get("attributes") if isinstance(data, dict) else None
    if (not isinstance(data, dict) or data.get("id") != app_id
            or not isinstance(attributes, dict) or attributes.get("bundleId") != bundle):
        raise ValueError("App Store Connect app ID and configured bundle ID do not match")
    url = API + "/v1/builds?" + urlencode({
        "filter[app]": app_id, "filter[preReleaseVersion.version]": version,
        "filter[preReleaseVersion.platform]": "IOS", "fields[builds]": "version", "limit": 200,
    })
    highest = (0, 0, 0)
    visited = set()
    while url:
        if url in visited:
            raise ValueError("Repeated App Store Connect pagination URL")
        visited.add(url)
        page = request_json(url, token)
        rows = page.get("data")
        links = page.get("links")
        if not isinstance(rows, list) or not isinstance(links, dict):
            raise ValueError("Invalid App Store Connect build list")
        self_link = links.get("self")
        if not isinstance(self_link, str) or not self_link or any(c.isspace() for c in self_link):
            raise ValueError("Invalid App Store Connect pagination self link")
        for row in rows:
            attributes = row.get("attributes") if isinstance(row, dict) else None
            number = attributes.get("version") if isinstance(attributes, dict) else None
            if not isinstance(number, str) or not re.fullmatch(r"[0-9]+(?:\.[0-9]+){0,2}", number):
                raise ValueError("App Store Connect build versions must have one to three numeric components")
            parts = tuple(map(int, number.split(".")))
            highest = max(highest, parts + (0,) * (3 - len(parts)))
        following = links.get("next")
        if following is not None and (
                not isinstance(following, str) or not following or any(c.isspace() for c in following)):
            raise ValueError("Invalid App Store Connect pagination URL")
        url = urljoin(url, following) if following is not None else None
    return highest
