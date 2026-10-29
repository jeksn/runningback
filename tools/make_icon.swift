// Generates build/AppIcon.icns from the lucide-lab football helmet
// (https://github.com/lucide-icons/lucide-lab, ISC license), stroked in
// two tones. Run from the repo root: swift tools/make_icon.swift
// The SVG path parser mirrors HelmetIcon.swift in the app source.
import CoreGraphics
import Foundation
import ImageIO

// Minimal SVG path parser: M/m, L/l, H/h, V/v, C/c, Z/z with implicit
// repetition. Paths stay in SVG y-down space; transform when drawing.
struct SVGPath {
    let cgPath: CGPath

    private enum Token {
        case cmd(Character)
        case num(Double)
    }

    init(d: String) {
        let chars = Array(d)
        var tokens: [Token] = []
        var i = 0
        while i < chars.count {
            let c = chars[i]
            if c.isLetter {
                tokens.append(.cmd(c))
                i += 1
            } else if c.isNumber || c == "." || c == "-" {
                var j = i
                if chars[j] == "-" { j += 1 }
                var dot = false
                while j < chars.count {
                    if chars[j].isNumber {
                        j += 1
                    } else if chars[j] == "." && !dot {
                        dot = true
                        j += 1
                    } else {
                        break
                    }
                }
                tokens.append(.num(Double(String(chars[i..<j])) ?? 0))
                i = j
            } else {
                i += 1
            }
        }

        let path = CGMutablePath()
        var idx = 0
        var cmd: Character = " "
        var cur = CGPoint.zero
        var sub = CGPoint.zero
        var firstMove = true

        func num() -> Double {
            guard idx < tokens.count, case .num(let n) = tokens[idx] else { return 0 }
            idx += 1
            return n
        }

        while idx < tokens.count {
            if case .cmd(let c) = tokens[idx] {
                cmd = c
                idx += 1
                if c == "z" || c == "Z" {
                    path.closeSubpath()
                    cur = sub
                    continue
                }
            }
            switch cmd {
            case "M", "m":
                let rel = cmd == "m"
                let x = num(), y = num()
                let p = rel ? CGPoint(x: cur.x + x, y: cur.y + y) : CGPoint(x: x, y: y)
                if firstMove {
                    path.move(to: p)
                    firstMove = false
                } else {
                    path.addLine(to: p)
                }
                cur = p
                sub = p
                cmd = rel ? "l" : "L"
            case "L", "l":
                let x = num(), y = num()
                let p = cmd == "L" ? CGPoint(x: x, y: y) : CGPoint(x: cur.x + x, y: cur.y + y)
                path.addLine(to: p)
                cur = p
            case "H", "h":
                let x = num()
                let p = CGPoint(x: cmd == "H" ? x : cur.x + x, y: cur.y)
                path.addLine(to: p)
                cur = p
            case "V", "v":
                let y = num()
                let p = CGPoint(x: cur.x, y: cmd == "V" ? y : cur.y + y)
                path.addLine(to: p)
                cur = p
            case "C", "c":
                let rel = cmd == "c"
                var n = [Double](repeating: 0, count: 6)
                for k in 0..<6 { n[k] = num() }
                let c1 = rel ? CGPoint(x: cur.x + n[0], y: cur.y + n[1]) : CGPoint(x: n[0], y: n[1])
                let c2 = rel ? CGPoint(x: cur.x + n[2], y: cur.y + n[3]) : CGPoint(x: n[2], y: n[3])
                let e = rel ? CGPoint(x: cur.x + n[4], y: cur.y + n[5]) : CGPoint(x: n[4], y: n[5])
                path.addCurve(to: e, control1: c1, control2: c2)
                cur = e
            case "A", "a":
                // Circular arcs (rx == ry, x-axis rotation ignored),
                // approximated with a single cubic - exact for the
                // quarter-circle corners lucide icons use.
                let rel = cmd == "a"
                var n = [Double](repeating: 0, count: 7)
                for k in 0..<7 { n[k] = num() }
                let r = max(n[0], n[1])
                let e = rel ? CGPoint(x: cur.x + n[5], y: cur.y + n[6]) : CGPoint(x: n[5], y: n[6])
                let dx = e.x - cur.x, dy = e.y - cur.y
                let d = (dx * dx + dy * dy).squareRoot()
                if d == 0 || r == 0 {
                    path.addLine(to: e)
                    cur = e
                    break
                }
                let rr = max(r, d / 2)
                let mid = CGPoint(x: (cur.x + e.x) / 2, y: (cur.y + e.y) / 2)
                let h = (rr * rr - (d / 2) * (d / 2)).squareRoot()
                let ux = dx / d, uy = dy / d
                let s: Double = n[4] > 0.5 ? 1 : -1
                let c = CGPoint(x: mid.x - uy * s * h, y: mid.y + ux * s * h)
                let a0 = CoreGraphics.atan2(cur.y - c.y, cur.x - c.x)
                let a1 = CoreGraphics.atan2(e.y - c.y, e.x - c.x)
                var delta = a1 - a0
                if s > 0 && delta < 0 { delta += 2 * CGFloat.pi }
                if s < 0 && delta > 0 { delta -= 2 * CGFloat.pi }
                let k = 4 / 3 * CoreGraphics.tan(delta / 4)
                let c1 = CGPoint(x: cur.x + k * rr * -CoreGraphics.sin(a0),
                                 y: cur.y + k * rr * CoreGraphics.cos(a0))
                let c2 = CGPoint(x: e.x - k * rr * -CoreGraphics.sin(a1),
                                 y: e.y - k * rr * CoreGraphics.cos(a1))
                path.addCurve(to: e, control1: c1, control2: c2)
                cur = e
            default:
                idx += 1
            }
        }
        cgPath = path
    }
}

