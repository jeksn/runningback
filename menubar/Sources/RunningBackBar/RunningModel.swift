import AppKit
import Foundation
import SwiftUI

struct Proc: Codable, Identifiable, Hashable {
    let pid: Int
    let pids: [Int]?
    let command: String
    let label: String
    let cwd: String
    let projectDir: String?
    let role: String?
    let project: String
    let kind: String
    let ports: [Int]
    let url: String?

    var id: Int { pid }

    var portsText: String {
        ports.map(String.init).joined(separator: ",")
    }

    var menuTitle: String {
        let port = ports.first.map { " \u{2014} \($0)" } ?? ""
        return project + port
    }

    // "vite" or the first words of the command line when unrecognized.
    var actionName: String {
        if let role = role, !role.isEmpty {
            return role
        }
        return String(label.prefix(24))
    }
}

@MainActor
final class RunningModel: ObservableObject {
    @Published var items: [Proc] = []
    @Published var lastError: String?

    private var started = false

    private static func binaryPath() -> String? {
        if let custom = ProcessInfo.processInfo.environment["RB_BIN"], !custom.isEmpty {
            return custom
        }
        let candidates = [
            "/opt/homebrew/bin/runningback",
            "/usr/local/bin/runningback",
            NSHomeDirectory() + "/.local/bin/runningback"
        ]
        return candidates.first { FileManager.default.isExecutableFile(atPath: $0) }
    }

    func start() {
        guard !started else { return }
        started = true
        refresh()
        Task { [weak self] in
            while !Task.isCancelled {
                try? await Task.sleep(nanoseconds: 5_000_000_000)
                self?.refresh()
            }
        }
    }

    func refresh() {
        guard let bin = Self.binaryPath() else {
            lastError = "runningback binary not found - run: make install"
            items = []
            return
        }
        do {
            let data = try runProcess(bin, ["list", "--json"])
            items = try JSONDecoder().decode([Proc].self, from: data)
            lastError = nil
        } catch {
            lastError = "error: \(error.localizedDescription)"
        }
    }

    func open(_ proc: Proc) {
        guard let url = proc.url, let u = URL(string: url) else { return }
        NSWorkspace.shared.open(u)
    }

    func copy(_ text: String) {
        NSPasteboard.general.clearContents()
        NSPasteboard.general.setString(text, forType: .string)
    }

    func kill(_ proc: Proc) {
        guard let bin = Self.binaryPath() else { return }
        for pid in proc.pids ?? [proc.pid] {
            let p = Process()
            p.executableURL = URL(fileURLWithPath: bin)
            p.arguments = ["kill", String(pid)]
            try? p.run()
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.4) { [weak self] in
            self?.refresh()
        }
    }

    private func runProcess(_ path: String, _ arguments: [String]) throws -> Data {
        let p = Process()
        p.executableURL = URL(fileURLWithPath: path)
        p.arguments = arguments
        let pipe = Pipe()
        p.standardOutput = pipe
        p.standardError = FileHandle.nullDevice
        try p.run()
        let data = pipe.fileHandleForReading.readDataToEndOfFile()
        p.waitUntilExit()
        return data
    }
}
