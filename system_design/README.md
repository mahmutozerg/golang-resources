
# Go ile Dağıtık Sistemler: Consistent Hashing ve Toy DynamoDB

Bu proje, dağıtık sistemlerde veri dağıtımı, yük dengeleme ve veri saklama stratejilerinin Go dili ile sıfırdan inşasını içerir.

Proje, basit bir modülo hashing mantığından başlayıp; **Replication**, **Sharding**, **Thread-Safety** ve **Coordinator** desenlerini kullanan, üretim standartlarına yakın bir **Distributed Key-Value Store** (Toy DynamoDB) mimarisine evrilen süreci modeller.

Amaç sadece kod yazmak değil; "Race Condition", "Deadlock", "TOCTOU" ve "Data Hotspots" gibi mühendislik problemlerini yerinde tespit edip çözmektir.

## Proje Yapısı

Proje, mimari evrimin her aşamasını temsil eden modüllerden oluşur:

```text
.
├── consistent_hashing                  # V2: Temel Halka (Ring) Yapısı
├── consistent_virtual_hashing          # V3: Sanal Düğümler (Virtual Nodes)
├── consistent_virtual_hashing_replicated # V4: Replikasyon Mantığı (Simulation)
├── modular_hashing                     # V1: Klasik Modülo Yöntemi
└── toy_dynamodb                        # V5: Storage Engine, Coordinator & KV Store
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

### 5. toy_dynamodb (V5 - Storage Engine & Coordinator)

Bu modül, projenin **"InMemory Distributed Database"** haline gelmiş sürümüdür. Artık sadece adres hesaplanmaz, veri gerçekten saklanır ve okunur.

#### Mimari Özellikler:

* **Package Separation & Encapsulation:**
* **`ring` Paketi (Coordinator):** Sistemin beynidir. İstemciden gelen isteği karşılar, verinin hangi node'lara (ve replikalara) gideceğini hesaplar ve trafiği yönetir.
* **`node` Paketi (Storage):** Sistemin kasasıdır. Sadece kendisine verilen veriyi saklar. Verinin nereden geldiğini veya başka kopyaları olup olmadığını bilmez.


* **Concurrency & Thread-Safety:**
* **Granular Locking:** `Ring` yapısı topoloji değişiklikleri (Node ekleme) için kendi `RWMutex`ini, her bir `Node` ise veri yazma/okuma işlemleri için kendi özel `RWMutex`ini kullanır. Bu sayede sistem kilitlenmeden (Deadlock-free) çalışır.
* **TOCTOU (Time-of-Check to Time-of-Use) Koruması:** Node ekleme sırasında oluşabilecek yarış durumları (Race Conditions), "Double-Check Locking" deseni ile engellenmiştir.


* **Dynamic Replication:**
* `RCount` (Replication Count) parametresi dinamik hale getirilmiştir. Sistem başlatılırken verinin kaç kopyasının tutulacağı (N) belirlenebilir.
* `Put` ve `Get` işlemleri, belirlenen replikasyon faktörü (N) kadar sunucuyu gezerek veri tutarlılığını sağlar.



## Nasıl Çalıştırılır?

En güncel Distributed KV Store versiyonunu çalıştırmak için:

```bash
cd toy_dynamodb
go run main.go
```

### Örnek Kullanım (V5)

Aşağıdaki örnekte 6 sunuculu bir küme oluşturulmuş ve replikasyon sayısı (`RCount`) 5 olarak ayarlanmıştır.

```go
func main() {
    // Ring başlatılır ve Replikasyon Sayısı 5 olarak ayarlanır
    ring := Ring.Ring{RCount: 5}
    ring.Init()

    // Node'lar güvenli bir şekilde eklenir
    nodes := []string{"foo", "bar", "bazz", "zapp", "zucc", "bars"}
    for _, n := range nodes {
        if err := ring.AddNode(n); err != nil {
            log.Fatalf("%v", err)
        }
    }

    // Veri yazma (Coordinator 5 replikaya da yazar)
    ring.Put("Mahmut", "Ozer")

    // Veri okuma (Coordinator replikalardan veriyi çeker)
    vals, found := ring.Get("Mahmut")
    
    if found {
        fmt.Println(vals) 
    }
}

```

**Beklenen Çıktı:**

```text
[Ozer Ozer Ozer Ozer Ozer]
```

*(Verinin 5 farklı fiziksel sunucuda başarıyla saklandığını ve okunduğunu gösterir.)*


