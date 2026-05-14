Complexity level: 🔴 Complex — real-time context aggregation, AI decision engine, multi-API integration, feedback loop, dan harus demo-ready untuk kompetisi. Mari kita buat blueprint-nya.Kompleksitas: 🔴 Complex. NowHere punya 3 layer yang menarik secara arsitektur — context aggregation, Gemini decision engine, dan preference feedback loop. Aku akan breakdown dalam 3 diagram terpisah biar clean.

Dimulai dari system overview dulu — big picture semua komponen dan bagaimana mereka ngobrol satu sama lain.Empat layer utama — klik tiap komponen untuk deep dive. Sekarang kita lihat **request flow**-nya: apa yang terjadi dari detik user buka app sampai rekomendasi muncul.Perhatikan feedback loop di kanan — kalau user reject, sistem tidak generate ulang dari nol. Dia re-use context yang sudah ada, cuma tambahkan constraint "exclude tempat ini" ke prompt Gemini. Itu yang bikin retry-nya cepat. Sekarang kita masuk ke **tech stack + project structure** lengkap.

---

## Tech stack

| Layer | Technology | Why |
|---|---|---|
| Frontend | React + Vite (PWA) | PWA supaya bisa akses GPS tanpa install; Vite untuk dev speed |
| Styling | Tailwind CSS | Utility-first, mobile-first cepat |
| Backend | Go + Fiber | Stateless, fast startup di Cloud Run, concurrency bagus untuk parallel API calls |
| AI | Gemini 2.0 Flash | Murah, cepat, cukup smart untuk single-decision reasoning; bukan overkill |
| Database | Firestore | Real-time, schema-less cocok untuk preference data yang evolve |
| Auth | Firebase Auth | Google SSO built-in — 1 klik untuk user |
| Deploy | Cloud Run | Scales to zero, cocok untuk hackathon budget; full Google stack |
| External | Google Maps + Places API | Lokasi + nama tempat yang akurat untuk Indonesia |
| External | Open-Meteo | Free weather API, no key needed untuk MVP |

---

## Project structure

```
nowhere/
├── frontend/                  # React PWA
│   ├── src/
│   │   ├── components/
│   │   │   ├── DecisionCard.jsx     # Core UI — satu rekomendasi
│   │   │   ├── ContextBadge.jsx     # GPS, cuaca, waktu display
│   │   │   └── ActionButtons.jsx    # Oke gas / Kasih yang lain
│   │   ├── hooks/
│   │   │   ├── useLocation.js       # GPS wrapper
│   │   │   └── useDecision.js       # API call + state
│   │   └── main.jsx
│   └── vite.config.js
│
├── backend/                   # Go + Fiber
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── handler/
│   │   │   ├── decide.go            # POST /decide — main endpoint
│   │   │   └── feedback.go          # POST /feedback — accept/reject
│   │   ├── service/
│   │   │   ├── context.go           # Agregasi GPS + weather + places
│   │   │   ├── gemini.go            # Prompt builder + Gemini call
│   │   │   └── preference.go        # Read/write Firestore preference
│   │   └── model/
│   │       ├── request.go
│   │       └── decision.go
│   ├── .env.example
│   └── Dockerfile
│
└── README.md
```

---

## Development roadmap

Ini konteks hackathon, jadi tiap fase harus menghasilkan sesuatu yang bisa di-demo.

**Phase 1 — Context works** *(Hari 1)*
- Setup repo, Cloud Run, Firebase project
- GPS + waktu → tampil di UI (buktikan context bisa diambil)
- Weather API call jalan lokal
- Milestone demo: "App tahu aku di mana dan cuacanya apa"

**Phase 2 — Gemini decides** *(Hari 2–3)*
- Backend endpoint `POST /decide` jalan
- Gemini prompt builder dengan semua context
- Single recommendation muncul di UI dengan reasoning-nya
- Milestone demo: "Satu rekomendasi muncul, ada alasannya"

**Phase 3 — Feedback loop** *(Hari 4)*
- Reject → retry dengan exclude constraint
- Accept → simpan ke Firestore
- Preference makin lama makin personal
- Milestone demo: "Reject 3x, rekomendasinya beda semua"

**Phase 4 — Polish & ship** *(Hari 5–6)*
- UI mobile-first yang clean
- Error handling (GPS denied, API timeout, Gemini error)
- Deploy ke Cloud Run
- Milestone demo: Full flow dari buka app → dapat rekomendasi → navigasi ke Maps

---

## Risks & considerations

- GPS di browser butuh HTTPS — pastikan dev pakai `localhost` atau langsung deploy ke Cloud Run dari awal
- Places API bisa costly kalau per-request tanpa caching — cache hasil Places berdasarkan radius 500m, expire 30 menit
- Gemini kadang verbose — prompt harus strict: "Respond ONLY in JSON format: `{place, reason, duration_minutes}`"
- Jika Gemini timeout (>3s): fallback ke cached recommendation dari sesi sebelumnya, jangan blank screen
- Scope risk: jangan tambahin fitur social/sharing dulu — MVP adalah "satu tombol, satu keputusan"

---

## Quick start

```bash
# 1. Clone dan setup
git clone https://github.com/rillyayidan/nowhere && cd nowhere
cp backend/.env.example backend/.env
# isi GEMINI_API_KEY, GOOGLE_MAPS_KEY, FIREBASE_PROJECT_ID

# 2. Jalankan backend
cd backend && go run ./cmd/main.go
# → http://localhost:8080

# 3. Jalankan frontend
cd ../frontend && npm install && npm run dev
# → http://localhost:5173
```
