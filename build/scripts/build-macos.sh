#!/usr/bin/env bash
set -euo pipefail
project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$project_root"
version_exports="$(python3 build/scripts/version.py --shell)"
eval "$version_exports"
export MACOSX_DEPLOYMENT_TARGET=13.3
export SDKROOT="$(xcrun --sdk macosx --show-sdk-path)"
clang_path="$(xcrun --sdk macosx --find clang)"
clangxx_path="$(xcrun --sdk macosx --find clang++)"
export CGO_ENABLED=1
export CGO_CFLAGS="-mmacosx-version-min=$MACOSX_DEPLOYMENT_TARGET"
export CGO_LDFLAGS="-mmacosx-version-min=$MACOSX_DEPLOYMENT_TARGET"
notarize="${MUSICALC_NOTARIZE:-false}"
[[ "$MUSICALC_RELEASE" != true ]] || notarize=true

# All disposable data lives in this checkout, separate from persistent credentials.
mkdir -p "$project_root/build/.cache"
work="$(mktemp -d "$project_root/build/.cache/macos.XXXXXX")"
trap 'rm -rf "$work"' EXIT
dist="$project_root/build/dist/macos"
mkdir -p "$dist"
app="$work/MusiCalc.app"
mkdir -p "$app/Contents/MacOS" "$app/Contents/Resources" "$work/appicon.iconset"
for arch in arm64 amd64; do
  clang_arch="$arch"
  [[ "$arch" != amd64 ]] || clang_arch=x86_64
  export CC="$clang_path -arch $clang_arch" CXX="$clangxx_path -arch $clang_arch"
  GOARCH="$arch" go test ./...
  GOOS=darwin GOARCH="$arch" go build -trimpath \
    -ldflags="-s -w -X main.version=$MUSICALC_VERSION" -o "$work/$arch" .
done
lipo -create "$work/arm64" "$work/amd64" -output "$app/Contents/MacOS/MusiCalc"
lipo "$app/Contents/MacOS/MusiCalc" -verify_arch arm64 x86_64
chmod 0755 "$app/Contents/MacOS/MusiCalc"
for size in 16 32 128 256 512; do
  sips -z "$size" "$size" icons/appicon.png --out "$work/appicon.iconset/icon_${size}x${size}.png" >/dev/null
  double=$((size * 2))
  sips -z "$double" "$double" icons/appicon.png --out "$work/appicon.iconset/icon_${size}x${size}@2x.png" >/dev/null
done
iconutil -c icns "$work/appicon.iconset" -o "$app/Contents/Resources/appicon.icns"
python3 - "$app/Contents/Info.plist" <<'PY'
import os,plistlib,sys
plist={"CFBundleDevelopmentRegion":"en","CFBundleDisplayName":"MusiCalc",
       "CFBundleExecutable":"MusiCalc","CFBundleIconFile":"appicon",
       "CFBundleIdentifier":"com.musicalc","CFBundleInfoDictionaryVersion":"6.0",
       "CFBundleName":"MusiCalc","CFBundlePackageType":"APPL",
       "CFBundleShortVersionString":os.environ["MUSICALC_MARKETING_VERSION"],
       "CFBundleVersion":os.environ["MUSICALC_MARKETING_VERSION"],
       "LSMinimumSystemVersion":os.environ["MACOSX_DEPLOYMENT_TARGET"],
       "NSHighResolutionCapable":True}
with open(sys.argv[1],"wb") as out: plistlib.dump(plist,out)
PY
plutil -lint "$app/Contents/Info.plist"
for arch in arm64 x86_64; do
  vtool -arch "$arch" -show-build "$app/Contents/MacOS/MusiCalc" > "$work/min-$arch.txt"
  grep -Eq 'minos +13\.3' "$work/min-$arch.txt"
done
otool -L "$app/Contents/MacOS/MusiCalc" > "$work/libraries.txt"
python3 - "$work/libraries.txt" <<'PY'
import sys
for line in open(sys.argv[1]):
    if line.startswith("\t") and not line.strip().startswith(("/System/Library/","/usr/lib/")):
        sys.exit("Unexpected non-system dynamic library: "+line.strip())
PY
if [[ "$notarize" == true ]]; then
  source build/scripts/apple-signing.sh
  unlock_signing_keychain
  identity="$(select_developer_identity)"
  xcrun notarytool history --keychain-profile developer-id-notary --keychain "$signing_keychain" >/dev/null
  codesign --force --timestamp --options runtime --keychain "$signing_keychain" --sign "$identity" "$app"
  codesign --verify --deep --strict --verbose=2 "$app"
  ditto -c -k --keepParent "$app" "$work/notarize.zip"
  xcrun notarytool submit "$work/notarize.zip" --keychain-profile developer-id-notary \
    --keychain "$signing_keychain" --wait --output-format json > "$work/app-notary.json"
  python3 -c 'import json,sys; sys.exit(0 if json.load(open(sys.argv[1])).get("status")=="Accepted" else "App notarization rejected")' "$work/app-notary.json"
  xcrun stapler staple "$app"
  xcrun stapler validate "$app"
  spctl --assess --type execute --verbose=2 "$app"
fi
zip="$dist/MusiCalc_${MUSICALC_VERSION}_universal.app.zip"
dmg="$dist/musicalc_${MUSICALC_VERSION}_universal.dmg"
ditto -c -k --keepParent "$app" "$zip"
mkdir "$work/dmg"
ditto "$app" "$work/dmg/MusiCalc.app"
ln -s /Applications "$work/dmg/Applications"
hdiutil create -volname MusiCalc -srcfolder "$work/dmg" -ov -format UDZO "$dmg"
hdiutil verify "$dmg"
if [[ "$notarize" == true ]]; then
  codesign --force --timestamp --keychain "$signing_keychain" --sign "$identity" "$dmg"
  xcrun notarytool submit "$dmg" --keychain-profile developer-id-notary \
    --keychain "$signing_keychain" --wait --output-format json > "$work/dmg-notary.json"
  python3 -c 'import json,sys; sys.exit(0 if json.load(open(sys.argv[1])).get("status")=="Accepted" else "DMG notarization rejected")' "$work/dmg-notary.json"
  xcrun stapler staple "$dmg"
  xcrun stapler validate "$dmg"
fi
echo "Created $zip and $dmg (notarized=$notarize)"
