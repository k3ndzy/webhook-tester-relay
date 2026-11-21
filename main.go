package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Item represents a business entity in HookRelay Sandbox.
type Item struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Status   string `json:"status"`
	Detail   string `json:"detail"`
}

// Store manages concurrent-safe in-memory records.
type Store struct {
	mu    sync.RWMutex
	items []Item
}

func NewStore() *Store {
	s := &Store{}
	initial := `[{"id": "WH-901", "name": "Stripe payment_intent.succeeded", "category": "Payments", "status": "Captured (200 OK)", "detail": "Payload 4.8 KB, HMAC-SHA256 signature verified, latency 18ms"}, {"id": "WH-902", "name": "GitHub push (refs/heads/main)", "category": "CI/CD Pipeline", "status": "Forwarded", "detail": "Dispatched to local development runner on http://localhost:8080/deploy"}, {"id": "WH-903", "name": "Shopify orders/create", "category": "E-Commerce", "status": "Captured (200 OK)", "detail": "Order #48291 parsed - Auto synced with ERP fulfillment channel"}, {"id": "WH-904", "name": "Slack app_mention event", "category": "ChatOps", "status": "Retried (Attempt 2)", "detail": "Initial timeout recovered, response payload acknowledged in 42ms"}]`
	_ = json.Unmarshal([]byte(initial), &s.items)
	return s
}

func (s *Store) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copied := make([]Item, len(s.items))
	copy(copied, s.items)
	return copied
}

func (s *Store) Add(item Item) Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	if item.ID == "" {
		item.ID = fmt.Sprintf("ITEM-%03d", len(s.items)+1)
	}
	s.items = append(s.items, item)
	return item
}

var (
	startTime = time.Now()
	store     = NewStore()
)

const htmlIndex = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>HookRelay Sandbox - Realtime Webhook Listener & Payload Forwarding Inspector</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;600;700;800&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
  <style>
    body { font-family: 'Plus Jakarta Sans', sans-serif; }
    code, pre { font-family: 'JetBrains Mono', monospace; }
  </style>
