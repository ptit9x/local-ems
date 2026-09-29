# Bản thiết kế hệ thống Local EMS (Energy Management System)

## Goal
Xây dựng hệ thống Local EMS bằng Go (Golang) cho BESS 5MWh. Hệ thống hỗ trợ giả lập 100% phần cứng (Simulator), giao tiếp Modbus/OCPP, hoạt động độc lập không cần mạng (Store-and-Forward + Local UI), và tích hợp các thuật toán điều khiển năng lượng chuyên sâu theo kiến trúc OpenEMS.

## Kiến trúc phần mềm (Clean Architecture + Constraint-based Controller Chain)

### 1. File Structure
```text
/cmd
  /ems-core             # Main process (entrypoint)
/internal
  /config               # YAML/SQLite Config loader
  /simulator            # GIẢ LẬP: Khởi tạo Fake Modbus Servers (Meters, BMS, PCS)
  /modbus               # Modbus Client (TCP/RTU) kết nối phần cứng
  /devices              # Adapter phần cứng (Meter, BMS, PCS)
  /ocpp                 # Xử lý Websockets cho EV Chargers
  /channel              # Channel system: safe concurrent read/write với Process Image snapshot mỗi cycle
  /engine               # BỘ NÃO QUYẾT ĐỊNH
     /cycle             # Cycle Manager: vòng lặp chính 1s (đọc → xử lý → ghi)
     /scheduler         # Scheduler: quyết định thứ tự chạy controllers (FixedOrder)
     /resolver          # ESS Power Resolver: tổng hợp constraints → setpoint cuối cùng ghi xuống PCS
     /controllers
        /limit_discharge      # Chống xả kiệt pin (configurable: minSoc, forceChargeSoc)
        /sell_to_grid_limit   # Giới hạn đẩy điện ngược lên lưới (configurable: maxSellToGridPower)
        /peak_shaving         # Cạo đỉnh tải (Bù pin khi tải văn phòng quá cao)
        /time_of_use          # Sạc đêm giá rẻ, xả ngày giá đắt
        /balancing            # Cân bằng tự tiêu thụ (Self-consumption)
  /health               # Watchdog & health check cho process
  /sync                 # Lưu InfluxDB nội bộ & Đồng bộ MQTT khi có mạng
  /ui                   # Local HMI (REST API / Websockets)
/web                    # Mã nguồn Frontend HMI
/test
  /integration          # Integration tests cho Controller Chain
  /simulation           # Replay data → verify output
```

### 2. Mô hình thuật toán (Kế thừa từ OpenEMS — Constraint Accumulation)

Thay vì một hàm "If-else" khổng lồ hay kiểu "override tuyệt đối", chúng ta sử dụng kiến trúc **Constraint-based Controller Chain** giống OpenEMS:

**Nguyên lý hoạt động:**
1. **Scheduler** quyết định thứ tự chạy controllers (priority cao → chạy trước).
2. **Mỗi controller đều được chạy**, và đặt **constraint** lên ESS (ví dụ: `MaxDischargePower ≤ 0`, `MaxChargePower ≤ 2000W`).
3. Controller có priority cao đặt constraint trước — constraint của nó sẽ **chặt hơn** và được ưu tiên giữ lại.
4. Sau khi tất cả controllers chạy xong, **ESS Power Resolver** tổng hợp toàn bộ constraints để tìm ra setpoint khả thi cuối cùng ghi xuống PCS.

```go
// Mỗi controller thêm constraints, KHÔNG override trực tiếp
type Constraint struct {
    MaxDischargePower *int   // Giới hạn công suất xả (W), nil = không giới hạn
    MaxChargePower    *int   // Giới hạn công suất sạc (W), nil = không giới hạn
    ForcePower        *int   // Ép buộc setpoint (chỉ dùng cho trường hợp force charge)
    Source            string // Tên controller tạo constraint
}

// Resolver chọn giá trị MIN của tất cả MaxDischarge, MIN của tất cả MaxCharge
// → đảm bảo constraint chặt nhất luôn thắng
```

