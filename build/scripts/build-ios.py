#!/usr/bin/env python3
"""Package an iPhone IPA and optionally upload it to App Store Connect."""
import argparse
import datetime
import hashlib
import json
import os
from pathlib import Path
import plistlib
import re
import shutil
import stat
import subprocess
import sys
import tempfile
from ios_build_number import current_build
from ios_app_store import latest_build
from version import ROOT, current


def run(*args, capture=False, env=None):
    result = subprocess.run(list(map(str, args)), cwd=ROOT, check=True,
                            stdout=subprocess.PIPE if capture else None, env=env)
    return result.stdout if capture else b""


def verify_bundle(info, profile, entitlements, bundle, team, version, build):
    expected = {"CFBundleIdentifier": bundle, "CFBundleShortVersionString": version,
                "CFBundleVersion": str(build), "UIDeviceFamily": [1], "MinimumOSVersion": "15.0"}
    for key, value in expected.items():
        if info.get(key) != value:
            raise ValueError(f"Unexpected iPhone bundle field: {key}")
    if int(str(info.get("DTSDKName", "")).removeprefix("iphoneos").split(".")[0] or "0") < 26:
        raise ValueError("App must be built with iOS 26 SDK or newer")
    if profile.get("TeamIdentifier") != [team]:
        raise ValueError("Provisioning profile team does not match configuration")
    expiry = profile.get("ExpirationDate")
    if not isinstance(expiry, datetime.datetime) or expiry <= datetime.datetime.now(datetime.timezone.utc).replace(tzinfo=None):
        raise ValueError("Provisioning profile has expired")
    profile_entitlements = profile.get("Entitlements", {})
    for values in (profile_entitlements, entitlements):
        if values.get("application-identifier") != f"{team}.{bundle}":
            raise ValueError("Expected an explicit app identifier in distribution entitlements")
        if values.get("get-task-allow") is not False or values.get("beta-reports-active") is not True:
            raise ValueError("App Store distribution/TestFlight entitlements are missing")
    if entitlements.get("com.apple.developer.team-identifier") != team:
        raise ValueError("Signed application team identifier is incorrect")


