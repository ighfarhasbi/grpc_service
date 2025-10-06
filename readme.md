<div>
  <img style="width: 100%" src="https://capsule-render.vercel.app/api?type=waving&height=80&section=header&reversal=false&fontSize=70&fontColor=FFFFFF&fontAlign=50&fontAlignY=50&stroke=-&descSize=20&descAlign=50&descAlignY=50&color=gradient" />
</div>

<h1 align="left">Order & Inventory Management (gRPC)</h1>

---

## 📑 Index
- ERD  
- Architecture design  
- Asumsi dan alur  
- Run and setup  

---

## 🧩 ERD

### 🗃️ Inventory_db

<div align="center">
  <img style="width: 100%" src="https://github.com/user-attachments/assets/b5598162-a809-41fd-a0b8-0e35ecd6a18b" alt="Inventory ERD" />
</div>

---

### 📦 Order_db

<div align="center">
  <img style="width: 100%" src="https://github.com/user-attachments/assets/e04cfef0-24ee-42bb-8ef7-8a7d28caec34" alt="Order ERD" />
</div>

---

## 🏗️ Architecture Design

<div align="center">
  <img style="width: 100%" src="https://github.com/user-attachments/assets/2afc164a-7b32-4fe9-828f-6166226469a0" alt="Architecture Design" />
</div>

---

## 💡 Asumsi dan Alur

### 🧠 Asumsi Dasar
1. **Service terpisah sepenuhnya (microservice-based)**  
   - Setiap service berjalan di container berbeda dalam satu jaringan Docker Compose.  
   - Komunikasi antar service menggunakan **gRPC**   
   - Service tidak saling mengakses database secara langsung; interaksi dilakukan hanya melalui gRPC call.

2. **Database terpisah per domain**  
   - `order_db` hanya digunakan oleh **Order Service** untuk menyimpan data transaksi pemesanan.  
   - `inventory_db` hanya digunakan oleh **Inventory Service** untuk menyimpan data stok barang. 

3. **ACID Transaction dan Rollback Handling**  
   - Masing-masing service memiliki mekanisme transaksi internal (ACID).  
   - Jika proses di salah satu service gagal (misal reservasi stok gagal), maka service terkait akan melakukan rollback untuk menjaga konsistensi data.

4. **Client Layer sebagai REST Gateway**  
   - Client service (menggunakan Echo Framework) bertugas menerima request HTTP dari user/client eksternal.  
   - Client tidak menyimpan data; hanya meneruskan request ke Order Service melalui gRPC.

---

### 🔄 Alur Proses

#### 1. Create Order
1. Client (REST API) mengirimkan request `POST /order` ke **Client Service** dengan payload berisi `user_id` dan daftar `list_items`.
2. Client Service meneruskan request tersebut ke **Order Service** melalui **gRPC call**.
3. Order Service:
   - Menghitung total harga dari list item.  
   - Melakukan **gRPC call ke Inventory Service** untuk melakukan *stock reservation*.
4. Inventory Service:
   - Mengecek ketersediaan stok untuk setiap produk.  
   - Jika stok tersedia, maka stok akan di-*reserve* (dikurangi sementara).  
   - Jika stok tidak cukup, Inventory Service mengembalikan error dan tidak melakukan perubahan stok.
5. Jika semua stok berhasil di-*reserve*, Order Service akan:
   - Menyimpan data order ke `order_db`.  
   - Mengembalikan response sukses ke Client.
6. Jika salah satu proses gagal, maka:
   - Order Service melakukan rollback (menghapus data order yang sudah dibuat).  
   - Inventory Service juga melakukan rollback untuk *release stock* yang sebelumnya direserve.

---

#### 2. Cancel Order
1. Client mengirimkan request `PUT /cancel` dengan parameter `order_id`.  
2. Client Service meneruskan request tersebut ke Order Service melalui gRPC.
3. Order Service memeriksa order terkait di `order_db`, kemudian:
   - Mengubah status order menjadi *canceled*.  
   - Memanggil Inventory Service untuk melakukan *release stock* (mengembalikan stok yang sempat direserve).
4. Inventory Service menambahkan kembali stok barang di `inventory_db`.  

---

## ⚙️ Run dan Setup
1. Clone repository ini
2. Tambahkan file `.env.client`; `.env.inventory`; `.env.order`. sudah ada example masing-masing di repo ini
3. app ini sudah dengan swagger untuk hit API. saat ini host untuk akses swagger di set ke `172.16.148.101` (vps), bisa dilihat di file main.go yang ada di `./client/main.go`
   3a. jika dijalankan di localhost maka perlu penyesuaian :
     - ubah jadi localhost di baris ke 16 file `./client/main.go` (port biarkan)
4. jika semua sudah sesuai, lalu `docker compose up -d`
5. setelah container jalan semua, tinggal akses swagger ke `http://localhost:8083/swagger///index.html`
6. bisa langsung hit API order
7. jika butuh id product, sudah ada di data seed folder `migrations/inventory/0002_seed_products.up.sql`
8. untuk user_id (harus uuid), masih random uuid karena belum ada hubungan ke user service. Bisa generate di `https://www.uuidgenerator.net/`

### Docker Images
| Service | Image |
|----------|--------|
| Inventory Service | `ighfarhasbi/grpc_service:inventory` |
| Order Service | `ighfarhasbi/grpc_service:order` |



<div>
  <img style="width: 100%" src="https://capsule-render.vercel.app/api?type=waving&height=80&section=footer&reversal=false&fontSize=70&fontColor=FFFFFF&fontAlign=50&fontAlignY=50&stroke=-&descSize=20&descAlign=50&descAlignY=50&color=gradient" />
</div>

