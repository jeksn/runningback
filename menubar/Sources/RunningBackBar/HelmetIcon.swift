import AppKit

// Draws a side-view football helmet. Used for the menu bar icon
// (monochrome template) and for the app icon (colored, via
// tools/make_icon.swift, which repeats this geometry).
enum HelmetIcon {
    static let canvas: CGFloat = 18

    // Shared geometry on an 18x18 canvas, helmet facing right.
    static let shellRect = NSRect(x: 1, y: 2.6, width: 13.2, height: 13.2)
    static let faceOpening = NSRect(x: 10.2, y: 4.4, width: 5, height: 8)
    static let earCircle = NSRect(x: 6.5, y: 7.5, width: 2.2, height: 2.2)
    static let bar1 = NSRect(x: 10, y: 5.2, width: 6, height: 1.5)
    static let bar2 = NSRect(x: 10, y: 8.6, width: 6, height: 1.5)
    static let barVertical = NSRect(x: 14.5, y: 5.2, width: 1.5, height: 4.9)

    // Monochrome template image for the menu bar: macOS recolors it to
    // match light/dark mode automatically.
    static func menuBarImage() -> NSImage {
        let image = NSImage(size: NSSize(width: canvas, height: canvas))
        image.lockFocus()
        NSColor.black.setFill()
        drawShell()
        drawMask()
        image.unlockFocus()
        image.isTemplate = true
        return image
    }

    static func drawShell() {
        let shell = NSBezierPath(ovalIn: shellRect)
        let opening = NSBezierPath(roundedRect: faceOpening, xRadius: 2, yRadius: 2)
        let ear = NSBezierPath(ovalIn: earCircle)
        shell.append(opening)
        shell.append(ear)
        shell.windingRule = .evenOdd
        shell.fill()
    }

    static func drawMask() {
        for rect in [bar1, bar2, barVertical] {
            NSBezierPath(roundedRect: rect, xRadius: rect.height / 2, yRadius: rect.height / 2).fill()
        }
    }
}
