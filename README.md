> 📌 Mirror any X (Twitter) thread on‑chain with a single mention — permanent, searchable, and shareable. 

# ThreadMirror

<img src="logo.png" alt="ThreadMirror Logo" width="200" />

ThreadMirror lets you archive any X (Twitter) thread with a single mention. Reply **@threadmirror** under the thread—you'll receive a permanent link, an AI-generated summary, and a long shareable image, all stored immutably on Filecoin PDP.


## ✨ Highlights 

* **One-step archive (mention @threadmirror)** – No plugin, no signup: just mention `@threadmirror` beneath any thread and the bot does the rest.
* **Immutable storage on Filecoin PDP –** Threads are saved as content-addressable data that remain online permanently and cannot be tampered with, even after deletion.
* **Instant AI summary** – A concise, LLM-generated digest arrives with the reply so you can grasp the whole thread at a glance.
* **Shareable permalink** – Receive a clean `https://threadmirror.xyz/thread/<id>` link that never breaks—perfect for bookmarking or sharing anywhere.
* **End-to-end share image** – The bot also returns a single long image of the full thread, great for posting in chats or saving offline.
* **Multilingual & cross-platform roadmap** – Future support for translating full threads and ingestion from Telegram, Bluesky, TruthSocial, and more.

## 📝 Usage

1. On X (Twitter), reply to any thread with **@threadmirror**.  
2. The bot will:  
   1. Fetch the entire thread.  
   2. Store the tweets on PDP.  
   3. Generate an AI summary.  
   4. Reply with (a) a permanent link (e.g., `https://threadmirror.xyz/thread/<id>`), and (b) a long image snapshot of the thread for easy sharing.  

That's it—no additional setup required.

## 🎬 Demo

Watch ThreadMirror in action on YouTube: [https://www.youtube.com/watch?v=J-D1DlNxQPY](https://www.youtube.com/watch?v=J-D1DlNxQPY)

---

## 🛣️ Roadmap

- 🌐 Full‑thread translation

- 📥 Multi‑platform ingestion (Telegram, Bluesky, TruthSocial, …)

- 🖥️ Web UI upgrades: filtering, search, browser extension

## 🚀 Getting Started

### 1 · Prerequisites

* **Go** ≥ 1.24
* **Node.js** ≥ 22 (for frontend development)
* **Docker & Docker Compose** (optional, but the easiest way to start)

### 2 · Clone the repository

```bash
git clone https://github.com/ipfs-force-community/threadmirror.git
cd threadmirror
```

### 3 · Run with Docker Compose (recommended)

```bash
# Copy and adjust configuration
cp example.env .env

# Build images and start services (backend web, bot, db, redis)
docker compose up --build
```

Once started:

* API: `http://localhost:8089`
* The sample frontend is located in the `web` sub-folder; you can deploy it separately.

### 4 · Local development workflow

1. Start Postgres & Redis (use the `db` and `redis` services in the compose file or your own instances).
2. Backend:

   ```bash
   make setup     # download deps & code generation
   make dev       # run in debug mode (equivalent to: go run ./cmd/*.go --debug server)
   ```

3. Frontend:

   ```bash
   cd web
   npm install
   npm start      # http://localhost:3000
   ```

4. Bot:

   ```bash
   ./bin/threadmirror --debug bot
   ```

## 🛠️ CLI Commands

| Command                    | Purpose                               |
| -------------------------- | ------------------------------------- |
| `threadmirror server`      | Start the HTTP API server             |
| `threadmirror migrate`     | Run database migrations               |
| `threadmirror bot`         | Run the @mention bot                  |
| `threadmirror reply`       | Manually reply to a given mention     |

Run `threadmirror <command> --help` for flag details.

---

## 🗂️ Directory Layout (short version)

```
threadmirror/
├── api/            # OpenAPI spec and templates
├── cmd/            # CLI entry points (server, bot, migrate …)
├── internal/       # Business logic (API, services, repos, jobs, bot …)
├── pkg/            # Reusable libraries (auth, db, ipfs, llm, xscraper …)
├── web/            # React frontend
├── Makefile        # Dev scripts
└── docker-compose.yml
```

## 🤝 Contributing

We ❤️ contributions! To get started:

1. Ensure make test passes.
2. Run make lint and fix any issues.
3. Follow the API coding guidelines and git‑flow.


## License

Apache-2.0 © IPFS Force Community 
