
# Go ile Dağıtık Sistemler: Consistent Hashing ve Toy DynamoDB

Bu proje, dağıtık sistemlerde veri dağıtımı, yük dengeleme ve veri saklama stratejilerinin Go dili ile sıfırdan inşasını içerir.

Proje, basit bir modülo hashing mantığından başlayıp; **Replication**, **Sharding**, **Thread-Safety**, **Coordinator** ve **Quorum Consensus** desenlerini kullanan, üretim standartlarına yakın bir **Distributed Key-Value Store** (Toy DynamoDB) mimarisine evrilen süreci modeller.

Amaç sadece kod yazmak değil; "Race Condition", "Deadlock", "Goroutine Leak" ve "Data Hotspots" gibi mühendislik problemlerini yerinde tespit edip çözmektir.

## Proje Yapısı

Proje, mimari evrimin her aşamasını temsil eden modüllerden oluşur:

```text
.
├── consistent_hashing                  # V2: Temel Halka (Ring) Yapısı
├── consistent_virtual_hashing          # V3: Sanal Düğümler (Virtual Nodes)
├── consistent_virtual_hashing_replicated # V4: Replikasyon Mantığı (Simulation)
├── modular_hashing                     # V1: Klasik Modülo Yöntemi
└── toy_dynamodb                        # V6: Quorum, Concurrency & Storage Engine
    ├── node                            # Fiziksel Depolama Birimi (Storage Engine)
    └── ring                            # Yönlendirme ve Orkestrasyon (Coordinator)

```

## Modüller ve Teknik Detaylar

### 1. modular_hashing (Klasik Yaklaşım)

`hash(key) % sunucu_sayısı` formülü. Sunucu sayısı değişince %100'e yakın "Cache Miss" yaşanır. Ölçeklenebilir değildir.

### 2. consistent_hashing (Halka Topolojisi)

Sunucular bir çembere yerleştirilir. Node ekleme/çıkarma maliyeti minimize edilir ancak veri dağılımı dengesizdir.

### 3. consistent_virtual_hashing (O(N) Optimizasyonu)

Her fiziksel sunucu yüzlerce "Sanal Node" ile temsil edilir. Yük dağılımı mükemmel dengeye (%1 varyasyon) ulaşır.

### 4. consistent_virtual_hashing_replicated (Replikasyon Mantığı)

Verinin `N=3` kopyasının mantıksal olarak hangi node'lara gideceğini hesaplar. Ancak veriyi fiziksel olarak saklamaz, sadece routing simülasyonudur.

---

### 5. toy_dynamodb (V6 - Quorum Consensus & Concurrency)

Bu modül, projenin **"Concurrent Distributed Database"** haline gelmiş sürümüdür. Veri tutarlılığı ve erişilebilirlik, istemci tarafından ayarlanabilir.

#### Mimari Özellikler:

* **Tunable Consistency (Ayarlanabilir Tutarlılık):**
    * **N (Replication Count):** Verinin toplam kaç kopyası olacağı.
    * **W (Write Quorum):** Yazma işleminin başarılı sayılması için gereken minimum onay sayısı.
    * **R (Read Quorum):** Okuma işleminin başarılı sayılması için gereken minimum cevap sayısı.
    * Kullanıcı `Put(k, v, w)` ve `Get(k, r)` çağrılarıyla **Hız vs. Tutarlılık** dengesini kendi yönetir.


* **Concurrent Execution (Fan-Out / Fan-In):**
    * V5'teki sıralı (sequential) döngüler terk edilmiştir.
    * İstek anında `N` adet Goroutine paralel olarak ateşlenir ("Fan-Out").
    * Sonuçlar `Buffered Channel` üzerinden toplanır ("Fan-In").
    * **Latency Hiding:** Sistem en yavaş sunucuyu beklemez. İstenen `W` veya `R` sayısına ulaşıldığı an ("Early Exit") cevap dönülür.


* **Thread-Safety & Deadlock Prevention:**
    * **Granular Locking:** `Ring` ve `Node` yapıları kendi `RWMutex`lerini yönetir.
    * **Channel Management:** Goroutine sızıntısını (Leak) önlemek için kanallar `N` kapasiteli (Fire-and-Forget) tasarlanmıştır.
    * **Double-Check Locking:** TOCTOU hatalarını önlemek için node ekleme sırasında çift kontrol mekanizması uygulanır.



## Nasıl Çalıştırılır?

En güncel Distributed KV Store versiyonunu çalıştırmak için:

```bash
cd toy_dynamodb
go run main.go
```

### Örnek Kullanım (V6)

Aşağıdaki örnekte 6 sunuculu bir küme oluşturulmuş, Replikasyon (N) 5 olarak ayarlanmış, ancak Yazma (W) ve Okuma (R) için 3 onay yeterli görülmüştür.

```go
func main() {
    // Ring başlatılır ve Replikasyon Sayısı (N) 5 olarak ayarlanır
    ring := Ring.Ring{ReplicaCount: 5}
    ring.Init()

    // Node'lar güvenli bir şekilde eklenir
    nodes := []string{"foo", "bar", "bazz", "zapp", "zucc", "bars"}
    for _, n := range nodes {
        if err := ring.AddNode(n); err != nil {
            log.Fatalf("%v", err)
        }
    }

    // W=3 (En az 3 sunucu "Yazdım" desin, gerisini bekleme)
    err := ring.Put("Mahmut", "Ozer", 3)
    if err != nil {
         log.Fatalf("Write failed: %v", err)
    }

    // R=3 (En az 3 sunucudan veri getir)
    vals, err := ring.Get("Mahmut", 3)

    if err == nil {
        fmt.Println(vals)
    } else {
        log.Fatalf("Read failed: %v", err)
    }
}
```
**Beklenen Çıktı:**
```text
map[bar:Ozer bazz:Ozer foo:Ozer]
```

*(Not: Çıktı bir Map olduğu için sıralama değişebilir ve Quorum sağlandığı an dönüldüğü için sadece en hızlı 3 sunucunun cevabını içerir.)*