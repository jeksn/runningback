import AppKit
import SwiftUI

@main
struct RunningBackApp: App {
    @StateObject private var model = RunningModel()

    var body: some Scene {
        MenuBarExtra(content: {
            MenuContent()
                .environmentObject(model)
                .onAppear {
                    model.start()
                    model.setMenuOpen(true)
                    model.refresh()
                }
                .onDisappear { model.setMenuOpen(false) }
        }, label: {
            let helmet = HelmetIcon.menuBarImage()
            if model.items.isEmpty {
                Image(nsImage: helmet)
            } else {
                Label { Text(String(model.items.count)) } icon: { Image(nsImage: helmet) }
            }
        })
        .menuBarExtraStyle(.menu)
    }
}

// DevGroup bundles every dev server that belongs to the same project
// checkout, so a Vite server and a php server show as one entry.
struct DevGroup: Identifiable {
    let dir: String
    var procs: [Proc]

    var id: String { dir }
    var name: String { procs.first?.project ?? dir }
    var portsText: String {
        var seen = Set<Int>()
        return procs
            .flatMap { $0.ports }
            .filter { seen.insert($0).inserted }
            .map(String.init)
            .joined(separator: ",")
    }
    var title: String { "\(name) \u{2014} \(portsText)" }
}

func devGroups(_ procs: [Proc]) -> [DevGroup] {
    var out: [DevGroup] = []
    for p in procs {
        let key = p.projectDir ?? p.cwd
        if !out.isEmpty && out[out.count - 1].dir == key {
            out[out.count - 1].procs.append(p)
        } else {
            out.append(DevGroup(dir: key, procs: [p]))
        }
    }
    return out
}

struct MenuContent: View {
    @EnvironmentObject private var model: RunningModel

    var body: some View {
        let dev = model.items.filter { $0.kind == "dev" }
        let services = model.items.filter { $0.kind != "dev" }

        if model.items.isEmpty {
            Text(model.lastError ?? "No dev servers running")
        }

        ForEach(devGroups(dev)) { group in
            Menu(group.title) {
                ForEach(group.procs) { proc in
                    if let url = proc.url {
                        Button("Open \(url) \u{2014} \(proc.actionName)") { model.open(proc) }
                    }
                }
                Divider()
                ForEach(group.procs) { proc in
                    Button("Stop \(proc.actionName) (pid \(proc.pid))") {
                        model.kill(proc)
                    }
                }
                Divider()
                Button("Stop all") { group.procs.forEach { model.kill($0) } }
                Button("Copy path \(group.dir)") { model.copy(group.dir) }
            }
        }

        if !services.isEmpty {
            Divider()
            ForEach(services) { proc in
                Menu("\(proc.project) \u{2014} \(proc.portsText)") {
                    Button("Copy ports") { model.copy(proc.portsText) }
                    Button("Stop (pid \(proc.pid))") { model.kill(proc) }
                }
            }
        }

        Divider()
        Button(model.loginEnabled ? "Launch at Login: On" : "Launch at Login: Off") {
            model.toggleLaunchAtLogin()
        }
        Button("Refresh") { model.refresh() }
        Button("Quit") { NSApplication.shared.terminate(nil) }
    }
}