def verify_private_key(private_key, runner_uid):
    private_key = private_key.absolute()
    try:
        metadata = private_key.stat()
    except FileNotFoundError:
        raise ValueError(
            f"ASC private key not found: {private_key}. Check key_directory and api_key_id in the runner config"
        ) from None
    except OSError as error:
        raise ValueError(f"Cannot inspect ASC private key {private_key}: {error.strerror}") from None
    if not stat.S_ISREG(metadata.st_mode):
        raise ValueError(f"ASC private key must be a regular file: {private_key}")
    problems = []
    if metadata.st_uid != runner_uid:
        problems.append(f"owner uid {metadata.st_uid}, expected runner uid {runner_uid}")
    mode = stat.S_IMODE(metadata.st_mode)
    if mode != 0o600:
        problems.append(f"mode {mode:04o}, expected 0600")
    if problems:
        raise ValueError(f"ASC private key {private_key}: " + "; ".join(problems))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--upload", action="store_true")
    args = parser.parse_args()
    if sys.platform != "darwin":
        parser.error("iPhone packaging requires the configured macOS runner")
    directory = Path.home() / ".config/musicalc-ios"
    config_path = directory / "config.json"
    if not config_path.is_file():
        parser.error(
            f"Missing runner-local iOS config: {config_path}. "
            "Create it as the runner account; see build/docs/README-APPLE.md"
        )
    if config_path.stat().st_uid != os.getuid() or stat.S_IMODE(config_path.stat().st_mode) != 0o600:
        parser.error("iOS config must be owned by the runner and mode 600")
    config = json.loads(config_path.read_text(encoding="utf-8"))
    bundle, team = config["bundle_id"], config["team_id"]
    if not re.fullmatch(r"[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)+", bundle) or not re.fullmatch(r"[A-Z0-9]{10}", team):
        parser.error("Invalid bundle or team identifier")
    certificate = config["certificate"]
    if not certificate.startswith("Apple Distribution: ") or not certificate.endswith(f"({team})"):
        parser.error("Configure the exact Apple Distribution certificate common name")
    identities = run("security", "find-identity", "-v", "-p", "codesigning", capture=True).decode()
    matches = re.findall(r'([A-Fa-f0-9]{40})\s+"' + re.escape(certificate) + r'"', identities)
    if len(set(matches)) != 1:
        parser.error("Expected one valid configured Apple Distribution identity with a private key")
    identity = matches[0]
    xcode = run("xcodebuild", "-version", capture=True).decode()
    sdk = run("xcrun", "--sdk", "iphoneos", "--show-sdk-version", capture=True).decode().strip()
    if int(re.search(r"Xcode (\d+)", xcode)[1]) < 26 or int(sdk.split(".")[0]) < 26:
        parser.error("Install/select Xcode 26+ and the iOS 26+ SDK before packaging")
    env = os.environ.copy()
    if args.upload:
        for key in ("app_id", "api_key_id", "api_issuer_id", "key_directory"):
            if not config.get(key):
                parser.error("Missing App Store Connect setting: " + key)
        if not re.fullmatch(r"[0-9]+", str(config["app_id"])):
            parser.error("app_id must be the numeric App Store Connect app ID")
        if not re.fullmatch(r"[A-Z0-9]{10}", str(config["api_key_id"])):
            parser.error("api_key_id must be the 10-character team API key ID")
        key_directory = Path(config["key_directory"]).expanduser()
        private_key = key_directory / f"AuthKey_{config['api_key_id']}.p8"
        try:
            verify_private_key(private_key, os.getuid())
        except ValueError as error:
            parser.error(str(error))
        env["API_PRIVATE_KEYS_DIR"] = str(key_directory)
    data = current()
    version = data["marketing_version"]
    tool = ROOT / "build/.cache/fyne-ios"
    tool.parent.mkdir(parents=True, exist_ok=True)
    icon = tool.parent / "ios-appicon.png"
    run("xcrun", "swift", ROOT / "build/scripts/ios-icon.swift", ROOT / "icons/appicon.png", icon)
    build = current_build()
    if args.upload:
        highest = latest_build(config, bundle, version)
        if (build, 0, 0) <= highest:
            previous = ".".join(map(str, highest))
            parser.error(f"Generated build {build} is not newer than App Store build {previous}; check the runner UTC clock")
    print(f"Building iPhone {version} build {build}", flush=True)
    run(sys.executable, ROOT / "build/scripts/prepare-ios-tool.py")
    run(tool, "release", "--os", "ios", "--name", "MusiCalc", "--app-id", bundle,
        "--app-version", version, "--app-build", build, "--icon", icon,
        "--certificate", certificate, "--profile", config["profile"])
    original = ROOT / "MusiCalc.ipa"
    cache = ROOT / "build/.cache"
    with tempfile.TemporaryDirectory(prefix="ipa-", dir=cache) as temporary:
        stage = Path(temporary)
        run("ditto", "-x", "-k", original, stage)
        app = stage / "Payload/MusiCalc.app"
        info = plistlib.loads((app / "Info.plist").read_bytes())
        profile = plistlib.loads(run("security", "cms", "-D", "-i", app / "embedded.mobileprovision", capture=True))
        entitlements = run("codesign", "--display", "--entitlements", ":-", app, capture=True)
        signed = plistlib.loads(entitlements)
        verify_bundle(info, profile, signed, bundle, team, version, build)
        allowed = {hashlib.sha1(cert).hexdigest().upper() for cert in profile["DeveloperCertificates"]}
        if identity.upper() not in allowed:
            parser.error("Signing certificate is absent from the embedded provisioning profile")
        shutil.copy2(ROOT / "build/packaging/PrivacyInfo.xcprivacy", app / "PrivacyInfo.xcprivacy")
        entitlement_file = stage / "entitlements.plist"
        entitlement_file.write_bytes(entitlements)
        run("codesign", "--force", "--sign", identity, "--entitlements", entitlement_file,
            "--generate-entitlement-der", app)
        run("codesign", "--verify", "--deep", "--strict", app)
        run("plutil", "-lint", app / "PrivacyInfo.xcprivacy")
        executable = app / info["CFBundleExecutable"]
        run("lipo", executable, "-verify_arch", "arm64")
        load_commands = run("vtool", "-show-build", executable, capture=True).decode()
        if not re.search(r"platform\s+IOS\b", load_commands) or not re.search(r"minos\s+15\.0\b", load_commands):
            parser.error("Unexpected iPhone Mach-O platform/deployment target")
        final = ROOT / f"build/dist/ios/MusiCalc_{version}_{build}_iphone.ipa"
        final.parent.mkdir(parents=True, exist_ok=True)
        run("ditto", "-c", "-k", "--keepParent", stage / "Payload", final)
    original.unlink()
    metadata = {"bundle": bundle, "version": version, "build": build, "commit": data["commit"],
                "uploaded": False}
    if args.upload:
        auth = ["--apiKey", config["api_key_id"], "--apiIssuer", config["api_issuer_id"]]
        run("xcrun", "altool", "--validate-app", "-f", final, "-t", "ios", *auth, "--output-format", "json", env=env)
        command = ["xcrun", "altool", "--upload-package", final, "-t", "ios",
                   "--apple-id", config["app_id"], "--bundle-id", bundle,
                   "--bundle-short-version-string", version, "--bundle-version", build, *auth,
                   "--output-format", "json"]
        if config.get("asc_public_id"):
            command += ["--asc-public-id", config["asc_public_id"]]
        run(*command, env=env)
        metadata["uploaded"] = True
        print("Upload accepted. Check Apple processing and assign the build to the internal TestFlight group.")
    final.with_suffix(".json").write_text(json.dumps(metadata, indent=2) + "\n", encoding="utf-8")
    print(f"IPA: {final}")


if __name__ == "__main__":
    try:
        main()
    except (KeyError, ValueError, OSError, subprocess.SubprocessError) as exc:
        sys.exit(str(exc))