let shellD = "M21.6 9c-1.3-4-5.1-7-9.6-7C6.5 2 2 6.5 2 12c0 2.6 1 5 3 7 1.4 1.3 3.6 1.4 4.9 0 .7-.7 1-1.6 1-2.5V13c0-1.7 1.3-3 3-3h6.8c.7 0 1-.4.9-1z"
let earD = "M7 14h.01"
let barsD = "M22 18H10.7 M11 14h9a2 2 0 0 1 2 2v4a2 2 0 0 1-2 2c-2.8 0-5-2.2-5-5v-3"

let shellColor = CGColor(srgbRed: 0.12, green: 0.14, blue: 0.20, alpha: 1)
let maskColor = CGColor(srgbRed: 0.78, green: 0.80, blue: 0.84, alpha: 1)

func writePNG(pixels: Int, to path: String) {
    let space = CGColorSpace(name: CGColorSpace.sRGB)!
    let ctx = CGContext(data: nil, width: pixels, height: pixels,
                        bitsPerComponent: 8, bytesPerRow: 0, space: space,
                        bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue)!

    // Inset the 24x24 design so the macOS rounded-rect mask does not
    // clip the strokes. Raw CGContexts export row 0 at the top of the
    // image, inverting y-up drawing - so map SVG y-down space directly.
    let margin = CGFloat(pixels) * 0.11
    let scale = CGFloat(pixels) * 0.78 / 24
    let t = CGAffineTransform(a: scale, b: 0, c: 0, d: scale,
                              tx: margin, ty: margin)
    ctx.setLineWidth(2 * scale)
    ctx.setLineCap(.round)
    ctx.setLineJoin(.round)

    for (d, color) in [(shellD, shellColor), (earD, shellColor), (barsD, maskColor)] {
        let mapped = CGMutablePath()
        mapped.addPath(SVGPath(d: d).cgPath, transform: t)
        ctx.setStrokeColor(color)
        ctx.addPath(mapped)
        ctx.strokePath()
    }

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
