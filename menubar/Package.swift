// swift-tools-version:5.10
import PackageDescription

let package = Package(
    name: "RunningBackBar",
    platforms: [.macOS(.v13)],
    targets: [
        .executableTarget(name: "RunningBackBar", path: "Sources/RunningBackBar")
    ]
)
