# Architecture — Mesh Video Conferencing (Signaling-Only Server)

Status: authoritative build reference.
Scope: ~10 participants, small-group conference, cross-network, zero-cost hosting.

## 1. System overview

Two components, no media infrastructure:

1. **Go signaling server** — my app. Room bookkeeping, signaling relay, chat fan-out. Zero media traffic; carries only JSON (SDP, ICE candidates, chat text).
2. **Public STUN server** — called by the *browser* (not the app). The app only serves the STUN URL to clients inside the page render / ICE config.

Media flows directly browser-to-browser over WebRTC (SRTP). The server never sees a media packet.

```
Browser A ──WS──┐                    ┌── WS ── Browser B
                ▼                    ▼
         ┌───────────────────────────────┐
         │  Go App (signaling server)    │
         │  registry actor (1)           │
         │  room actor (per room)        │
         │  peer session (per socket)    │
         │  zero media, text only        │
         └───────────────────────────────┘
Browser A ◄─────── SRTP (direct) ───────► Browser B
Browser A ◄─────── SRTP (direct) ───────► Browser C
   ...one peer connection per remote peer (N−1)
                ▲
                │  ICE candidates discovered via
        Public STUN  (browser calls out, free)
        default: stun:stun.nextcloud.com:443
```

### Stack decision (held firm)

- Full-mesh conference: every client holds N−1 `RTCPeerConnection`s.
- No TURN, no SFU, no relay, no media relaying of any kind on the server.
- Networking kept deliberately simple: STUN-only.
- This is a documented tradeoff, not an oversight. See §9 (Limitations).

## 2. Component inventory

### 2.1 Go server (internal/)

| Component | Responsibility |
|---|---|
| HTTP layer | Routes: `/` welcome, `/room/create` (UUID redirect), `/room/:id` page render, `/assets/*` static. Optional `?name=` for display name. No auth. |
| Registry actor (1 goroutine, process-wide) | Owns `map[roomID]*Room`. Get-or-create on join. Delete room on idle. Single owner → no locks. |
| Room actor (1 goroutine per room) | Owns roster `map[peerID]*session`. Relay of `signal` envelopes to targets. Broadcast `join`/`leave` presence. Chat broadcast. Idle TTL → self-deregister at registry. |
| Peer session | One per WebSocket. Reads WS → forwards to room actor; outbound channel → its own socket writer. Single-writer rule per socket. Server keepalive: WS ping every 30s, read deadline 60s. Removal paths — socket close, ping timeout, writer error — all notify room actor, which owns the roster. |

Goroutine sizing: `1 registry + 1 per room + 1 per socket (reader) + 1 per socket (writer)`. Room actor and peer session may be combined when a room is small; keep ownership as described regardless of packing.

### 2.2 Browser client (assets/javascript)

| Component | Responsibility |
|---|---|
| Media capture | `getUserMedia` once; share one local stream across all peer connections |
| PC manager | Create one `RTCPeerConnection` per remote peer on `joined`/`peer-joined`; close on `peer-left` |
| Negotiator | **Perfect negotiation** (webrtc-pc pattern). Roles per pair: lexicographically smaller peer ID is polite, larger is impolite — fixed, complementary, never random. BOTH sides may create offers; glare is expected. On glare the polite side rolls back its pending offer and answers; the impolite side ignores an incoming offer while its own offer is outstanding. |
| ICE | `onicecandidate` → send candidate to target peer via signaling WS |
| Signal handler | Inbound desc → `setRemoteDescription` on the PC keyed by sender ID. Inbound candidate → `addIceCandidate`; if the remote description is not yet set, buffer in a per-sender queue, flush after `setRemoteDescription`, drop on `peer-left`. |
| Renderer | Remote streams → video elements, grid layout, local preview, leave button |
| Reconnect | On WS drop: tear down peer connections, rejoin with backoff + jitter (fresh peerID). `peer-left`/`peer-joined` already fired → everyone rebuilds one PC and renegotiates for this client. Accepted v1 cost; a flapping client causes churn. |

### 2.3 STUN