**Thứ tự chạy Controllers (Scheduler FixedOrder):**
*   **Priority 1 (Bảo vệ phần cứng):** `limit_discharge`. SOC < `minSoc` (default: 15%) → constraint `MaxDischargePower = 0`. SOC < `forceChargeSoc` (default: 10%) → constraint `ForcePower = chargeRate`. Các controllers sau vẫn chạy nhưng constraints sẽ bị resolver ghi đè bởi constraint chặt hơn.
*   **Priority 2 (Tuân thủ điện lực):** `sell_to_grid_limit`. Nếu Grid Meter báo công suất sell-to-grid vượt `maxSellToGridPower` (default: 0W cho zero-export) → constraint tăng charge hoặc giảm discharge tương ứng.
*   **Priority 3 (Tối ưu chi phí):** `peak_shaving` — constraint xả pin khi tải vượt ngưỡng cắt đỉnh. `time_of_use` — constraint sạc trong khung giờ giá rẻ, xả trong khung giờ giá đắt.
*   **Priority 4 (Mặc định):** `balancing`. Đảm bảo điện mặt trời ưu tiên sạc pin, thiếu thì dùng lưới.

### 3. Vòng lặp Cycle (Cycle Manager)

Mỗi cycle (~1 giây), Cycle Manager thực hiện tuần tự:
```text
┌─────────────────────────────────────────────────┐
│ 1. BEFORE_PROCESS_IMAGE                         │
│    → Snapshot tất cả channel values             │
│                                                 │
│ 2. AFTER_PROCESS_IMAGE                          │
│    → Tính toán Sum (tổng công suất, SOC...)     │
│                                                 │
│ 3. BEFORE_CONTROLLERS                           │
│    → Reset constraint list                      │
│                                                 │
│ 4. EXECUTE_CONTROLLERS                          │
│    → Scheduler gọi từng controller theo thứ tự  │
│    → Mỗi controller đọc process image, thêm    │
│      constraints vào resolver                   │
│                                                 │
│ 5. AFTER_CONTROLLERS                            │
│    → Resolver tổng hợp constraints → setpoint   │
│                                                 │
│ 6. EXECUTE_WRITE                                │
│    → Ghi setpoint xuống PCS qua Modbus          │
│                                                 │
│ 7. AFTER_WRITE                                  │
│    → Log, ghi InfluxDB, emit events             │
└─────────────────────────────────────────────────┘
```

## Tasks Roadmap (Các Giai Đoạn Phát Triển)

### Phase 0: Spike — Validate kiến trúc Constraint Resolver (1-2 ngày)
- [x] Task 0.1: PoC nhỏ gồm 2 fake controllers + constraint resolver. Chạy trên hardcoded data, verify constraint chặt nhất luôn thắng.
- [x] Task 0.2: Viết unit test cho resolver: nhiều constraints chồng chéo → output đúng.

### Phase 1: Nền tảng & Giả lập (Simulator)
- [x] Task 1.1: Khởi tạo project Go, setup cấu trúc thư mục, `go.mod`.
- [ ] Task 1.2: Viết module `/internal/channel` — Channel system với Process Image (snapshot safe cho concurrent read).
- [x] Task 1.3: Viết module `/internal/simulator` tạo Fake Modbus Server cho Meter, BMS, PCS.
- [x] Task 1.4: Dựng script sinh dữ liệu giả lập: Hình sin cho Điện mặt trời, Random cho Tải văn phòng, SOC ramp cho BMS.
- [x] Task 1.5: Viết unit tests cho simulator (data ranges hợp lệ, Modbus registers đúng format).

### Phase 2: Kết nối thiết bị & Giao tiếp cốt lõi
- [x] Task 2.1: Viết Modbus Client đọc thông số từ Fake Meter (Lưới & Mặt trời).
- [x] Task 2.2: Viết Adapter đọc trạng thái Fake BMS (SOC, Nhiệt độ, Cell voltage limits).
- [x] Task 2.3: Viết Adapter ghi lệnh sạc/xả xuống Fake PCS (Active/Reactive Power setpoint).
- [x] Task 2.4: Viết unit tests cho mỗi adapter (mock Modbus server → verify parsed values).

