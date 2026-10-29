// Generates build/AppIcon.icns from a vector football helmet.
// Run from the repo root: swift tools/make_icon.swift
import CoreGraphics
import Foundation
import ImageIO

let shellColor = CGColor(srgbRed: 0.12, green: 0.14, blue: 0.20, alpha: 1)
let maskColor = CGColor(srgbRed: 0.78, green: 0.80, blue: 0.84, alpha: 1)
let stripeColor = CGColor(srgbRed: 0.90, green: 0.28, blue: 0.30, alpha: 1)
let earColor = CGColor(srgbRed: 0.30, green: 0.33, blue: 0.41, alpha: 1)

// Geometry on an 18x18 canvas, helmet facing right; mirrors the
// template icon drawn in HelmetIcon.swift.
let shellRect = CGRect(x: 1, y: 2.6, width: 13.2, height: 13.2)
let openingRect = CGRect(x: 10.2, y: 4.4, width: 5, height: 8)
let earRect = CGRect(x: 6.5, y: 7.5, width: 2.2, height: 2.2)
let barRects = [CGRect(x: 10, y: 5.2, width: 6, height: 1.5),
                CGRect(x: 10, y: 8.6, width: 6, height: 1.5),
                CGRect(x: 14.5, y: 5.2, width: 1.5, height: 4.9)]

func drawHelmet(_ ctx: CGContext) {
    // Shell with the face opening carved out.
    ctx.setFillColor(shellColor)
    ctx.addEllipse(in: shellRect)
    ctx.addPath(CGPath(roundedRect: openingRect, cornerWidth: 2, cornerHeight: 2, transform: nil))
    ctx.fillPath(using: .evenOdd)

    // Center stripe over the crown, clipped to the shell.
    ctx.saveGState()
    ctx.addEllipse(in: shellRect)
    ctx.clip()
    ctx.setFillColor(stripeColor)
    ctx.fill(CGRect(x: 5.6, y: 11, width: 2.2, height: 6))
    ctx.restoreGState()

    // Ear pad.
    ctx.setFillColor(earColor)
    ctx.fillEllipse(in: earRect)

    // Facemask bars.
    ctx.setFillColor(maskColor)
    for rect in barRects {
        ctx.addPath(CGPath(roundedRect: rect,
                           cornerWidth: rect.height / 2, cornerHeight: rect.height / 2,
                           transform: nil))
        ctx.fillPath()
    }
}

func writePNG(pixels: Int, to path: String) {
    let space = CGColorSpace(name: CGColorSpace.sRGB)!
    let ctx = CGContext(data: nil, width: pixels, height: pixels,
                        bitsPerComponent: 8, bytesPerRow: 0, space: space,
                        bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue)!
    ctx.scaleBy(x: CGFloat(pixels) / 18, y: CGFloat(pixels) / 18)
    // Raw CGContexts render row 0 at the top of the image, which flips
    // y-up drawing; compensate so the design appears as drawn.
    ctx.translateBy(x: 0, y: 18)
    ctx.scaleBy(x: 1, y: -1)
    drawHelmet(ctx)
    let image = ctx.makeImage()!
    let dest = CGImageDestinationCreateWithURL(URL(fileURLWithPath: path) as CFURL,
                                               "public.png" as CFString, 1, nil)!
    CGImageDestinationAddImage(dest, image, nil)
    CGImageDestinationFinalize(dest)
}

let fm = FileManager.default
try? fm.createDirectory(atPath: "build", withIntermediateDirectories: true)
let iconset = "build/AppIcon.iconset"
try? fm.removeItem(atPath: iconset)
try! fm.createDirectory(atPath: iconset, withIntermediateDirectories: true)

let sizes: [(name: String, pixels: Int)] = [
    ("icon_16x16.png", 16), ("icon_16x16@2x.png", 32),
    ("icon_32x32.png", 32), ("icon_32x32@2x.png", 64),
    ("icon_128x128.png", 128), ("icon_128x128@2x.png", 256),
    ("icon_256x256.png", 256), ("icon_256x256@2x.png", 512),
    ("icon_512x512.png", 512), ("icon_512x512@2x.png", 1024),
]
for spec in sizes {
    writePNG(pixels: spec.pixels, to: "\(iconset)/\(spec.name)")
}

let iconutil = Process()
iconutil.executableURL = URL(fileURLWithPath: "/usr/bin/iconutil")
iconutil.arguments = ["-c", "icns", iconset, "-o", "build/AppIcon.icns"]
try! iconutil.run()
iconutil.waitUntilExit()
if iconutil.terminationStatus != 0 {
    fatalError("iconutil failed")
}
print("wrote build/AppIcon.icns")
