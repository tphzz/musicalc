# Apple build runner and TestFlight setup

Use the runner account on the Apple Silicon Mac. Workflows select
`/Applications/Xcode.app/Contents/Developer`. macOS packages support **13.3+**.
iPhone builds target **iOS 15+** and require **Xcode 26+ / iOS 26 SDK+** for current
App Store Connect uploads.

## Repository checkout for manual commands

Automated builds use the scripts from the workflow's own
`actions/checkout` of the selected branch or tag. A separate manual clone,
branch switch or build-number seed command on the Mac is not required.

Select the desired ref in GitHub Actions when starting the aggregate or iPhone
workflow. Packaging reads that checkout's marketing version through
[version.py](../scripts/version.py), and reads the bundle ID from the runner's
configuration. Build identifiers are generated automatically from current UTC time.

Only manual development/debugging commands such as `build/scripts/build-ios.py`
need a terminal at a public MusiCalc checkout's root. Runner setup commands below
use absolute home-directory paths and do not require a checkout.

Persistent signing configuration and optional upload keys live outside checkout
under `$HOME/.config/musicalc-ios/`. GitHub Actions replaces source checkouts;
it must retain those credentials. Build numbering requires no persistent file.

## Existing macOS signing

Reuse the persistent login keychain:
`$HOME/Library/Keychains/login.keychain-db`.

The keychain password file is:
`$HOME/.config/developer-id-signing/login-keychain-password`.
It must be owned by the runner, mode 600, and contain one nonempty line.
Keep this file and the password out of repositories and workflow logs.

Install exactly one valid **Developer ID Application** identity with its private
key. The workflows unlock that keychain and configure signing access using the
same persistent-keychain model. The notarization credentials must already be
stored in the **developer-id-notary** keychain profile.

Tagged builds always sign with hardened runtime/timestamps, notarize the app,
staple it, build the ZIP/DMG, then sign/notarize/staple the DMG. Branch snapshots
can enable notarization with the workflow input; otherwise they are unsigned
development artifacts. Intel tests require Rosetta.

The job first tests architecture verification against real universal and thin
Mach-O fixtures. Packaging verifies ARM64 and Intel slices in separate `lipo`
invocations because Xcode 27 rejects the combined verification form. The two
architectures are still merged into one universal app.

~~~bash
mkdir -p build/.cache
bash build/scripts/build-macos.sh
# Optional for a branch snapshot:
MUSICALC_NOTARIZE=true bash build/scripts/build-macos.sh
~~~

## Signing keychain lifetime

Both Apple workflows use [apple-signing.sh](../scripts/apple-signing.sh) to unlock
the runner's login keychain, set its automatic-lock timeout to **two hours
(7200 seconds)**, and grant signing access. Locking on system sleep remains enabled.
This updates the keychain's saved settings; each signing run applies them again.
The timeout covers the Apple jobs' 90-minute limit, including compilation and
notarization waits. Any unlock, timeout-setting, or access-setting error stops
packaging. The password still comes from the runner-owned mode-600 file above.

A valid certificate chain does not prove that its private key is accessible.
If signing fails with `errSecInternalComponent`, test an immediate unlock and
signing retry as the runner account. For manual builds, set the same bounded
policy before starting (the first command prompts locally for the keychain
password):

~~~bash
security unlock-keychain "$HOME/Library/Keychains/login.keychain-db" &&
security set-keychain-settings -lut 7200 "$HOME/Library/Keychains/login.keychain-db"
~~~

