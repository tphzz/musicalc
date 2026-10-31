import AppKit

// App Store's flat icon path requires an opaque 1024px PNG. Preserve the
// existing artwork and composite its transparent corners onto white.
guard CommandLine.arguments.count == 3,
      let source = NSImage(contentsOfFile: CommandLine.arguments[1]),
      let bitmap = NSBitmapImageRep(bitmapDataPlanes: nil, pixelsWide: 1024,
                                    pixelsHigh: 1024, bitsPerSample: 8,
                                    samplesPerPixel: 3, hasAlpha: false,
                                    isPlanar: false, colorSpaceName: .deviceRGB,
                                    bytesPerRow: 0, bitsPerPixel: 0),
      let context = NSGraphicsContext(bitmapImageRep: bitmap) else {
    fatalError("Cannot create opaque iOS icon")
}
NSGraphicsContext.saveGraphicsState()
NSGraphicsContext.current = context
context.imageInterpolation = .high
let bounds = NSRect(x: 0, y: 0, width: 1024, height: 1024)
NSColor.white.setFill()
bounds.fill()
source.draw(in: bounds, from: .zero, operation: .sourceOver, fraction: 1)
context.flushGraphics()
NSGraphicsContext.restoreGraphicsState()
guard let data = bitmap.representation(using: .png, properties: [:]) else {
    fatalError("Cannot encode iOS icon")
}
let destination = URL(fileURLWithPath: CommandLine.arguments[2])
try data.write(to: destination, options: .atomic)
guard let verified = NSBitmapImageRep(contentsOf: destination),
      verified.pixelsWide == 1024, verified.pixelsHigh == 1024,
      !verified.hasAlpha else {
    fatalError("iOS icon must be opaque and 1024 x 1024")
}
