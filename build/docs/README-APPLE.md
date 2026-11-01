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
configuration. The build-number helper contains no fixed product bundle ID.

Only manual development/debugging commands such as `build/scripts/build-ios.py`
need a terminal at a public MusiCalc checkout's root. Runner setup commands below
use absolute home-directory paths and do not require a checkout.

Persistent runner files live outside checkout under
`$HOME/.config/musicalc-ios/`: configuration, API keys and the automatically
managed build-number journal. GitHub Actions replaces source checkouts; it
must retain that directory.

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
7. Create an App Store Connect team API key with Developer access and appropriate
   app access. Record Key ID and Issuer ID; download its `AuthKey_<KEY_ID>.p8`
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

Use the JSON below, replacing every example value. All fields are required for
both build-only and upload runs. App Store Connect read access is used to assign
build numbers even when uploading is disabled:

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

Store `AuthKey_<KEY_ID>.p8` in the configured key directory, owned by the runner
and mode 600. Restrict the configuration/key directories to mode 700. The API
key authorizes read-only app/build queries for every build and upload when
enabled. No API token or key content is written to logs. For accounts with multiple providers,
also set `asc_public_id` to the selected provider's public ID as reported by
`xcrun altool --list-providers`.

The iPhone job uses the existing signing-keychain password file to unlock the
installed Apple Distribution private key. No new GitHub credential secret names
are required.

## Automatic build numbers

Every iPhone run automatically:

1. Takes the marketing version from the selected Git ref and the bundle ID from
   `config.json`.
2. Verifies that the configured numeric App Store Connect app ID belongs to that
   bundle, then reads all pages of iOS builds for that marketing version.
3. Reserves `max(highest Apple build, highest local reservation) + 1` under the
   local file lock, persisting the number before packaging.

The first successful lookup for an unused version returns zero, so its first
build reserves **1**. A new marketing version and a missing journal are
initialized automatically from Apple's records. There is no per-version seed
command, separate clone or manually edited version list.

`build-numbers.json` is generated runtime bookkeeping, not source configuration.
It remembers numbers consumed by failed or unuploaded builds, which Apple
cannot know about. Aggregate runs, standalone runs and retries share it; a failed
packaging run does not return its reservation. Keep the complete iPhone job
serialized through its non-cancelling host group. All producers for this app
must share the same reservation journal/host lane.

The API key and app access must work even when **Upload to App Store Connect**
is disabled. Authentication, network, app/bundle mismatch, malformed responses
and exhausted counters stop the build rather than assuming zero. Existing
bundle mismatches or corrupt local state are not silently rewritten.

Back up the journal with the runner setup. If it is lost, restore the newest
backup and reconcile any higher locally reserved numbers before restarting
jobs: Apple can recover uploaded history, but cannot recover unpublished local
reservations. The helper's explicit `--seed` option remains a recovery tool for
a reconciled high-water mark; normal builds never need it. Build values are
**1–9999** for each marketing version; exhaustion fails instead of wrapping.

For example, a reservation of 1 followed by an icon/packaging failure remains
consumed; the next packaging attempt reserves at least 2. The workflow's native
helper tests and actual icon preflight now run before reserving a number.

Apple API references: [list builds](https://developer.apple.com/documentation/appstoreconnectapi/get-v1-builds),
[team JWT authentication](https://developer.apple.com/documentation/appstoreconnectapi/generating-tokens-for-api-requests).

## First internal TestFlight run

1. Manually run **16 Build - iPhone** against the desired branch or tag with upload disabled.
   Verify the signed IPA artifact, bundle ID, version/build and workflow summary.
2. Run again with **Upload to App Store Connect** enabled. This reserves a fresh
   number, validates the IPA, and uploads it using `altool --upload-package`.
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

An opaque 1024px iOS icon is generated from the existing artwork on a white
background with Core Graphics and ImageIO; other platform assets remain unchanged.
The drawing buffer is supported 32-bit RGBX; the saved PNG is verified as 1024px
RGB without alpha. Native icon tests run before signing or number reservation.

The app bundle includes `PrivacyInfo.xcprivacy` at its root before final signing.
The declarations cover monotonic in-app timing and metadata of the app's own
files/user-selected documents. Preferences use a sandbox JSON file, not
NSUserDefaults. Reassess the manifest whenever platform/library API usage changes.

No workflow submits a public App Store release or automatically invites external
testers.
