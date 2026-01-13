# KTANE TUI Implementation Plan

## Overview

Build a Terminal User Interface (TUI) for "Keep Talking and Nobody Explodes" using:
- **Go + Bubbletea** for the TUI framework
- **Charmbracelet Wish** for SSH server delivery
- **Existing gRPC backend** (`keep-talking/`) for game logic

The TUI will be a **new service** that acts as a gRPC client to the existing backend.

---

## Project Structure

```
keep-talking-tui/
├── cmd/
│   └── server/
│       └── main.go              # SSH server entry point
├── internal/
│   ├── client/
│   │   └── grpc.go              # gRPC client to backend
│   ├── tui/
│   │   ├── app.go               # Main Bubbletea model
│   │   ├── header.go            # Timer + strikes header component
│   │   ├── footer.go            # Command hints footer
│   │   ├── module_list.go       # Module overview/selection
│   │   └── modules/
│   │       ├── base.go          # Module interface + helpers
│   │       ├── wires.go         # Wires module TUI
│   │       ├── big_button.go    # Big Button module TUI
│   │       ├── simon.go         # Simon Says module TUI
│   │       ├── password.go      # Password module TUI
│   │       ├── keypad.go        # Keypad module TUI
│   │       ├── whos_on_first.go # Who's On First module TUI
│   │       ├── memory.go        # Memory module TUI
│   │       ├── morse.go         # Morse Code module TUI
│   │       ├── maze.go          # Maze module TUI
│   │       ├── needy_vent.go    # Needy Vent Gas module TUI
│   │       └── needy_knob.go    # Needy Knob module TUI
│   └── styles/
│       └── styles.go            # Lipgloss styles
├── proto/                       # Symlink or copy from keep-talking/proto
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Phase 1: Foundation (SSH Server + Basic Shell)

### 1.1 Project Setup
- [x] Initialize Go module: `github.com/ZaneH/keep-talking-tui`
- [x] Add dependencies:
  - `github.com/charmbracelet/wish` (SSH server)
  - `github.com/charmbracelet/bubbletea` (TUI framework)
  - `github.com/charmbracelet/lipgloss` (styling)
  - `github.com/charmbracelet/bubbles` (reusable components)
  - `google.golang.org/grpc` (gRPC client)
- [x] Copy proto files from `keep-talking/proto/`
- [x] Generate Go protobuf client code

### 1.2 gRPC Client
- [x] Create client wrapper in `internal/client/grpc.go`
- [x] Implement methods:
  ```go
  type GameClient interface {
      CreateGame(ctx context.Context) (sessionID string, err error)
      GetBombs(ctx context.Context, sessionID string) ([]*pb.Bomb, error)
      SendInput(ctx context.Context, input *pb.PlayerInput) (*pb.PlayerInputResult, error)
  }
  ```

### 1.3 SSH Server Setup
- [x] Create Wish server with Bubbletea middleware
- [x] Configure SSH key handling (auto-generate host keys)
- [x] Basic connection logging
- [x] Configuration via environment variables:
  - `TUI_SSH_PORT` (default: 2222)
  - `TUI_GRPC_ADDR` (default: localhost:50051)

---

## Phase 2: Core TUI Architecture (COMPLETE)

### 2.1 Main Application Model

```go
type AppState int
const (
    StateLoading AppState = iota
    StateBombSelection
    StateModuleActive
    StateGameOver
)

type Model struct {
    state AppState

    grpcAddr     string
    gameClient   client.GameClient
    sessionID    string
    bombs        []*pb.Bomb
    selectedBomb int

    currentFace int

    selectedModule int

    activeModule modules.ModuleModel

    width  int
    height int
    err    error

    startedAt        time.Time
    duration         time.Duration
    strikeFlashUntil time.Time
    flashStrike      bool
}
```

### 2.2 Header Component
Displays at top of screen (always visible):
```
╔══════════════════════════════════════════════════════════════════════╗
║  KEEP TALKING AND NOBODY EXPLODES   Time: 04:40  Serial: MXSDCZ      ║
║ [ ] [ ] [ ]   Batteries: 3  Ports: RJ45, PS2, SER                    ║
╚══════════════════════════════════════════════════════════════════════╝
```
- [x] Timer countdown (updates every second via `tea.Tick`)
- [x] Strike indicators ([X] struck, [ ] empty)
- [x] Serial number display
- [x] Flashing/color change when strike occurs (red flash)
- [x] Warning colors (<60s yellow, <30s red)
- [x] Batteries and ports display

### 2.3 Footer Component
Command hints (context-sensitive):
```
[ENTER] Pick up bomb | [↑/↓] Navigate | [Q]uit           (Bomb Selection)
[ENTER] Select module | [<]/[>] Flip face | [ESC] Put down | [Q]uit  (Bomb View)
[1-6] Cut wire | [ESC] Back to bomb | [Q]uit                         (Wires Module)
```
- [x] Context-sensitive hints based on current state

### 2.4 Module List/Bomb View
Grid of all modules showing status:
```
BOMB 1 - FRONT

