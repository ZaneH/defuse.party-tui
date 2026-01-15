module github.com/ZaneH/defuse.party-tui

go 1.23.4

require (
	github.com/charmbracelet/bubbletea v1.3.1
	github.com/charmbracelet/lipgloss v1.0.0
	github.com/charmbracelet/ssh v0.0.0-20240202115812-f4ab1009799a
	github.com/charmbracelet/wish v1.3.1
	github.com/muesli/termenv v0.15.2
	google.golang.org/grpc v1.72.0
)

require (
	github.com/anmitsu/go-shlex v0.0.0-20200514113438-38f4b401e2be // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/keygen v0.5.0 // indirect
	github.com/charmbracelet/log v0.3.1 // indirect
	github.com/charmbracelet/x/ansi v0.8.0 // indirect
	github.com/charmbracelet/x/errors v0.0.0-20240117030013-d31dba354651 // indirect
	github.com/charmbracelet/x/exp/term v0.0.0-20240202113029-6ff29cf0473e // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/creack/pty v1.1.21 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

// Development: Uncomment to use local backend (requires ../keep-talking to exist)
// This allows instant feedback when changing proto definitions locally.
// DO NOT commit with this uncommented - it breaks production builds.
// replace github.com/ZaneH/defuse.party-go => ../keep-talking