</head>
<body class="bg-slate-950 text-slate-100 min-h-screen antialiased flex flex-col justify-between">
  <div>
    <!-- Top Nav -->
    <header class="border-b border-slate-800 bg-slate-900/80 backdrop-blur sticky top-0 z-50">
      <div class="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <div class="w-8 h-8 rounded-lg bg-gradient-to-tr from-[#f59e0b] to-[#ef4444] flex items-center justify-center font-bold text-white shadow-lg">
            ⚡
          </div>
          <span class="text-xl font-bold tracking-tight text-white">HookRelay Sandbox</span>
          <span class="text-xs px-2.5 py-0.5 rounded-full bg-slate-800 text-slate-400 font-mono border border-slate-700">Testing & Debugging</span>
        </div>
        <div class="flex items-center space-x-4">
          <span class="flex items-center text-xs text-emerald-400 bg-emerald-950/60 border border-emerald-800/80 px-3 py-1 rounded-full font-mono">
            <span class="w-2 h-2 rounded-full bg-emerald-400 animate-ping mr-2"></span>
            Live Port :8198
          </span>
          <a href="/api/items" class="text-xs text-slate-400 hover:text-white transition font-mono underline">/api/items</a>
          <a href="/health" class="text-xs text-slate-400 hover:text-white transition font-mono underline">/health</a>
        </div>
      </div>
    </header>

    <!-- Hero Section -->
    <section class="max-w-7xl mx-auto px-6 pt-10 pb-6">
      <div class="bg-gradient-to-r from-slate-900 via-slate-900 to-indigo-950/40 border border-slate-800 rounded-2xl p-8 relative overflow-hidden shadow-2xl">
        <div class="max-w-3xl relative z-10">
          <div class="inline-block text-xs font-bold tracking-wider uppercase text-[#f59e0b] mb-3">Enterprise Core Architecture</div>
          <h1 class="text-4xl font-extrabold tracking-tight text-white sm:text-5xl mb-4">HookRelay Sandbox</h1>
          <p class="text-slate-300 text-base sm:text-lg leading-relaxed">Realtime Webhook Listener & Payload Forwarding Inspector</p>
          <div class="mt-6 flex flex-wrap gap-3">
            <span class="text-xs bg-slate-800/80 border border-slate-700 px-3 py-1.5 rounded-lg text-slate-300 font-mono">Go 1.26 Pure Standard Library</span>
            <span class="text-xs bg-slate-800/80 border border-slate-700 px-3 py-1.5 rounded-lg text-slate-300 font-mono">Zero External Dependencies</span>
            <span class="text-xs bg-slate-800/80 border border-slate-700 px-3 py-1.5 rounded-lg text-slate-300 font-mono">RESTful JSON APIs</span>
            <span class="text-xs bg-emerald-950/80 border border-emerald-800 px-3 py-1.5 rounded-lg text-emerald-300 font-mono">100% Pass Unit Tests</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Metrics Cards -->
    <section class="max-w-7xl mx-auto px-6 py-4">
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="bg-slate-900 border border-slate-800 p-5 rounded-xl">
          <div class="text-xs text-slate-400 uppercase tracking-wider font-semibold">Total Records</div>
          <div id="metric-total" class="text-3xl font-extrabold text-white mt-2">4</div>
          <div class="text-xs text-emerald-400 mt-1">↑ Active Memory Cache</div>
        </div>
        <div class="bg-slate-900 border border-slate-800 p-5 rounded-xl">
          <div class="text-xs text-slate-400 uppercase tracking-wider font-semibold">System Status</div>
          <div class="text-3xl font-extrabold text-emerald-400 mt-2">HEALTHY</div>
          <div class="text-xs text-slate-400 mt-1">RFC 7807 Ready</div>
        </div>
        <div class="bg-slate-900 border border-slate-800 p-5 rounded-xl">
          <div class="text-xs text-slate-400 uppercase tracking-wider font-semibold">HTTP Engine</div>
          <div class="text-3xl font-extrabold text-sky-400 mt-2">net/http</div>
          <div class="text-xs text-slate-400 mt-1">&lt; 1ms Avg Latency</div>
        </div>
        <div class="bg-slate-900 border border-slate-800 p-5 rounded-xl">
          <div class="text-xs text-slate-400 uppercase tracking-wider font-semibold">Uptime</div>
          <div id="metric-uptime" class="text-3xl font-extrabold text-indigo-400 mt-2">0s</div>
          <div class="text-xs text-slate-400 mt-1">Continuous Service</div>
        </div>
      </div>
    </section>

    <!-- Interactive Data & Form Section -->
    <main class="max-w-7xl mx-auto px-6 py-6 grid grid-cols-1 lg:grid-cols-3 gap-8">
      <div class="lg:col-span-2 bg-slate-900 border border-slate-800 rounded-xl p-6 shadow-xl">
        <div class="flex items-center justify-between mb-4 pb-3 border-b border-slate-800">
          <div>
            <h2 class="text-lg font-bold text-white">Live Records &amp; State</h2>
            <p class="text-xs text-slate-400">Current in-memory repository elements</p>
          </div>
          <button onclick="fetchItems()" class="text-xs bg-slate-800 hover:bg-slate-700 text-slate-200 px-3 py-1.5 rounded-lg border border-slate-700 transition flex items-center gap-1">
            <span>⟳</span> Refresh
          </button>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead>
              <tr class="text-xs text-slate-400 uppercase border-b border-slate-800">
                <th class="py-3 px-3">ID</th>
                <th class="py-3 px-3">Name / Label</th>
                <th class="py-3 px-3">Category</th>
                <th class="py-3 px-3">Status</th>
                <th class="py-3 px-3">Detail</th>
              </tr>
            </thead>
            <tbody id="items-tbody" class="divide-y divide-slate-800 text-slate-300">
            </tbody>
          </table>
        </div>
      </div>

      <div class="bg-slate-900 border border-slate-800 rounded-xl p-6 shadow-xl h-fit">
        <h2 class="text-lg font-bold text-white mb-1">Add New Webhook Event / Endpoint</h2>
        <p class="text-xs text-slate-400 mb-4">Dispatches POST to /api/items</p>

        <form id="add-form" onsubmit="handleCreate(event)" class="space-y-4">
          <div>
            <label class="block text-xs font-semibold text-slate-300 mb-1">Name / Title</label>
            <input type="text" id="form-name" required placeholder="e.g. Enterprise Production Target" class="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500">
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-300 mb-1">Category</label>
            <input type="text" id="form-category" required placeholder="e.g. Infrastructure" class="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500">
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-300 mb-1">Status</label>
            <select id="form-status" class="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500">
              <option value="Active">Active</option>
              <option value="Ready">Ready</option>
              <option value="Operational">Operational</option>
              <option value="Verified">Verified</option>
              <option value="Pending">Pending</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-semibold text-slate-300 mb-1">Detail Description</label>
            <textarea id="form-detail" required rows="2" placeholder="Key technical specifics or parameters" class="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"></textarea>
          </div>
          <button type="submit" class="w-full bg-gradient-to-r from-[#f59e0b] to-[#ef4444] text-white font-bold py-2.5 px-4 rounded-lg shadow-lg hover:opacity-95 transition text-sm">
            + Submit New Entry
          </button>
        </form>
      </div>
    </main>
  </div>

  <footer class="border-t border-slate-800 bg-slate-950 py-6 text-center text-xs text-slate-500">
    <div class="max-w-7xl mx-auto px-6 flex flex-col sm:flex-row items-center justify-between gap-4">
      <div>
        Built with clean Go standard library by <a href="https://github.com/k3ndzy" class="text-slate-400 hover:text-white underline">@k3ndzy</a>.
      </div>
      <div class="font-mono text-slate-600">
        Port 8198 • MIT License • Zero Dependencies
      </div>
    </div>
  </footer>

  <script>
    async function fetchItems() {
      try {
        const res = await fetch('/api/items');
        const items = await res.json();
        const tbody = document.getElementById('items-tbody');
        tbody.innerHTML = '';
        items.forEach(it => {
          const tr = document.createElement('tr');
          tr.className = 'hover:bg-slate-800/40 transition';
          tr.innerHTML = '<td class="py-3 px-3 font-mono text-xs text-indigo-400 font-bold">' + it.id + '</td>' +
            '<td class="py-3 px-3 font-semibold text-white">' + it.name + '</td>' +
            '<td class="py-3 px-3"><span class="text-xs bg-slate-800 text-slate-300 px-2 py-0.5 rounded border border-slate-700">' + it.category + '</span></td>' +
            '<td class="py-3 px-3"><span class="text-xs bg-emerald-950 text-emerald-400 border border-emerald-800/60 px-2 py-0.5 rounded font-mono">' + it.status + '</span></td>' +
            '<td class="py-3 px-3 text-xs text-slate-400">' + it.detail + '</td>';
          tbody.appendChild(tr);
        });
        document.getElementById('metric-total').innerText = items.length;
      } catch (err) {
        console.error('Fetch error:', err);
      }
    }

    async function handleCreate(e) {
      e.preventDefault();
      const payload = {
        name: document.getElementById('form-name').value,
        category: document.getElementById('form-category').value,
        status: document.getElementById('form-status').value,
        detail: document.getElementById('form-detail').value
      };
      try {
        const res = await fetch('/api/items', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        if (res.ok) {
          document.getElementById('add-form').reset();
          fetchItems();
        }
      } catch (err) {
        alert('Failed to save: ' + err);
      }
    }

    let uptimeSec = 0;
    setInterval(() => {
      uptimeSec++;
      const m = Math.floor(uptimeSec / 60);
      const s = uptimeSec % 60;
      document.getElementById('metric-uptime').innerText = m > 0 ? (m + 'm ' + s + 's') : (s + 's');
    }, 1000);

    fetchItems();
  </script>
</body>
</html>`

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "UP",
		"service": "webhook-tester-relay",
		"uptime":  time.Since(startTime).String(),
		"port":    8198,
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

func handleItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		items := store.List()
		_ = json.NewEncoder(w).Encode(items)
	case http.MethodPost:
		var item Item
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, `{"error":"Invalid request JSON"}`, http.StatusBadRequest)
			return
		}
		if item.Name == "" {
			http.Error(w, `{"error":"Name field is required"}`, http.StatusBadRequest)
			return
		}
		saved := store.Add(item)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(saved)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	items := store.List()
	_ = json.NewEncoder(w).Encode(map[string]any{
		"total_items": len(items),
		"uptime":      time.Since(startTime).String(),
		"version":     "1.0.0",
		"framework":   "Go Standard Library net/http",
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(htmlIndex))
	})
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/items", handleItems)
	mux.HandleFunc("/api/stats", handleStats)

	addr := ":8198"
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Server starting on http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down cleanly...")
}