| Component | Responsibility |
|---|---|
| Public STUN | ICE server for NAT mapping (srflx candidates). Default: `stun:stun.nextcloud.com:443`, operated by Nextcloud GmbH (EU). Alternative: `stun:stun.cloudflare.com:3478` or self-hosted coturn. Configurable via `STUN_URL`. |

STUN is contacted by browsers, never by the app server.

## 3. Concurrency model (no locks)

All shared state is owned by exactly one goroutine. Anything else reaches it only by sending a command on that goroutine's channel. This is the `pkg/chat/hub.go` pattern (register / unregister / broadcast channels) extended to rooms and signaling.

### Ownership rules (enforce at review)

1. Registry goroutine is the only writer of the room map.
2. Room actor is the only writer of its roster.
3. Peer session writer goroutine is the only writer of its socket.
4. Shutdown: a channel is never closed by a sender; an actor that exits closes its own `done` channel; every send is guarded by `select { case ch <- cmd: case <-done: }`. The guard alone is insufficient — with `done` closed AND channel ready, `select` picks randomly, so a send may land in a dead actor. Every send path must handle failure: the registry, on losing a send to a room actor, deletes the room from its map, creates a fresh room, and re-sends. A `join` must never be dropped — a dropped join leaves the client without a `joined` reply forever.
5. No `sync.Mutex` anywhere in the app. (If one appears, the ownership rule was violated.)

### Room lifecycle (GC)

Room actor runs an idle timer (`ROOM_TTL`, default 5m). Any activity resets it. On expiry with zero members the actor:
1. sends `delete` to the registry;
2. closes its `done`; returns.
Registry removes the room from its map. No room leaks, no goroutine leaks.

### Join/GC serialization

Race: an actor's idle timer fires while a `join` is queued on its channel; the actor counts zero members (session not yet registered), exits, and closes `done` — the queued join dies. Resolution: the registry send-failure retry (rule 4) is self-healing. Belt-and-braces: the client times out a missing `joined` (5s) and retries. A join must not be able to hang a session.

### Member liveness

A member is removed on socket close, WS ping timeout, or writer error (§2.1). Without keepalive, a half-open socket (client power loss, roaming NAT expiry — no TCP FIN) keeps the roster non-empty forever: GC requires zero members, so the room never dies and peers hold dead connections. The ping deadline is mandatory, not optional.

## 4. Wire protocol

One WebSocket per member: `ws://host/room/:id/ws` (wss behind TLS). All messages JSON.

### Client → server

```json
{ "type": "join",  "name": "alice" }
{ "type": "signal", "target": "<peerID>", "data": { /* RTCSessionDescription | RTCIceCandidateInit */ } }
{ "type": "chat", "text": "hello" }
```

### Server → client

```json
{ "type": "joined",     "id": "<peerID>", "peers": [ {"id":"...","name":"..."} ] }  // peers excludes self
{ "type": "peer-joined", "peer": {"id":"...","name":"..."} }
{ "type": "peer-left",   "id": "<peerID>" }
{ "type": "signal",     "from": "<peerID>", "data": { /* desc | candidate */ } }
{ "type": "chat",       "text": "hello", "from": "alice" }
```

Server behavior for `signal`: validate `target` member exists, forward verbatim (from = sender). No SDP parsing, no candidate inspection. Dumb pipe, smart client.

Presence events are authoritative: clients create/close peer connections based on `peer-joined` / `peer-left`. `joined.peers` seeds the initial roster.

## 5. Flows

### Join
1. Browser → page → WebSocket → `join`.
2. Server: registry get-or-create room → room actor registers session → replies `joined` (roster) → broadcasts `peer-joined` to others.
3. Each existing client creates a PC toward the newcomer; the newcomer creates PCs toward all of `joined.peers`. Both sides negotiate simultaneously — glare is normal and resolved by polite-rollback per pair (§2.2). ICE trickles both ways; early candidates are buffered until the remote description is set.
4. Media flows directly between browsers.

### Room create
`/room/create` → 302 to `/room/<uuid>`. The room exists implicitly on first join (registry get-or-create). No pre-registration step.

