import base64
import json
from pathlib import Path
import subprocess
import sys
import tempfile
import time
import unittest


@unittest.skipUnless(sys.platform == "darwin", "requires macOS CryptoKit and Xcode")
class AppStoreTokenTests(unittest.TestCase):
    def test_native_team_token_has_valid_es256_signature_and_claims(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary)
            key, public = directory / "key.p8", directory / "public.pem"
            subprocess.run(["openssl", "genpkey", "-algorithm", "EC",
                            "-pkeyopt", "ec_paramgen_curve:prime256v1", "-out", str(key)],
                           check=True, capture_output=True, timeout=30)
            subprocess.run(["openssl", "pkey", "-in", str(key), "-pubout", "-out", str(public)],
                           check=True, capture_output=True, timeout=30)
            result = subprocess.run(
                ["xcrun", "swift", str(Path(__file__).with_name("ios-asc-token.swift")),
                 str(key), "0123456789", "example-issuer"],
                capture_output=True, text=True, timeout=120,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            header, payload, signature = result.stdout.strip().split(".")
            decode = lambda value: base64.urlsafe_b64decode(value + "=" * (-len(value) % 4))
            self.assertEqual(json.loads(decode(header)), {"alg": "ES256", "kid": "0123456789", "typ": "JWT"})
            claims = json.loads(decode(payload))
            self.assertEqual((claims["iss"], claims["aud"]), ("example-issuer", "appstoreconnect-v1"))
            self.assertLessEqual(claims["iat"], time.time())
            self.assertGreater(claims["exp"], time.time())
            self.assertEqual(claims["exp"] - claims["iat"], 300)
            raw = decode(signature)
            self.assertEqual(len(raw), 64)
            # OpenSSL independently verifies the JOSE r||s signature, encoded as DER.
            integers = []
            for part in (raw[:32], raw[32:]):
                value = part.lstrip(b"\0") or b"\0"
                if value[0] & 0x80:
                    value = b"\0" + value
                integers.append(b"\x02" + bytes([len(value)]) + value)
            body = b"".join(integers)
            der = directory / "signature.der"
            der.write_bytes(b"\x30" + bytes([len(body)]) + body)
            verified = subprocess.run(
                ["openssl", "dgst", "-sha256", "-verify", str(public), "-signature", str(der)],
                input=(header + "." + payload).encode("ascii"), capture_output=True, timeout=30,
            )
            self.assertEqual(verified.returncode, 0, verified.stderr)


if __name__ == "__main__":
    unittest.main()
