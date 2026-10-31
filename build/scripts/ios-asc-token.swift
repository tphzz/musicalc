import CryptoKit
import Foundation

func base64url(_ data: Data) -> String {
    data.base64EncodedString()
        .replacingOccurrences(of: "+", with: "-")
        .replacingOccurrences(of: "/", with: "_")
        .replacingOccurrences(of: "=", with: "")
}

do {
    guard CommandLine.arguments.count == 4 else {
        throw NSError(domain: "MusiCalcASC", code: 1)
    }
    let key = try P256.Signing.PrivateKey(
        pemRepresentation: String(contentsOfFile: CommandLine.arguments[1], encoding: .utf8))
    let now = Int(Date().timeIntervalSince1970)
    let header = try JSONSerialization.data(withJSONObject: [
        "alg": "ES256", "kid": CommandLine.arguments[2], "typ": "JWT"])
    let claims: [String: Any] = [
        "iss": CommandLine.arguments[3], "iat": now - 30,
        "exp": now + 270, "aud": "appstoreconnect-v1"]
    let payload = try JSONSerialization.data(withJSONObject: claims)
    let message = base64url(header) + "." + base64url(payload)
    let signature = try key.signature(for: Data(message.utf8))
    print(message + "." + base64url(signature.rawRepresentation))
} catch {
    FileHandle.standardError.write(Data("Cannot sign App Store Connect token\n".utf8))
    exit(1)
}