### Chat
Client → `chat` → room actor broadcasts to all members → clients render. Same socket as signaling; no dedicated chat path.

### Teardown
Socket close/error → reader goroutine → room actor removes session → `peer-left` broadcast → remaining clients close that PC. Room idles toward GC.

## 6. Configuration (env vars)

| Var | Default | Purpose |
|---|---|---|
| `PORT` | `:8080` | HTTP server bind address |
| `STUN_URL` | `stun:stun.nextcloud.com:443` | ICE server served to clients |
| `ROOM_TTL` | `5m` | Room idle timeout before GC |
| `ENVIRONMENT` | `development` | `production` → wss:// + TLS expectations (reverse proxy) |

The STUN URL is injected into the rendered page (ICE config embedded in a small inline script or served from the room handler). No separate config endpoint needed at this scale.

## 7. Deployment

Server is text-only at the network layer: it never consumes media bandwidth. It runs on any always-on box:

- Home machine + router port-forward 80/443 — only if the ISP provides a public IP (check for CGNAT first; dynamic DNS for name).
- Free-tier PaaS / worker hosting for WebSocket rooms.
- Old hardware (Raspberry Pi class) behind DDNS.

Browser media does not pass through the deployment box at all — only signaling text does.

## 8. Build order

1. `pkg/rooms` — registry + room actors, presence, GC. Unit tests (race).
2. WS protocol + handlers — join/signal/chat relay.
3. HTTP routes — welcome, room/create, room page render, static assets.
4. Client `mesh.js` — PC manager, perfect negotiation, signal handler, renderer.
5. Chat UI — fold onto same socket.
6. Integration test: two+ browser sessions in one room.

## 9. Limitations (accepted, documented)

| Limitation | Impact | Notes |
|---|---|---|
| STUN-only, no TURN | Symmetric NAT, CGNAT, strict corporate firewalls fail to connect (~15-30% of random Internet peers) | Deliberate simplicity cut. Revisit if coverage matters: see §10 |
| Full mesh bandwidth | Each peer uploads N−1 streams. 10 peers @ 720p ≈ 12+ Mbps up/down per peer | Requires roughly symmetric connections; reduce to 480p or shrink rooms on asymmetric ISPs |
| No media in server | One host disconnects = its streams vanish for everyone | Expected in a mesh; client treats per-peer loss independently |
| Signaling reachability | Server must be publicly reachable for joins | Home hosting depends on public IP / port-forward / PaaS |
| No auth | Anyone with the link joins | Out of scope v1 |
| No persistence | Chat/rooms vanish on idle GC or restart | Out of scope v1 |

## 10. Deferred (design notes only)

- **TURN relay**: the only universal NAT fallback. Adding later = serve TURN creds in ICE config; browser already handles relay automatically. No client logic change. coturn can dual-role as STUN+TURN on the same box (the "self-host instead of Nextcloud" option).
- **SFU**: not needed at ≤10 participants; if rooms grow, SFU moves media concentration server-side and the client PC manager collapses to one connection.
- **Peer-hosted / role-switching SFU**: documented alternative; rejected for v1 (role handoff = full renegotiation + failover protocol, not worth it at this scale).
- **Viewer (broadcast-only) mode**: each viewer would force N−1 uploads from participants; deferred.
- **Simulcast / recording / chat history / rate limiting**: deferred.

## 11. Security model

- Media: WebRTC DTLS-SRTP per leg, always on. Server cannot see media (it is never routed through it).
- Signaling: plain WS in development; wss behind reverse proxy in production. Trust note: the signaling server is a trusted party for media confidentiality. SDP carries DTLS fingerprints; a compromised server can substitute its own SDP, MITM the DTLS-SRTP handshake per leg, and decrypt/re-encrypt media. "Server never routes media" excludes it from packet handling, not from key-material influence. wss protects against passive network observers only. Accepted for v1.
- STUN: sees only source IP:port for NAT mapping; no media, no payload. Nextcloud GmbH operates the default; alternatives available via `STUN_URL`.
- No secrets in the app (STUN requires none; TURN creds, if ever added, must be ephemeral/HMAC and env-injected — never hardcoded).