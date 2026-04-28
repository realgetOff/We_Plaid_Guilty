# Frontend — ft_transcendence · We Plaid Guilty

> **Role:** mforest- — Frontend Developer

---

## Description

This is the frontend of **ft_transcendence**, a real-time multiplayer web application built for the 42 curriculum. The project is a **Gartic Phone**-inspired game where players write prompts, draw them, guess what others drew, and laugh at the results. It also features an **AI Game mode** where an LLM generates the prompts and players vote on each other's drawings.

The UI is deliberately styled after **Apple HyperCard** (1987), giving the whole app a retro Mac aesthetic: pixel-perfect windows, a simulated menu bar with a live clock, and a footer with legal links. Everything runs inside a single "Mac window" component that contextually updates its title and card header based on the current route.

The application is **desktop-only** (mobile is blocked at the App level by user-agent detection).

---

## Tech Stack

| Layer | Technology |
|---|---|
| Framework | React 18 (Vite) |
| Routing | React Router v6 |
| Real-time | WebSocket (custom singleton client) |
| Notifications | React Context API |
| Toasts | react-hot-toast + custom ToastContainer |
| Styling | Custom CSS - BEM typo, no CSS framework |
| Auth | JWT (localStorage) + 42 OAuth callback |
| Build tool | Vite |

<p align="center">
  <img src="../readme_img/stack_explanation.png">
</p>

---

## Project Structure

```
src/
├── api/
│   ├── auth.js          # REST calls for /api/auth/*
│   ├── rooms.js         # REST calls for /api/rooms/* and /api/ai-rooms/*
│   └── socket.js        # WebSocket singleton (connect/send/addListener)
├── components/
│   └── common/
│       ├── MacWindow.jsx          # HyperCard window wrapper (title bar, footer)
│       ├── Navbar.jsx             # Simulated Mac menu bar + live clock
│       ├── NotificationBell.jsx   # Dropdown bell for game invites
│       ├── NotificationContext.jsx # Global invite/notification state
│       └── ToastContainer.jsx     # Pop-up toasts for incoming game invites
├── pages/
│   ├── Auth/
│   │   ├── Login.jsx      # Guest login + 42 OAuth login
│   │   └── Callback.jsx   # OAuth callback handler (token → localStorage)
│   ├── Game/              # Classic Gartic Phone mode
│   │   ├── HomeGame.jsx   # Game mode selector (create/join classic or AI)
│   │   ├── CreateGame.jsx # Host lobby with room code, chat, friends sidebar
│   │   ├── JoinGame.jsx   # Validates room code, redirects to Lobby
│   │   ├── Lobby.jsx      # Non-host waiting room (chat, player list)
│   │   ├── Game.jsx       # Main game controller (write → draw → guess → gallery)
│   │   ├── WritePrompt.jsx # Timed prompt input
│   │   ├── DrawBoard.jsx  # Full canvas drawing board (tools, shapes, fill, undo)
│   │   ├── GuessPrompt.jsx # Timed guess input with drawing display
│   │   └── Gallery.jsx    # End-of-game chain viewer
│   ├── AIGame/            # AI-assisted game mode
│   │   ├── AICreateGame.jsx # AI host lobby
│   │   ├── AIJoinGame.jsx   # Validates AI room, redirects to AILobby
│   │   ├── AILobby.jsx      # AI game waiting room
│   │   ├── AIGame.jsx       # AI game controller (draw → vote → gallery)
│   │   ├── AIDrawBoard.jsx  # Drawing board with title/description fields
│   │   ├── AIVotePanel.jsx  # Star-rating vote panel for each drawing
│   │   └── AIGallery.jsx    # Results gallery with ranking and scores
│   ├── Profile/
│   │   └── Profile.jsx    # View/edit username, avatar, color, font style
│   ├── Friends/
│   │   └── Friends.jsx    # Friend list, requests, online status, invite to room
│   ├── Home/
│   │   └── Home.jsx       # Landing page
│   ├── Legal/
│   │   ├── Privacy.jsx    # Privacy Policy page
│   │   ├── Tos.jsx        # Terms of Service page
│   │   └── Credits.jsx    # Team credits (with a hidden "find lviravon" easter egg unlock)
│   └── NotFound/
│       └── NotFound.jsx   # 404 page / "Find lviravon" hidden game
└── styles/
    ├── global.css         # Base reset, body, layout
    └── hypercard.css      # HyperCard design system (window, menubar, footer, cards)
```

---

## Features

### Authentication
- **Guest session**: one-click anonymous login via `POST /api/auth/player`; token stored in `localStorage`
- **42 OAuth**: redirects to `/login/42`, backend returns JWT via query param at `/callback`
- **Persistent session**: token presence checked on every route/navbar render
- **Sign out**: clears all localStorage keys and reloads

<p align="center">
  <img src="../readme_img/auth_explanation.png">
</p>

### Navigation & Layout
- **MacWindow**: wraps every page in a simulated HyperCard window; title and card header update per route
- **Navbar**: retro Mac menu bar with live AM/PM clock, username display, login/logout toggle, profile and friends links
- **NotificationBell**: shows a badge count of pending game invites; dropdown to accept or refuse each one
- **ToastContainer**: floating pop-up invites with a 15-second auto-dismiss progress bar

