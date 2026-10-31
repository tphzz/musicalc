# Apple build runner and TestFlight setup

Use the runner account on the Apple Silicon Mac. Workflows select
`/Applications/Xcode.app/Contents/Developer`. macOS packages support **13.3+**.
iPhone builds target **iOS 15+** and require **Xcode 26+ / iOS 26 SDK+** for current
App Store Connect uploads.

## Repository checkout for manual commands

Run the commands in this guide in a macOS terminal, using the same account that
runs the self-hosted runner. Relative paths such as `build/scripts/...` start at
the **public MusiCalc repository root**.

If this account does not already have a checkout, create one in a directory of
your choice, then enter it:

~~~bash
git clone --branch 0.9.0 https://github.com/tphzz/musicalc.git
cd musicalc
~~~

If a checkout already exists, enter its root and ensure it contains the
`0.9.0` build scripts. The helpers are tracked in the repository:
[ios_build_number.py](../scripts/ios_build_number.py) and its imported
[version.py](../scripts/version.py) stay together in `build/scripts/`.

The persistent runner files live separately under `$HOME/.config/musicalc-ios/`:
`config.json`, the API key directory and `build-numbers.json`. The seed command
below creates/updates that counter JSON through `--file`; it runs the Python
helper directly from the repository. GitHub Actions checks out the build scripts
automatically when a workflow runs.

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
2. In Certificates, Identifiers & Profiles, register an explicit App ID for
   **com.musicalc** if available. If unavailable, choose a replacement before
   creating certificates/profiles; use it consistently in the configuration below.
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

Use the JSON below, replacing every example value. For packaging with upload
disabled, `bundle_id`, `team_id`, `certificate` and `profile` are required.
The app/API-key fields are required when enabling upload:

~~~json
{
  "bundle_id": "com.musicalc",
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
key is used only when upload is enabled. For accounts with multiple providers,
also set `asc_public_id` to the selected provider's public ID as reported by
`xcrun altool --list-providers`.

The iPhone job uses the existing signing-keychain password file to unlock the
installed Apple Distribution private key. No new GitHub credential secret names
are required.

## Seed and preserve build numbers

From that repository root, as the runner account, explicitly seed a brand-new
app/version at zero:

~~~bash
python3 build/scripts/ios_build_number.py \
  --file "$HOME/.config/musicalc-ios/build-numbers.json" \
  --bundle com.musicalc --version 0.9.0 --seed 0
~~~

Seeding at zero records the starting point; the first workflow build reserves
number **1**. For example, seeding an existing version at 12 makes its next
reservation **13**.

The `--bundle` value must exactly match `bundle_id` in the runner's
`config.json`. Replace the example bundle consistently in both setup commands
and configuration. The counter stores that identity in its `bundle` field;
an identity mismatch is rejected before reserving a number. Preserve the counter
and its version high-water marks when reconciling an incorrect setup value.
For a counter belonging to a different app, keep it separately and initialize
the intended app's counter using reconciled App Store Connect/reservation values.

For an existing version, use the highest number already present in App Store
Connect, or any higher locally reserved number, instead of zero. Seed each new
marketing version explicitly. Back up the counter along with the runner setup.
If state is lost, reconcile with App Store Connect before building again.

Aggregate runs, platform runs and retries reserve from this same locked counter.
A reservation is persisted before packaging and is never rolled back. Gaps after
failures are intentional. Values are 1–9999; exhaustion fails rather than wraps.
Keep the whole iPhone job serialized and non-cancelling through its host group.

## First internal TestFlight run

1. Manually run **16 Build - iPhone** against branch `0.9.0` with upload disabled.
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
background with the native AppKit tool; other platform assets remain unchanged.

The app bundle includes `PrivacyInfo.xcprivacy` at its root before final signing.
The declarations cover monotonic in-app timing and metadata of the app's own
files/user-selected documents. Preferences use a sandbox JSON file, not
NSUserDefaults. Reassess the manifest whenever platform/library API usage changes.

No workflow submits a public App Store release or automatically invites external
testers.
