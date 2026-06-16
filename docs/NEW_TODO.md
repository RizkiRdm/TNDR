# NEW_TODO.md

## Project Direction

**TNDR is NOT an Enterprise AI Gateway.** 
Positioning baru: **"AI Cost & Observability Platform for Solo Developers"**. 

Fokus utama produk adalah Visibilitas Biaya, Kesederhanaan Integrasi (TUI/CLI first), dan Engineering Memory. TNDR berjalan lokal (satu *binary*), ringan, dan bertindak sebagai asisten finansial + *debugger* untuk developer yang menggunakan banyak LLM, bukan sekadar *router proxy*.

---

## Features To Add

### High Priority (Core Value - Quick & Medium Wins)

- [ ] **Hard Spend Kill-Switch:** Fitur untuk nge-blokir *request* otomatis jika limit budget harian/bulanan (misal $5/hari) tercapai. Menghindari tagihan jebol.
- [ ] **Offline Token Estimator:** CLI command (`tndr est "prompt" --model claude-3-haiku`) untuk menghitung estimasi biaya *sebelum* request dikirim, menggunakan kalkulasi *token* lokal.
- [ ] **Unified Local SQLite Logger:** Menyimpan semua *request*, *response*, *latency*, dan *cost* ke database lokal yang ringan tanpa butuh setup server terpisah.
- [ ] **TUI Cost Dashboard:** Tampilan terminal interaktif yang menunjukkan grafik ASCII pengeluaran harian, model mana yang paling menyedot biaya, dan *latency* rata-rata.
- [ ] **Smart Model Advisor (CLI):** Command `tndr suggest --task code_gen --budget low`, mengembalikan rekomendasi model terbaik (misal: Gemini Flash atau DeepSeek Coder) berdasarkan rasio harga/kualitas.
- [ ] **Prompt Playground (CLI/TUI):** Fitur untuk mengirim satu prompt ke 2-3 provider sekaligus secara paralel dan menampilkan perbandingan *cost*, *latency*, dan *output* berdampingan.

### Medium Priority (Enhancing the Solo Dev Workflow)

- [ ] **Cost-Aware Fallback Routing:** Jika model utama gagal, TNDR tidak asal lempar ke model lain, tapi mencari model *fallback* yang harganya sama atau lebih murah.
- [ ] **Local Prompt Caching (Semantic/Exact):** Menyimpan *response* dari *prompt* yang sering diulang (misal saat *testing* kode berulang-ulang) secara lokal untuk memotong biaya API 100%.
- [ ] **Prompt Replay CLI:** Command `tndr replay <request_id> --model new_model` untuk mengetes ulang prompt lama dengan model berbeda tanpa harus me-rebuild request dari awal.
- [ ] **Context Window Trim Warning:** Memberi peringatan (atau otomatis memotong *fluff*) jika *prompt* hampir menyentuh batas maksimum *context window* model yang dipilih.
- [ ] **.env Key Vault:** Manajemen API Keys sentral yang aman secara lokal, sehingga *user* tidak perlu *hardcode* key di setiap project aplikasi yang mereka bangun.
- [ ] **CLI Export to CSV/JSON:** Kemudahan mengekspor data metrik dan riwayat *prompt* untuk dianalisis lebih lanjut secara mandiri.

### Low Priority (Future Observability & Expansion)

- [ ] **Engineering Memory / Graphify Integration:** Memetakan relasi antara *prompt*, jenis *task*, dan performa model dalam bentuk *knowledge graph* untuk melihat tren penggunaan jangka panjang (`tndr trace`).
- [ ] **Anomaly Spend Alerts:** Notifikasi sistem (desktop notification/CLI warning) jika ada lonjakan penggunaan *token* yang tidak wajar dalam rentang 1 jam terakhir.
- [ ] **Auto-Prompt Compression:** Eksperimen agen lokal kecil (seperti model LLM lokal) yang meringkas *prompt* sebelum dikirim ke provider berbayar mahal untuk memangkas *input tokens*.

---

## Features To Modify

- [ ] **Basic Routing Configuration:** Ubah dari format konfigurasi YAML/JSON yang kompleks ala enterprise menjadi struktur yang sangat minimalis dan intuitif (contoh: cukup set `DEFAULT_MODEL` dan `FALLBACK_MODEL`).
- [ ] **TUI (Terminal User Interface):** Jangan *over-polish* UI. Modifikasi fokus TUI saat ini HANYA untuk menampilkan 3 hal krusial: Total Requests, Total Cost ($), dan Error Rates. Estetika belakangan.
- [ ] **Multi-provider Logic:** Modifikasi agar TNDR menggunakan satu *unified schema* (seperti format OpenAI) untuk *request* dan mem-parsingnya ke format spesifik (Anthropic, Gemini) di belakang layar secara mulus.

