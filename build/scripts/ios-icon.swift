import CoreGraphics
import Foundation
import ImageIO
import UniformTypeIdentifiers

func fail(_ message: String) -> Never {
    FileHandle.standardError.write(Data((message + "\n").utf8))
    exit(1)
}

guard CommandLine.arguments.count == 3 else {
    fail("Usage: ios-icon.swift source.png destination.png")
}
let sourceURL = URL(fileURLWithPath: CommandLine.arguments[1])
let destinationURL = URL(fileURLWithPath: CommandLine.arguments[2])
guard let source = CGImageSourceCreateWithURL(sourceURL as CFURL, nil),
      let artwork = CGImageSourceCreateImageAtIndex(source, 0, nil) else {
    fail("Cannot decode source iOS icon")
}

// Quartz draws RGB into a 32-bit buffer. The skipped byte is not alpha,
// so ImageIO encodes an opaque RGB PNG after compositing onto white.
guard let context = CGContext(data: nil, width: 1024, height: 1024,
                              bitsPerComponent: 8, bytesPerRow: 1024 * 4,
                              space: CGColorSpaceCreateDeviceRGB(),
                              bitmapInfo: CGImageAlphaInfo.noneSkipLast.rawValue) else {
    fail("Cannot create RGBX iOS icon context")
}
let bounds = CGRect(x: 0, y: 0, width: 1024, height: 1024)
context.setFillColor(CGColor(red: 1, green: 1, blue: 1, alpha: 1))
context.fill(bounds)
context.interpolationQuality = .high
context.draw(artwork, in: bounds)
guard let image = context.makeImage(),
      let destination = CGImageDestinationCreateWithURL(
        destinationURL as CFURL, UTType.png.identifier as CFString, 1, nil) else {
    fail("Cannot create iOS icon PNG")
}
CGImageDestinationAddImage(destination, image, nil)
guard CGImageDestinationFinalize(destination) else {
    fail("Cannot write iOS icon PNG")
}
guard let saved = CGImageSourceCreateWithURL(destinationURL as CFURL, nil),
      let verified = CGImageSourceCreateImageAtIndex(saved, 0, nil),
      verified.width == 1024, verified.height == 1024,
      [.none, .noneSkipFirst, .noneSkipLast].contains(verified.alphaInfo) else {
    fail("iOS icon must be opaque and 1024 x 1024")
}