### Classic Game Mode (Gartic Phone)
- **HomeGame**: entry point to create or join a classic or AI room; validates room codes client-side and server-side before navigating
- **CreateGame / Lobby**: real-time lobby with WebSocket-synced player list, chat, and a collapsible friends sidebar to invite online friends directly
- **Game flow** (orchestrated by `Game.jsx`):
  - `write` phase → `WritePrompt`: timed prompt input
  - `draw` phase → `DrawBoard`: full canvas tool with pen, eraser, flood fill, color picker, line/rect/circle shapes (outline + filled), size slider, 16-color palette + custom color picker, undo (up to 30 steps), clear
  - `guess` phase → `GuessPrompt`: timed guess with drawing display
  - `gallery` phase → `Gallery`: scrollable chain viewer showing the full prompt → drawing → guess sequence
- All phases auto-submit on timer expiry

<p align="center">
  <img src="./readme_img/aigame_explanation.png">
</p>

### AI Game Mode
- Same lobby flow as classic, but triggers an AI-generated prompt server-side
- `AIDrawBoard` adds optional **title** and **description** fields to drawings
- `AIVotePanel`: 1–5 star voting on every other player's drawing; submits only when all drawings are rated
- `AIGallery`: sorted leaderboard with medal rankings and star display

<p align="center">
  <img src="../readme_img/game_explanation.png">
</p>

### Profile
- View any user's profile by username (`/profile/:username`)
- Edit own profile: username, avatar upload (base64 → WebSocket), username color (8 presets), font style (normal / bold / italic)
- Live preview of username style before saving
- Guest accounts see a banner and cannot edit

### Friends
- Add friends by exact username (with duplicate / self-add guards)
- Accept or reject incoming requests; cancel outgoing requests
- Online/offline status updates in real time via WebSocket
- Send a game invite to an online friend by entering a room code
- Guest accounts see a blocked state

<p align="center">
  <img src="../readme_img/friendsprofile.png">
</p>

### Legal
- `/privacy` — Privacy Policy (data collection, usage, user rights)
- `/tos` — Terms of Service (conduct, multi-user requirements, disclaimer)
- Both pages are accessible from the footer on every screen

### Easter Egg — "Find lviravon"
The 404 page is a **Where's Waldo** mini-game. Five themed photos hide a tiny "lviravon" character at a precise coordinate. Each correct click triggers a win card with a fabricated quote. Completing all five levels redirects to the Credits page with lviravon's credit replaced by pmilner-'s expanded role. Result is persisted in `localStorage`.

---

## WebSocket Architecture

All real-time communication goes through a **singleton WebSocket client** (`src/api/socket.js`):

- **`connect()`** — opens the socket, sends an `authenticate` message with the JWT immediately on open
- **`send(payload)`** — queues messages if the socket is not yet authenticated (`wsAuthReady`), flushes on `auth_ok`
- **`addListener(fn)` / `removeListener(fn)`** — pub/sub pattern used by every page component
- Handles `profile_updated` messages to refresh the JWT and username in localStorage automatically
- Closes with code `4000` → forces logout redirect

<p align="center">
  <img src="../readme_img/socket_explanation.png">
</p>

---

## API Integration

| File | Calls |
|---|---|
| `api/auth.js` | `GET /api/auth/me`, `POST /api/auth/logout`, OAuth URL builder |
| `api/rooms.js` | `GET /api/rooms/:code`, `GET /api/ai-rooms/:code` |
| `pages/Auth/Login.jsx` | `POST /api/auth/player` (guest), redirect to `/login/42` |

Base URL is read from `VITE_API_URL` env var; falls back to `window.location.origin`.

---

## Environment Variables

```env
VITE_API_URL=https://your-backend-url
VITE_WS_URL=wss://your-backend-url/ws
```

Both fall back gracefully to `window.location.origin` / `ws[s]://host/ws` if not set.

---

## Running the Frontend

### Prerequisites

- Node.js ≥ 18
- A running backend (see backend README)

### With Docker (full project)

```bash
docker compose up --build
```

The frontend is served by the Docker setup. A single command starts the whole application.

---

## Modules Covered (frontend contribution)

| Module | Type | Points | Notes |
|---|---|---|---|
| Use a frontend framework (React) | Minor | 1 | React 18 + Vite |
| Real-time features (WebSocket) | Major | 2 | Full WS singleton, lobby chat, game phases |
| User interaction (chat, profile, friends) | Major | 2 | Chat in lobby, friends system, profile page |
| Standard user management | Major | 2 | Avatar upload, username/style edit |
| AI LLM interface (drawing + voting) | Major | 2 | AIGame flow, AIVotePanel, AIGallery |
| Custom design system | Minor | 1 | HyperCard theme, 10+ reusable CSS components |

---

## AI Usage

AI was used during frontend development for:
- For repetitive component structures (lobby pages, CSS skeletons)
- Debugging edge cases in the WebSocket authentication flow
- Generating initial CSS for the HyperCard design system

---

## Known Limitations

- Mobile devices are blocked intentionally (the canvas drawing board is not touch-optimized)
- CSS not working correct on Firefox
- No offline support

---

*This README covers the frontend part only. See the global README for the full project description, backend architecture, database schema, and deployment instructions.*