---

## Features To Remove

- [ ] **Distributed Rate Limiting (Redis):** *Overkill*. Hapus dependensi eksternal. Solo dev tidak butuh *rate limiting* antar ratusan *server*. Gunakan *in-memory counter* atau SQLite lokal.
- [ ] **Role-Based Access Control (RBAC):** *Remove*. TNDR dirancang untuk satu *user* lokal. Tidak perlu manajemen *user*, *admin*, atau JWT tokens yang rumit.
- [ ] **PII (Personally Identifiable Information) Redaction:** Sulit di-maintain dan membebani komputasi lokal. Solo dev biasanya sadar apa yang mereka kirim ke LLM.
- [ ] **Kubernetes/Helm Deployment Scripts:** Buang. Cukup sediakan satu *compiled binary* (Linux/macOS/Windows) dan satu *Dockerfile* standar.
- [ ] **Complex Load Balancing (Round Robin, Weighted):** Hapus. Solo dev jarang punya 5 API key OpenAI yang berbeda untuk di-*load balance*. *Fallback* sederhana sudah cukup.

---

## PLAN.md Audit

*Asumsi list di bawah adalah fitur tipikal Gateway yang sering direncanakan namun harus difilter ulang berdasarkan visi baru.*

| Item | Decision | Reason |
|--------|--------|--------|
| Multi-provider Unified API | KEEP | Esensial agar developer tidak perlu belajar banyak SDK. |
| Redis Distributed Cache | REMOVE | *Over-engineered*. Bikin susah di-setup. Ganti dengan SQLite lokal. |
| User Auth & Organizations | REMOVE | Target kita solo dev/indie hacker. Ini fitur enterprise B2B. |
| Basic Retry & Fallback | KEEP | Sangat berguna jika provider tiba-tiba RTO (*Request Timeout*). |
| Advanced Load Balancing | REMOVE | Tidak relevan untuk trafik skala kecil/menengah dari satu *developer*. |
| Streaming Support (SSE) | KEEP | Wajib ada untuk UX aplikasi yang sedang dibangun oleh *developer*. |
| Cost Tracking & Limits | MODIFY | Jadikan ini fitur utama, angkat posisinya dari sekadar "tambahan" menjadi layar utama di TUI. |

---

## V2 Roadmap (High Priority - Validasi Produk)

*Fokus: "Bikin LLM bill lo transparan dan gampang dikontrol."*

1.  **Core Proxy & Unified API:** Memastikan *routing* dasar (OpenAI format) ke provider populer (Gemini, Claude, OpenRouter) berjalan stabil termasuk *streaming*.
2.  **Cost Engine V1:** Intersepsi dan kalkulasi harga berdasarkan token *input/output* secara *real-time*.
3.  **Local Storage (SQLite):** Sistem *logging* untuk menyimpan riwayat *request* dan pengeluaran tanpa dependensi rumit.
4.  **TUI Minimalis:** Dashboard di terminal untuk melihat total pengeluaran hari ini dan bulan ini.
5.  **Hard Spend Limit:** *Kill-switch* otomatis untuk mencegah tagihan membengkak.

---

## V3 Roadmap (Setelah Ada Validasi Pengguna)

*Fokus: "Bikin lo pintar milih model dan eksperimen dengan murah."*

1.  **Smart Model Advisor:** Rekomendasi otomatis model termurah untuk spesifik tugas (berdasarkan rasio biaya/konteks).
2.  **Benchmark Playground:** CLI untuk tes A/B satu *prompt* ke beberapa model untuk melihat perbedaan *cost*, *latency*, dan kualitas.
3.  **Local Exact Caching:** Mencegah panggilan API berulang untuk *prompt* yang 100% sama (sangat menghemat biaya saat *debugging* kode).
4.  **Prompt Replay:** Fitur untuk mengulang *request* lama dari *history* dengan model yang berbeda.

---

## Future Ideas (Optional / Eksplorasi)

- **Semantic Caching:** Menggunakan *local embedding* (misal via arsitektur RAG ringan) untuk mengenali *prompt* yang "mirip" dan mengembalikan hasil dari *cache*, bukan hanya yang *exact match*.
- **Engineering Memory Graph:** Membangun *knowledge graph* dari log interaksi AI untuk memberikan *insight* dalam tentang kebiasaan *prompting* *developer* dan mendeteksi anomali biaya jangka panjang.
- **Auto-Prompt Trimming:** Agen heuristik yang otomatis menghapus spasi, komentar, atau baris tidak relevan dari kode sebelum dikirim ke LLM untuk menekan jumlah *input token*.