See [Apple's keychain/signing troubleshooting](https://developer.apple.com/forums/thread/712005)
and the [security command's settings options](https://github.com/apple-oss-distributions/Security/blob/main/SecurityTool/macOS/security.c).

## CI signing preflight and keychain selection

The iPhone packager selects the runner's explicit login keychain for identity
lookup, the Xcode signing flags, Fyne's IPA signing, and the final signing after
privacy metadata is added. The patched tool receives the same path through
`MUSICALC_IOS_SIGNING_KEYCHAIN`, set by the packaging script. Existing Xcode
`OTHER_CODE_SIGN_FLAGS` from the environment are retained; the keychain path is
quoted for paths containing spaces.

Before compiling the Fyne tool or app, packaging prints the runner account,
keychain search list and login-keychain settings, then signs and verifies a
temporary copy of `/usr/bin/true` using the configured identity and explicit
keychain. It does not execute that copy. A failure stops before icon generation,
build-number generation, App Store Connect queries or tool compilation. The
check copies file contents without system metadata, uses disposable data under
`build/.cache/`, and prints no passwords or private-key data. Original copy/sign/
verify error details remain visible on failure.

A successful interactive-terminal signing test does not establish that a runner
service has the same keychain access. If the preflight fails only in Actions,
inspect its account/search-list/settings output and the runner's launchd service
context. If it passes and Xcode signing fails, retain the Xcode CodeSign command
and error. See [Apple's execution-context guidance](https://developer.apple.com/forums/thread/712005)
and [GitHub's macOS runner-service diagnostics](https://docs.github.com/en/actions/how-tos/manage-runners/self-hosted-runners/monitor-and-troubleshoot?platform=mac).

## iOS signing and notarization

The current iPhone workflow distributes through **App Store Connect / TestFlight**.
It uses an Apple Distribution identity and matching provisioning profile, then
validates/uploads the IPA for Apple's processing. This route has no separate
macOS `notarytool` submission or ticket-stapling step.

For direct macOS distribution, the workflow signs with Developer ID, submits to
Apple's macOS notarization service and staples the result. Apple's **iOS
alternative distribution** has a separate App Store Connect Notarization review
and alternative distribution package flow; that route is not configured here.

Apple references: [distribution methods](https://help.apple.com/xcode/mac/current/en.lproj/dev31de635e5.html),
[TestFlight distribution](https://help.apple.com/xcode/mac/current/en.lproj/dev2539d985f.html),
[alternative distribution notarization](https://developer.apple.com/help/app-store-connect/managing-alternative-distribution/submit-for-notarization/).

## First App Store Connect app

These are one-time account actions. Do them before the first iPhone workflow.

1. Ensure Apple Developer Program membership and current agreements are active.
2. In Certificates, Identifiers & Profiles, register an explicit App ID using
   a reverse-DNS identifier you control. Use the registered identifier in the
   configuration below and the matching distribution profile.
   This iOS identifier does not change the existing Android application ID.
3. In App Store Connect → Apps → **+ → New App**, select iOS. Use app name
   **MusiCalc** if available, English as the primary language, the registered
   bundle ID, and a unique SKU such as **MUSICALC-IOS**.
4. Open App Information and record the app's numeric **Apple ID**. This is the
   app record's ID, not your Apple account email or Team ID.
5. Create an **Apple Distribution** certificate with a private key on this Mac,
   or import an existing certificate/private-key pair authorized for the team.
   Developer ID Application certificates cannot replace iOS distribution signing.
6. Create/download an **App Store Connect distribution provisioning profile**
   for this explicit App ID and certificate. Install it for the runner account.
   Record its profile **name**, not only its downloaded file path.
7. For upload-enabled runs, create an App Store Connect team API key with
   Developer access and appropriate app access. Record Key ID and Issuer ID; download its `AuthKey_<KEY_ID>.p8`
   once and store it securely on the runner. Do not paste its contents into chat.

Apple guides: [create an app record](https://developer.apple.com/help/app-store-connect/create-an-app-record/add-a-new-app/),
[distribution profile](https://developer.apple.com/help/account/provisioning-profiles/create-an-app-store-provisioning-profile),
[upload builds](https://developer.apple.com/help/app-store-connect/manage-builds/upload-builds/).

## Runner-local iPhone configuration

The packager reads `$HOME/.config/musicalc-ios/config.json` using the
**self-hosted runner account's home directory**. A configuration in another
Mac user's home or inside the repository checkout will not be found.

Run the following in a terminal logged in as the runner account. The first two
commands identify the account and exact destination:

~~~bash
whoami
printf 'iOS configuration: %s/.config/musicalc-ios/config.json\n' "$HOME"
mkdir -p "$HOME/.config/musicalc-ios/private_keys"
chmod 700 "$HOME/.config/musicalc-ios" "$HOME/.config/musicalc-ios/private_keys"
nano "$HOME/.config/musicalc-ios/config.json"
chmod 600 "$HOME/.config/musicalc-ios/config.json"
~~~

Use the JSON below, replacing every example value. Build-only runs require
`bundle_id`, `team_id`, `certificate` and `profile`. The remaining fields are
required only when uploading; omit them for a build-only runner:

~~~json
{
  "bundle_id": "io.example.musicalc",
  "team_id": "ABCDEFGHIJ",
  "certificate": "Apple Distribution: Your Team Name (ABCDEFGHIJ)",
  "profile": "MusiCalc App Store",
  "app_id": "1234567890",
  "api_key_id": "YOURKEYID",
  "api_issuer_id": "YOUR-ISSUER-UUID",
  "key_directory": "~/.config/musicalc-ios/private_keys"
}
~~~

For uploads, store `AuthKey_<KEY_ID>.p8` in the configured key directory, owned
by the runner and mode 600. Restrict the configuration/key directories to mode 700.
The API key authorizes destination/build validation and upload. No API token or key content is written to logs. For accounts with multiple providers,
also set `asc_public_id` to the selected provider's public ID as reported by
`xcrun altool --list-providers`.

The iPhone job uses the existing signing-keychain password file to unlock the
installed Apple Distribution private key. No new GitHub credential secret names
are required.

## Diagnosing the App Store Connect key

For upload-enabled runs, the key preflight reports the expanded absolute filename
and the specific failure:
missing/non-regular file, inaccessible metadata, owner UID mismatch, mode mismatch,
or both owner and mode. The existing owner and exact `0600` requirements still apply.
This is the App Store Connect API key, separate from the signing-keychain identity.

In a terminal at the public checkout root, logged in as the **runner account**,
inspect the exact file selected by config.json:

~~~bash
id
asc_key="$(python3 -c 'import json,pathlib; c=json.loads((pathlib.Path.home()/".config/musicalc-ios/config.json").read_text()); print((pathlib.Path(c["key_directory"]).expanduser()/ ("AuthKey_"+c["api_key_id"]+".p8")).absolute())')"
ls -lnL "$asc_key"
~~~

Only the path and ownership/permission metadata are printed. If the file is missing,
copy the downloaded matching `AuthKey_<KEY_ID>.p8` there or correct `key_directory`
and `api_key_id`. If the existing file has the wrong owner or mode, the following
repairs that file for the account currently logged in:

~~~bash
sudo chown "$(id -u):$(id -g)" "$asc_key" && chmod 600 "$asc_key"
ls -lnL "$asc_key"
~~~

The owner UID must match `id -u`, with `-rw-------` permissions. A failed preflight
stops upload-enabled runs before icon generation, Apple queries or packaging.
Build-only runs skip API-key checks and App Store Connect requests entirely.

## Automatic build numbers

Each iPhone job generates a numeric UTC timestamp, `YYYYMMDDHHMMSS`, for
`CFBundleVersion`; for example, `20261102140133`. The marketing version, such as
`0.9.0`, still comes from the selected branch/tag. Both workflow entry points use
the same helper. A retry uses current time, including when rebuilding an older
commit. New marketing versions need no seed or configuration change.

Keep the whole iPhone job serialized through its non-cancelling host group, and
keep the Mac's clock synchronized. The helper pauses one second before taking
the timestamp so even an immediately failed attempt and its successor use distinct
seconds. All producers for this app must use this serialized lane. Other platform
workflows retain their independent scheduling and numbering.

Build-only packaging requires no App Store API key and makes no App Store Connect
requests. When upload is enabled, the packager verifies the configured app/bundle in App Store Connect
and checks that the generated number is newer than every existing build for the
selected marketing version before packaging. Existing one-to-three-component
build strings are compared numerically. A non-increasing timestamp stops with a
clock diagnostic; malformed/API/authentication failures stop the upload-enabled run.
Signing, provisioning and final IPA metadata validation still apply to every run.

No journal, counter lock, seed, backup or bundle migration is involved. The JSON
sidecar beside a completed IPA is artifact metadata, not a mutable counter.

References: [Apple build-version format](https://developer.apple.com/documentation/bundleresources/information-property-list/cfbundleversion),
[distribution version requirements](https://developer.apple.com/documentation/xcode/preparing-your-app-for-distribution),
[App Store build query](https://developer.apple.com/documentation/appstoreconnectapi/get-v1-builds).

## Removing the old counter files

The packager never reads or writes `build-numbers.json`, `build-numbers.lock` or
old counter backups. Missing, corrupt or incorrectly labelled old files cannot
block a new build. Do not run the former seed or bundle-migration commands.

Once no jobs or selected refs are using the old allocator, the runner account
can remove the obsolete journal and lock:

~~~bash
rm -f "$HOME/.config/musicalc-ios/build-numbers.json" "$HOME/.config/musicalc-ios/build-numbers.lock"
~~~

Keep `config.json`, `private_keys/` and signing-keychain files. Old counter backups
are also obsolete and can be removed separately. No cleanup is required to build.

## First internal TestFlight run

1. Manually run **16 Build - iPhone** against the desired branch or tag with upload disabled.
   Verify the signed IPA artifact, bundle ID, version/build and workflow summary.
2. Run again with **Upload to App Store Connect** enabled. This generates a fresh
   timestamp, validates the IPA, and uploads it using `altool --upload-package`.
3. In App Store Connect → MusiCalc → TestFlight, wait for processing. Resolve any
   export-compliance questions using the app's actual encryption behavior.
4. Under **Internal Testing**, create a group such as **MusiCalc Internal**.
   Add your eligible App Store Connect user; add the processed build and fill
   in What to Test. Initially keep distribution internal.
5. Install Apple's **TestFlight** app on an iPhone, accept the invitation, and
   install MusiCalc. Check all calculators, keyboard/input behavior, scrolling,
   safe areas, help, preset import/export and the About dialog when available.
   Mobile test-tone playback currently remains a stub.

A signed IPA, an accepted upload, completed Apple processing, and an installed
test build are separate verification milestones. Internal testers must be
eligible App Store Connect users. Builds expire after 90 days.
[Apple's internal tester instructions](https://developer.apple.com/help/app-store-connect/test-a-beta-version/add-internal-testers/)

## Packaging compatibility and privacy

The pinned Fyne 1.7.3 source is copied into a temporary build directory and patched
for consistent iOS 15 deployment targets, iPhone device family, and preservation
of Xcode's concrete distribution entitlements. The module cache is never edited.
The metadata-injection patch returns an error on failure and removes the import
that became unused when replacing the upstream logging call.

Validate changes to this patch by compiling the complete helper from the public
checkout root, without signing or uploading:

~~~bash
python3 build/scripts/prepare-ios-tool.py
~~~

This verifies the pinned source, dry-runs/applies the patch and runs the Go compiler.

An opaque 1024px iOS icon is generated from the existing artwork on a white
background with Core Graphics and ImageIO; other platform assets remain unchanged.
The drawing buffer is supported 32-bit RGBX; the saved PNG is verified as 1024px
RGB without alpha. Native icon tests run before signing or timestamp generation.

The app bundle includes `PrivacyInfo.xcprivacy` at its root before final signing.
The declarations cover monotonic in-app timing and metadata of the app's own
files/user-selected documents. Preferences use a sandbox JSON file, not
NSUserDefaults. Reassess the manifest whenever platform/library API usage changes.

No workflow submits a public App Store release or automatically invites external
testers.