[1] WIRES        [2] PASSWORD
  ○ PENDING        ○ PENDING

[3] BIG BUTTON   [4] SIMON
  ○ PENDING        ○ PENDING
```
- [x] Arrow key navigation (↑/↓/←/→ or vim keys h/j/k/l)
- [x] Enter to select/focus module
- [x] Visual indication of solved/pending (○ PENDING, ✓ SOLVED)
- [x] Module sorting by position (row, col) - fixes non-deterministic map iteration
- [x] Face navigation with [<] and [>] to flip bomb

### 2.5 State Machine
```
StateLoading → StateBombSelection → StateBombView → StateModuleActive
                    ↑                    ↑               ↓
                    └────────────────────┴───────────────┘
```
- [x] Bomb selection state (resting - see bombs on table)
- [x] Bomb view state (picked up - zoomed in on bomb face)
- [x] Module active state (interacting with specific module)
- [x] ESC key to go back from module to bomb view
- [x] ESC key to put down bomb from bomb view

---

## Phase 3: Module Implementations (IN PROGRESS)

Each module implements a common interface:

```go
type ModuleModel interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (tea.Model, tea.Cmd)
    View() string

    ID() string
    ModuleType() pb.Module_ModuleType
    IsSolved() bool
    UpdateState(mod *pb.Module)

    Footer() string
}
```

### 3.1 Wires Module
**Input**: Number key (1-6) to cut wire
**Display**:
```
╔════════════════════════════════════╗
║            WIRES                   ║
║                                    ║
║  1: ▓▓▓▓▓▓▓▓▓▓▓▓▓  RED            ║
║  2: ▓▓▓▓▓▓▓▓▓▓▓▓▓  BLUE           ║
║  3: ───────────────────  YELLOW    ║
║  4: ▓▓▓▓▓▓▓▓▓▓▓▓▓  BLACK           ║
║                                    ║
╚════════════════════════════════════╝
```
- [x] Colored wire blocks using lipgloss
- [x] Show cut state (dashed line)
- [x] 3-6 wires dynamically
- [x] Send WiresInput to backend
- [x] Handle strike/success feedback
- [x] Fix: Local wire state tracking after cut

### 3.2 Unimplemented Module
**Display** for unsupported modules:
```
╔════════════════════════════════════╗
║          PASSWORD                  ║
║                                    ║
║  This module type is not yet       ║
║  implemented.                      ║
║                                    ║
║  Press [ESC] to return to the bomb ║
╚════════════════════════════════════╝
```
- [x] Placeholder for unimplemented module types
- [x] Shows module type name
- [x] Press ESC to go back

### 3.3 Big Button Module
**Input**: `[T]` Tap, `[H]` Hold, `[R]` Release
**Display**:
```
    ┌─────────────┐    
   ╱               ╲   
  ╱                 ╲  
 ╱                   ╲ 
│                     │
│       ABORT         │   ← Label centered
│   Color: RED        │   ← Color centered  
│                     │
 ╲                   ╱ 
  ╲                 ╱  
   ╲               ╱   
    └─────────────┘    
