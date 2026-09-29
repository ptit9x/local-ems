# Kế hoạch phát triển Cloud ERP (VPP / Energy Dashboard)

## Goal
Xây dựng một hệ thống Cloud ERP tập trung để giám sát, quản lý từ xa (Fleet Management), và tính cước (Billing) cho hàng loạt trạm Local EMS (BESS 5MWh) và trụ sạc xe điện (EV Chargers) theo thời gian thực.

## Technology Stack (Kiến trúc Cloud)

### 1. Tầng Thu thập dữ liệu (IoT Ingestion)
*   **MQTT Broker:** `EMQX` (Khuyên dùng) - Chịu tải hàng triệu kết nối từ các Local EMS, xử lý định tuyến bản tin (Rules Engine) đẩy dữ liệu thẳng vào Database.

### 2. Tầng Cơ sở dữ liệu (Databases)
*   **Time-series DB:** `TimescaleDB` (PostgreSQL extension) - Lưu trữ dữ liệu viễn trắc (Telemetry) thay đổi liên tục: Nhiệt độ, Công suất lưới, % Pin (SOC).
*   **Relational DB:** `PostgreSQL` - Lưu trữ dữ liệu nghiệp vụ: Thông tin khách hàng, Trạm sạc, Lịch sử hóa đơn, Phân quyền (RBAC).
*   **Cache & Real-time:** `Redis` - Lưu đệm trạng thái thiết bị (Online/Offline) và xử lý Pub/Sub cho Websocket.

### 3. Tầng Backend (API & Xử lý nghiệp vụ)
*   **Framework:** `Node.js` kết hợp `NestJS` (TypeScript).
*   **Chức năng chính:**
    *   Cung cấp REST/GraphQL API cho Frontend.
    *   Xử lý logic tài chính, đối soát, xuất hóa đơn (Billing).
    *   Websocket Server nhận dữ liệu Real-time đẩy lên màn hình ERP.
    *   Giao tiếp chuẩn OCPP Central System (phục vụ kết nối dự phòng từ các trụ sạc EV).

### 4. Tầng Frontend (Web ERP / Dashboard)
*   **Core:** `React` + `Vite` (Client-Side Rendering) - Tốc độ build siêu nhanh, deploy cực kỳ đơn giản qua AWS S3 hoặc Nginx (Không cần Server-side rendering như Next.js).
*   **State Management:** `Zustand` hoặc `Redux Toolkit`.
*   **UI Library:** `Ant Design (AntD)` hoặc `MUI` - Phù hợp dựng các cấu trúc bảng (Tables), form cấu hình phức tạp của ERP.
*   **Charting:** `Apache ECharts` - Vẽ đồ thị năng lượng, luồng điện mượt mà kể cả với lượng data khổng lồ (vài chục ngàn điểm).

---

## Các Module & Tasks chính (Giai đoạn 1 - MVP Cloud)

### Module 1: Hạ tầng Cloud & Database (DevOps)
- [ ] Task 1.1: Khởi tạo Docker/Kubernetes cluster. Deploy EMQX broker. → Verify: Test bằng MQTT Client kết nối thành công.
- [ ] Task 1.2: Thiết lập hệ cơ sở dữ liệu TimescaleDB, PostgreSQL và Redis. → Verify: Dùng công cụ (như DataGrip) kết nối thành công tới các DB.

### Module 2: Backend Core (NestJS)
- [ ] Task 2.1: Setup base project NestJS, cấu hình TypeORM/Prisma kết nối PostgreSQL và Timescale. → Verify: API `/api/health` trả về 200 OK.
- [ ] Task 2.2: Viết IoT Service, đăng ký nhận bản tin từ EMQX qua Webhook hoặc giao thức MQTT. → Verify: Log ra màn hình khi Local EMS gửi dữ liệu giả lập.
- [ ] Task 2.3: Viết logic insert dữ liệu Telemetry (Dòng, Áp, SOC) vào TimescaleDB. → Verify: Truy vấn DB thấy data được lưu trữ thành công theo dạng chuỗi thời gian.

### Module 3: Frontend Web ERP (React + Vite)
- [ ] Task 3.1: Khởi tạo project (`npm create vite@latest`), cài đặt Ant Design, React Router. → Verify: Chạy `npm run dev` lên giao diện mặc định.
- [ ] Task 3.2: Dựng Layout chính của ERP (Sidebar, Header, Menu bao gồm: Dashboard Tổng, Quản lý Trạm sạc, Báo cáo). → Verify: UI hiển thị chuẩn theo Layout thiết kế.
- [ ] Task 3.3: Dựng trang **Multi-site Dashboard** tích hợp ECharts, vẽ biểu đồ SOC% theo thời gian thực (nhận data qua Websocket từ NestJS). → Verify: Biểu đồ giật line và thay đổi ngay khi có data mới.

### Module 4: Billing & Cấu hình từ xa (Business Features)
- [ ] Task 4.1: Xây dựng cấu trúc bảng và API quản lý Biểu giá điện (Time-of-use tariffs). → Verify: User tạo được các mốc giá (Cao điểm, Thấp điểm) qua giao diện.
- [ ] Task 4.2: Viết luồng đồng bộ giá điện (Remote config) xuống Local EMS qua MQTT. → Verify: Local EMS log ra thông báo nhận được giá điện mới.

## Done When
- [ ] Toàn bộ hạ tầng Cloud được đưa vào hoạt động (Local).
- [ ] Local EMS (từ kế hoạch trước) publish dữ liệu MQTT lên EMQX thành công và lưu vào DB.
- [ ] Ban Giám đốc/Kỹ sư vận hành đăng nhập vào Web ERP, theo dõi được hệ thống BESS 5MWh đang hoạt động với biểu đồ SOC% và Energy Flow chuẩn xác.
