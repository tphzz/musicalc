import io
import json
from pathlib import Path
import unittest
from unittest.mock import patch, Mock
from urllib.error import HTTPError
from urllib.parse import parse_qs, urlsplit

import ios_app_store as asc


CONFIG = {"app_id": "1234567890", "api_key_id": "ABCDEFGHIJ",
          "api_issuer_id": "example-issuer", "key_directory": "/example/keys"}
APP = {"data": {"type": "apps", "id": CONFIG["app_id"], "attributes": {"bundleId": "io.example.musicalc"}}}


def builds(*values, next_url=None):
    return {"data": [{"type": "builds", "attributes": {"version": value}} for value in values],
            "links": {"self": "https://api.appstoreconnect.apple.com/v1/builds", "next": next_url}}


class AppStoreTests(unittest.TestCase):
    def latest(self, responses):
        with patch.object(asc, "api_token", return_value="secret-token"), \
             patch.object(asc, "request_json", side_effect=responses) as request:
            value = asc.latest_build(CONFIG, "io.example.musicalc", "0.10.0")
            return value, request.call_args_list

    def test_new_version_has_zero_high_water_mark(self):
        value, calls = self.latest([APP, builds()])
        self.assertEqual(value, (0, 0, 0))
        query = parse_qs(urlsplit(calls[1].args[0]).query)
        self.assertEqual(query["filter[app]"], [CONFIG["app_id"]])
        self.assertEqual(query["filter[preReleaseVersion.version]"], ["0.10.0"])
        self.assertEqual(query["filter[preReleaseVersion.platform]"], ["IOS"])

    def test_every_page_is_compared_numerically(self):
        next_url = "https://api.appstoreconnect.apple.com/v1/builds?cursor=next"
        value, calls = self.latest([APP, builds("9", next_url=next_url), builds("10", "2")])
        self.assertEqual(value, (10, 0, 0))
        self.assertEqual(calls[2].args[0], next_url)

    def test_timestamp_and_dotted_builds_are_compared_as_numeric_components(self):
        value, _ = self.latest([APP, builds("9999", "20261102120000", "20261102120000.1")])
        self.assertEqual(value, (20261102120000, 1, 0))
        value, _ = self.latest([APP, builds("9.99.99", "10.1", "10.0.10")])
        self.assertEqual(value, (10, 1, 0))

    def test_query_relative_pagination_keeps_the_current_endpoint(self):
        value, calls = self.latest([APP, builds("9", next_url="?cursor=next"), builds("10")])
        self.assertEqual(value, (10, 0, 0))
        self.assertEqual(calls[2].args[0], "https://api.appstoreconnect.apple.com/v1/builds?cursor=next")

    def test_wrong_app_bundle_fails_before_build_query(self):
        other = {"data": {**APP["data"], "attributes": {"bundleId": "other.app"}}}
        with self.assertRaisesRegex(ValueError, "bundle"):
            self.latest([other])

    def test_malformed_pages_and_unsupported_numbers_fail_closed(self):
        for page in ({}, {"data": {}}, builds("1.2.3.4"), builds("-1"), builds("1..2"),
                     builds("abc"), {"data": [None]}, {"data": [], "links": []},
                     {"data": []}, {"data": [], "links": {}},
                     {"data": [], "links": {"self": ""}},
                     builds("2", next_url=""), builds("2", next_url=" "), builds("2", next_url=7)):
            with self.subTest(page=page), self.assertRaises(ValueError):
                self.latest([APP, page])

    def test_repeated_page_is_rejected(self):
        url = "https://api.appstoreconnect.apple.com/v1/builds?cursor=loop"
        with self.assertRaisesRegex(ValueError, "pagination"):
            self.latest([APP, builds("2", next_url=url), builds("3", next_url=url)])


    def test_token_signer_failure_does_not_expose_captured_output(self):
        with patch.object(asc.subprocess, "run", return_value=Mock(
                returncode=1, stdout="private key contents", stderr="private diagnostic")):
            with self.assertRaisesRegex(ValueError, "team API key") as stopped:
                asc.api_token(CONFIG)
            self.assertNotIn("private key contents", str(stopped.exception))
            self.assertNotIn("private diagnostic", str(stopped.exception))

    def test_malformed_app_response_fails_before_build_query(self):
        for data in (None, [], {"id": CONFIG["app_id"], "attributes": None}):
            with self.subTest(data=data), self.assertRaises(ValueError):
                self.latest([{"data": data}])

    def test_request_sends_bearer_and_never_follows_redirects(self):
        response = Mock()
        response.__enter__ = Mock(return_value=io.BytesIO(b'{"data":[]}'))
        response.__exit__ = Mock(return_value=False)
        with patch.object(asc, "build_opener") as build:
            build.return_value.open.return_value = response
            self.assertEqual(asc.request_json("https://api.appstoreconnect.apple.com/v1/builds", "secret-token"), {"data": []})
            sent = build.return_value.open.call_args.args[0]
            self.assertEqual(sent.get_header("Authorization"), "Bearer secret-token")
            self.assertEqual(sent.get_method(), "GET")
            self.assertIsNone(build.call_args.args[0].redirect_request(None, None, 302, "", {}, "https://example.com"))

    def test_untrusted_pagination_url_never_receives_token(self):
        for url in ("https://example.com/v1/builds", "http://api.appstoreconnect.apple.com/v1/builds"):
            with self.subTest(url=url), patch.object(asc, "build_opener") as build:
                with self.assertRaises(ValueError):
                    asc.request_json(url, "secret-token")
                build.assert_not_called()

    def test_http_failure_reports_status_without_token_or_response_body(self):
        with patch.object(asc, "build_opener") as build:
            build.return_value.open.side_effect = HTTPError("https://api.appstoreconnect.apple.com/v1/builds",
                403, "secret-token", {}, io.BytesIO(b"private details"))
            with self.assertRaisesRegex(ValueError, "403") as stopped:
                asc.request_json("https://api.appstoreconnect.apple.com/v1/builds", "secret-token")
            self.assertNotIn("secret-token", str(stopped.exception))
            self.assertNotIn("private details", str(stopped.exception))


if __name__ == "__main__":
    unittest.main()