```
- [x] Button color and label display (centering fixed)
- [x] Hold action returns strip color
- [x] Release action with timestamp
- [x] Strike/success feedback
- [x] Key repeat protection (prevents multiple HOLD sends)
- [x] More circular button shape using curved box characters

### 3.4 Morse Code Module
**Input**: `[←]/[→]` or `[h]/[l]` Adjust frequency, `[ENTER]` Transmit
**Display**:
```
╔════════════════════════════════════════════════════════════════╗
║                       MORSE CODE                                ║
╠════════════════════════════════════════════════════════════════╣
║                                                                 ║
║      ━━━━━━━━━━━━━━━━━┤●├━━━━━━━━━━━━━━━━━                      ║
║                         (light blinking)                        ║
║                                                                 ║
║   ┌─────────────────────────────────────────┐                   ║
║   │              FREQUENCY                  │                   ║
║   │         ◄──  3.505 MHz  ──►             │                   ║
║   │  ───────────────────●────────────────   │                   ║
║   └─────────────────────────────────────────┘                   ║
║                                                                 ║
║                    ┌───────────┐                                ║
║                    │    TX     │                                ║
║                    └───────────┘                                ║
║                                                                 ║
╚════════════════════════════════════════════════════════════════╝
```
- [x] Animated blinking light (dots/dashes with timing from reference)
- [x] Amber-colored light with glow effect
- [x] Wire-mounted light design (resistor-like, single horizontal wire)
- [x] Frequency slider with 16 positions using `strings.Builder` for Unicode safety
- [x] TX button for submission
- [x] Send MorseInput (frequency change and TX) to backend
- [x] Handle strike/success feedback
- [x] Pattern loops continuously until solved
- [x] Real-time clock animation (uses `time.Since()` instead of tick counter)
- [x] Index derived from frequency (fallback when backend returns 0)

### 3.5 Simon Says Module
**Input**: Arrow keys (↑/↓/←/→) or letter keys (R/G/B/Y)
**Display**:
```
╔════════════════════════════════════════════════════════════════╗
║                       SIMON SAYS                                ║
║                                                                 ║
║              [RED]                                             ║
║                                                                 ║
║    [YELLOW]          [BLUE]                                    ║
║                                                                 ║
║              [GREEN]                                           ║
║                                                                 ║
║     [↑/↓/←/→] or [R/G/B/Y] Press button | [ESC] Back to bomb   ║
╚════════════════════════════════════════════════════════════════╝
```
- [x] Full button flash (entire button lights up when active)
- [x] Continuous looping sequence with 2s pause between iterations
- [x] Arrow key input: ↑=RED, ↓=GREEN, ←=YELLOW, →=BLUE
- [x] Letter key alternative: R/G/B/Y
- [x] Real-time animation using `time.Since()` for smooth timing
- [x] Send SimonInput to backend on color press
- [x] Handle strike/success feedback
- [x] Reset sequence position on strike

**Timing constants** (from reference):
- FLASH_DURATION = 0.3s
- SEQUENCE_DELAY = 0.75s (time between flashes)
- SEQUENCE_PAUSE = 2.0s (pause between sequences)

### 3.6-3.12 Remaining Modules (TODO)
- [x] Simon Says Module
- [ ] Who's On First Module
- [ ] Memory Module
- [x] Morse Code Module
- [ ] Maze Module
- [ ] Needy Vent Gas Module
- [ ] Needy Knob Module

---

## Phase 4: Game Flow & Polish

### 4.1 Game State Management
- [ ] Create game on SSH connection
- [ ] Fetch bombs and initialize modules
- [ ] Handle module switching
- [ ] Track strikes (flash screen red on strike)
- [ ] Win condition (all modules solved)
- [ ] Loss condition (3 strikes or timer expires)

### 4.2 Timer System
- [ ] Real-time countdown using `tea.Tick`
- [ ] Sync with server `started_at` timestamp
- [ ] Warning colors as time runs low (<1 min, <30s)

### 4.3 Visual Polish
- [ ] Lipgloss color scheme (consistent across modules)
- [ ] Box-drawing characters for borders
- [ ] Strike flash effect (red background momentarily)
- [ ] Solved module green highlight
- [ ] Responsive layout (adapt to terminal size)

### 4.4 Sound Effects (Optional)
- [ ] Terminal bell (`\a`) on strike
- [ ] Terminal bell pattern on explosion
- [ ] Consider: ANSI escape for sound on supporting terminals

---

## Phase 5: Deployment & Testing

### 5.1 Docker Support
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o /tui-server ./cmd/server

FROM alpine:latest
COPY --from=builder /tui-server /tui-server
EXPOSE 2222
CMD ["/tui-server"]
```

### 5.2 Docker Compose Integration
Update existing `docker-compose.yml` to include TUI service:
```yaml
services:
  grpc-server:
    # existing...

  tui-server:
    build: ./keep-talking-tui
    ports:
      - "2222:2222"
    environment:
      - TUI_GRPC_ADDR=grpc-server:50051
    depends_on:
      - grpc-server
```

### 5.3 Testing
- [ ] Unit tests for module rendering
- [ ] Integration tests for gRPC client
- [ ] Manual SSH testing with various terminals

---

## Technical Considerations