### Phase 3: Thuật toán Điều khiển (Decision Engine)
- [x] Task 3.1: Code `Controller` interface + `Constraint` struct. Viết Cycle Manager (vòng lặp 1s).
- [x] Task 3.2: Code `Scheduler` (FixedOrder) — nhận danh sách controller IDs, gọi theo thứ tự.
- [x] Task 3.3: Code `ESS Power Resolver` — nhận []Constraint → tính setpoint cuối cùng.
- [x] Task 3.4: Implement `limit_discharge` (configurable `minSoc`, `forceChargeSoc`). + unit tests.
- [x] Task 3.5: Implement `sell_to_grid_limit` (configurable `maxSellToGridPower`, default 0). + unit tests.
- [x] Task 3.6: Implement `peak_shaving` (configurable `peakThreshold`). + unit tests.
- [x] Task 3.7: Implement `time_of_use` (configurable schedule giá điện theo giờ). + unit tests.
- [x] Task 3.8: Implement `balancing` (self-consumption optimization). + unit tests.
- [x] Task 3.9: Integration test: lắp ráp tất cả controllers → Scheduler → Resolver → chạy trên simulator data.

### Phase 4: Local UI & Offline Sync
- [x] Task 4.1: Nhúng SQLite, thiết lập dịch vụ ghi đệm (Buffer). *(InfluxDB → SQLite cho MVP)*
- [ ] Task 4.2: Code `sync` module đẩy MQTT khi Online (Store-and-Forward). *(deferred to post-MVP)*
- [x] Task 4.3: Dựng REST API + Websocket cho Local Web UI (Hiển thị luồng điện thời gian thực).
- [x] Task 4.4: Code `/internal/health` — watchdog heartbeat, restart policy khi controller panic.

### Phase 5: EV Charger & Multi-device (Q&A Update)
- [x] Task 5.1: Device Plugin Architecture — `Device`, `ControllableDevice` interfaces, `DeviceRegistry`
- [x] Task 5.2: EV Charger Adapter — Modbus read/write for EV charger (status, power, start/stop)
- [x] Task 5.3: Multi-brand PCS Register Map — PCSProfile registry (generic/sungrow/sinexcel)
- [x] Task 5.4: EV Charging Controller — DLM constraint controller (P3 priority)
- [x] Task 5.5: Configurable Device Topology — YAML config for site, bess_groups, topology
- [x] Task 5.6: EV Charger Dashboard — Widget, device chips, detail page on dashboard
- [x] Task 5.7: Aggregator EV Charger — EVChargerState, UpdateEVCharger, EV counts/totals
- [x] Task 5.8: Distributor master/individual mode — BESS group command mode support
- [x] Task 5.9: SQLite WAL mode + 7-day auto-retention
- [ ] Task 5.10: EV active control — auto WriteMaxPower from DLM controller

### Phase 6: Real Hardware & Multi-rate Polling (Planned)
- [ ] Task 6.1: Multi-rate poller (200ms/1s/5s goroutines)
- [ ] Task 6.2: Remove simulation-only guard
- [ ] Task 6.3: Robust data integrity (WAL + retry)

### Phase 7: Cloud Sync (Planned)
- [ ] Task 7.1: MQTT publisher module
- [ ] Task 7.2: Store-and-forward buffer
- [ ] Task 7.3: Cloud failover signal for EV Charger

### Phase 8: Production Hardening (Planned)
- [ ] Task 8.1: Config hot-reload
- [ ] Task 8.2: Per-device health monitoring

## Done When
- [x] `go run ./cmd/ems-core --simulate` khởi chạy thành công, simulator tạo dữ liệu sin/random liên tục.
- [x] **Zero-export:** Grid Meter fake = -500W (sell) → sell_to_grid_limit constraint kích hoạt → PCS tăng charge → grid power ≥ 0W trong ≤ 3 cycles.
- [x] **Chống kiệt pin:** SOC giảm đến 15% → discharge constraint = 0. SOC giảm đến 10% → force charge bật.
- [x] **Peak shaving:** Tải vượt peak threshold → PCS discharge bù, tải nhìn từ grid giảm dưới threshold.
- [x] **Time-of-use:** Trong khung giờ giá rẻ → tự động charge. Trong khung giờ giá đắt → tự động discharge.
- [x] **Constraint conflict:** Khi peak_shaving muốn discharge nhưng SOC < minSoc → limit_discharge constraint thắng, discharge = 0.
- [x] SQLite local ghi data liên tục mỗi cycle. *(MQTT sync deferred to post-MVP)*
- [x] Tất cả unit tests pass. Integration test coverage > 80%.
- [x] EV Charger visible on dashboard with real-time status updates
- [x] Constraint-based EV DLM controller in scheduler
- [ ] Cloud sync via MQTT every 5 minutes
- [ ] Multi-rate hardware polling (200ms fast meter)

