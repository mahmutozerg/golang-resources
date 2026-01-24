# Go ile Dağıtık Sistemler: Consistent Hashing ve Key-Value Store Implementasyonu

Bu proje, dağıtık sistemlerde veri ve yük dağıtımı için kullanılan hashing algoritmalarının ve veri saklama stratejilerinin Go dili ile yazılmış adım adım implementasyonunu içerir.

Proje, temel bir modülo yaklaşımından başlayıp, **Replication (Yedekleme)** ve **Concurrency (Eşzamanlılık)** özelliklerine sahip, production standartlarına yakın bir **Distributed Key-Value Store** çekirdeğine doğru evrilen bir süreci modeller.

Amaç sadece çalışan kod yazmak değil; "Cache Miss Storm", "Data Hotspots" ve "Single Point of Failure" gibi mühendislik problemlerini analiz edip her aşamada çözmektir.

## Proje Yapısı

Proje, algoritmaların gelişim sırasına göre dört ana modülden oluşur:

```text
system_design
├── modular_hashing             # V1: Klasik Modülo Yöntemi
├── consistent_hashing          # V2: Temel Halka (Ring) Yapısı
├── consistent_virtual_hashing  # V3: Sanal Düğümler ve Yük Dengeleme
└── consistent_virtual_hashing_replicated    # V4: Replikasyon, Thread-Safety ve KV Store

```

## Modüller ve Teknik Detaylar

### 1. modular_hashing (Klasik Yaklaşım)

Yük dağıtımı için kullanılan en ilkel yöntemdir. `hash(key) % sunucu_sayısı` formülü ile çalışır.

* **Problem:** Sistemdeki sunucu sayısı değiştiğinde (scaling), var olan anahtarların neredeyse tamamının yeri değişir. Bu durum "Cache Miss Storm" felaketine yol açar.
* **Sonuç:** Ölçeklenebilir sistemler için uygun değildir.

### 2. consistent_hashing (Halka Topolojisi)

Consistent Hashing algoritmasının temel implementasyonudur. Sunucular ve veriler bir sayı doğrusuna (halka) yerleştirilir. Veri, saat yönünde kendisine en yakın sunucuya atanır.

* **İyileştirme:** Sunucu eklendiğinde/çıkarıldığında sadece komşu düğümler etkilenir. Veri taşıma maliyeti minimize edilir.
* **Yeni Problem:** Veri dağılımı homojen değildir (Data Hotspots).

### 3. consistent_virtual_hashing (Sanal Düğümler & O(N) Optimizasyonu)

Dağılım dengesizliğini çözmek için **Virtual Nodes** tekniği uygulanmıştır. Her fiziksel sunucu halkada yüzlerce sanal kopya ile temsil edilir.

* **Teknik:** Go 1.21 `slices` paketi ile "Single-Pass Filtering" yapılarak node silme işlemi optimize edilmiştir.
* **Sonuç:** Mükemmele yakın yük dağılımı (%1 varyasyon).


### 4. consistent_virtual_hashing_replicated (Final: Replikasyon & Thread Safety)

Bu modül, dağıtık bir Key-Value Store'un **"Veri Yerleşimi ve Yönlendirme" (Data Placement & Routing)** katmanını simüle eder. Henüz verinin kendisini diske yazmaz, ancak verinin hangi sunucularda yedekleneceğinin matematiğini çözer.

#### Teknik Özellikler:

* **Replication Strategy (N=3):** Veri tek bir sunucuda değil, belirlenen `N` sayısı kadar sunucuda yedeklenir. Bir sunucu ölse bile veri kaybolmaz.
* **Distinct Physical Node Selection:** Algoritma, yedekleri seçerken sanal node'ları atlayıp **birbirinden farklı fiziksel sunucuları** bulur. (Örn: A sunucusuna veriyi yazdıysa, yedeği tekrar A'nın sanal node'una yazmaz).
* **Concurrency Control:** `sync.RWMutex` kullanılarak yapı **Thread-Safe** hale getirilmiştir. Yüksek trafik altında `Fatal Error: Concurrent map write` hataları engellenmiştir.
* **Ring Wrap-Around:** Halkanın sonuna gelindiğinde başa dönerek aramanın devam etmesini sağlayan matematiksel döngü (`modulo arithmetic`) kurulmuştur.

## Nasıl Çalıştırılır?

Test etmek istediğiniz modülün dizinine gidip `go run` komutunu kullanabilirsiniz.

```bash
cd consistent_virtual_hashing_replicated
go run main.go

```

### Örnek Test Sonucu (V4 - Replication)

1 milyon verinin 3 kopya (Replica=3) ile 4 sunucuya dağıtıldığı ve `node_1`'in sistemden ani düşüşünün simüle edildiği senaryo:

```text
--- Başlangıç Dağılımı (1.000.000 Veri x 3 Kopya) ---
node_2: 760,436
node_1: 747,766
node_3: 751,179
node_4: 740,619
# Not: Dağılım son derece dengeli.

--- NODE_1 SİLİNDİKTEN SONRA (Failover) ---
node_2: 1,000,000 (%100)
node_3: 1,000,000 (%100)
node_4: 1,000,000 (%100)

```