### Keypad Symbols Unicode Mapping
```go
var symbolMap = map[pb.Symbol]string{
    pb.Symbol_COPYRIGHT:    "©",
    pb.Symbol_FILLEDSTAR:   "★",
    pb.Symbol_HOLLOWSTAR:   "☆",
    pb.Symbol_SMILEYFACE:   "☺",
    pb.Symbol_DOUBLEK:      "Ж",
    pb.Symbol_OMEGA:        "Ω",
    pb.Symbol_SQUIDKNIFE:   "Ѯ",
    pb.Symbol_PUMPKIN:      "Ѫ",
    pb.Symbol_HOOKN:        "Ҩ",
    pb.Symbol_SIX:          "б",
    pb.Symbol_SQUIGGLYN:    "Ҋ",
    pb.Symbol_AT:           "Ѧ",
    pb.Symbol_AE:           "Æ",
    pb.Symbol_MELTEDTHREE:  "Ӭ",
    pb.Symbol_EURO:         "€",
    pb.Symbol_NWITHHAT:     "Ñ",
    pb.Symbol_DRAGON:       "Ψ",
    pb.Symbol_QUESTIONMARK: "¿",
    pb.Symbol_PARAGRAPH:    "¶",
    pb.Symbol_RIGHTC:       "Ͽ",
    pb.Symbol_LEFTC:        "Ͼ",
    pb.Symbol_PITCHFORK:    "Ѱ",
    pb.Symbol_CURSIVE:      "ϗ",
    pb.Symbol_TRACKS:       "☰",
    pb.Symbol_BALLOON:      "Ѳ",
    pb.Symbol_UPSIDEDOWNY:  "λ",
    pb.Symbol_BT:           "Ƀ",
}
```

### Wire Colors via Lipgloss
```go
var wireColors = map[pb.Color]lipgloss.Style{
    pb.Color_RED:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Background(lipgloss.Color("196")),
    pb.Color_BLUE:   lipgloss.NewStyle().Foreground(lipgloss.Color("21")).Background(lipgloss.Color("21")),
    pb.Color_YELLOW: lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Background(lipgloss.Color("226")),
    pb.Color_BLACK:  lipgloss.NewStyle().Foreground(lipgloss.Color("232")).Background(lipgloss.Color("232")),
    pb.Color_WHITE:  lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("255")),
}
```

### Maze Rendering Challenge
The backend doesn't expose maze walls - it only validates moves. Options:
1. **Option A**: Hardcode the 9 maze variants in the TUI (they're static)
2. **Option B**: Show only player/goal positions, no walls (simpler but less visual)
3. **Option C**: Add maze wall data to proto response (requires backend change)

**Recommendation**: Option A - hardcode mazes. They're defined in the original game and don't change.

---

## Estimated Timeline

| Phase | Description | Effort |
|-------|-------------|--------|
| 1 | Foundation (SSH + gRPC client) | 1-2 days |
| 2 | Core TUI architecture | 2-3 days |
| 3 | Module implementations (11 modules) | 5-7 days |
| 4 | Game flow & polish | 2-3 days |
| 5 | Deployment & testing | 1-2 days |
| **Total** | | **11-17 days** |

---

## Questions Resolved

- **Manual not needed**: Defuser-only view
- **New service**: Separate Go binary, gRPC client to existing backend
- **No emojis**: Unicode/ASCII art only
- **SSH delivery**: Charmbracelet Wish with Bubbletea middleware

---

## Backend API Reference

### gRPC Service
- **Endpoint**: `localhost:50051`
- **Service**: `GameService`

### RPC Methods
1. `CreateGame` - Creates a new game session
2. `GetBombs` - Retrieves all bombs for a session
3. `SendInput` - Sends player input to a module

### Proto Files
Proto files are symlinked/copied from `keep-talking/proto/`:
- `game.proto` - Main service definition
- `player.proto` - Player input/output messages
- `session.proto` - Session messages
- `bomb.proto` - Bomb entity
- `modules.proto` - Module definitions
- `common.proto` - Common types (Color, Direction, etc.)
- `*_module.proto` - Module-specific types

### Module Types (pb.ModuleType)
- `CLOCK = 1`
- `WIRES = 2`
- `PASSWORD = 3`
- `BIG_BUTTON = 4`
- `SIMON = 5`
- `KEYPAD = 6`
- `WHOS_ON_FIRST = 7`
- `MEMORY = 8`
- `MORSE = 9`
- `NEEDY_VENT_GAS = 10`
- `NEEDY_KNOB = 11`
- `MAZE = 12`
