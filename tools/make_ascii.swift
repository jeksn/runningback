// Generates ASCII art of the lucide-lab football helmet for the TUI
// header. Renders the SVG strokes at high resolution, then maps each
// character cell's coverage onto an ASCII ramp. The output is pasted
// into tui.go as helmetArt.
// Run from the repo root: swift tools/make_ascii.swift [gridW gridH]
import CoreGraphics
import Foundation

// SVG path parser (same subset as HelmetIcon.swift / make_icon.swift).
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

let gridW = CommandLine.arguments.count > 2 ? Int(CommandLine.arguments[1])! : 26
let gridH = CommandLine.arguments.count > 2 ? Int(CommandLine.arguments[2])! : 13

// Rasterize at high resolution.
let R = 8 // pixels per SVG unit
let size = 24 * R
let ctx = CGContext(data: nil, width: size, height: size, bitsPerComponent: 8, bytesPerRow: 0,
                    space: CGColorSpace(name: CGColorSpace.sRGB)!,
                    bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue)!
// SVG y-down maps directly: row 0 of the bitmap is the image top.
let t = CGAffineTransform(a: CGFloat(R), b: 0, c: 0, d: CGFloat(R), tx: 0, ty: 0)
ctx.setLineWidth(CGFloat(2 * R))
ctx.setLineCap(.round)
ctx.setLineJoin(.round)
ctx.setStrokeColor(CGColor(gray: 1, alpha: 1))
let paths: [String]
if CommandLine.arguments.contains("dot") {
    // Debug: single dot at SVG (23,1), the top-right corner.
    paths = ["M23 1h.01"]
} else {
    paths = [shellD, earD, barsD]
}
for d in paths {
    let mapped = CGMutablePath()
    mapped.addPath(SVGPath(d: d).cgPath, transform: t)
    ctx.addPath(mapped)
    ctx.strokePath()
}
let data = ctx.data!.bindMemory(to: UInt8.self, capacity: size * size * 4)

// Downsample to the character grid; each cell keeps its coverage ratio.
let ramp: [Character] = [" ", ".", ":", "-", "=", "+", "*", "#", "%", "@"]
var lines: [String] = []
for gy in 0..<gridH {
    var line = ""
    for gx in 0..<gridW {
        let x0 = gx * size / gridW, x1 = (gx + 1) * size / gridW
        let y0 = gy * size / gridH, y1 = (gy + 1) * size / gridH
        var hit = 0, total = 0
        var y = y0
        while y < y1 {
            var x = x0
            while x < x1 {
                // Bitmap row 0 is the bottom of the SVG design space.
                let i = ((size - 1 - y) * size + x) * 4 + 3
                if data[i] > 128 { hit += 1 }
                total += 1
                x += 1
            }
            y += 1
        }
        let cov = total == 0 ? 0 : Double(hit) / Double(total)
        let idx = min(ramp.count - 1, Int(cov * Double(ramp.count)))
        line.append(ramp[idx])
    }
    lines.append(line)
}

let out = lines.joined(separator: "\n")
print(out)
