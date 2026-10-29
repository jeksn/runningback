import AppKit
import SwiftUI

@main
struct RunningBackApp: App {
    @StateObject private var model = RunningModel()

    var body: some Scene {
        MenuBarExtra(content: {
            MenuContent()
                .environmentObject(model)
                .onAppear { model.start() }
        }, label: {
            if model.items.isEmpty {
                Image(systemName: "figure.run")
            } else {
                Label(String(model.items.count), systemImage: "figure.run")
            }
        })
        .menuBarExtraStyle(.menu)
    }
}

struct MenuContent: View {
    @EnvironmentObject private var model: RunningModel

    var body: some View {
        let dev = model.items.filter { $0.kind == "dev" }
        let services = model.items.filter { $0.kind != "dev" }

        if model.items.isEmpty {
            Text(model.lastError ?? "No dev servers running")
        }

        ForEach(dev) { proc in
            Menu(proc.menuTitle) {
                if let url = proc.url {
                    Button("Open \(url)") { model.open(proc) }
                    Button("Copy URL") { model.copy(url) }
                    Divider()
                }
                Button("Copy path \(proc.cwd)") { model.copy(proc.cwd) }
                Divider()
                Button("Stop \(proc.label) (pid \(proc.pid))") { model.kill(proc) }
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
        Button("Refresh") { model.refresh() }
        Button("Quit") { NSApplication.shared.terminate(nil) }
    }
}
