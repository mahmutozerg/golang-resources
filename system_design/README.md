Harika, emojilerden arındırılmış, tamamen teknik yetkinliğe ve sürecin mantığına odaklanan, temiz bir `README.md` hazırladım. Bunu projenin kök dizinine ekleyebilirsin.

---

# Dağıtık Sistemler ve Consistent Hashing Implementasyonu

Bu proje, dağıtık sistemlerde veri ve yük dağıtımı için kullanılan hashing algoritmalarının Go dili ile yazılmış adım adım implementasyonunu içerir. Proje, en temel yöntemden başlayarak production seviyesindeki optimize edilmiş Consistent Hashing yapısına kadar uzanan bir evrimi modeller.

Amaç sadece çalışan bir kod yazmak değil, her aşamada karşılaşılan mühendislik problemlerini (veri dağılımı dengesizliği, yeniden ölçekleme maliyeti, performans darboğazları) analiz edip bir sonraki aşamada bunları çözmektir.

## Proje Yapısı

Proje, algoritmaların gelişim sırasına göre üç ana modülden oluşur:

```text
system_design
├── modular_hashing          # V1: Klasik Modülo Yöntemi
├── consistent_caching       # V2: Temel Halka (Ring) Yapısı
└── modular_virtual_hashing  # V3: Sanal Düğümler ve O(N) Optimizasyonu

```

## Modüller ve Teknik Detaylar

### 1. modular_hashing (Klasik Yaklaşım)

Yük dağıtımı için kullanılan en ilkel yöntemdir. `hash(key) % sunucu_sayısı` formülü ile çalışır.

* **Problem:** Sistemdeki sunucu sayısı değiştiğinde (scaling), var olan anahtarların neredeyse tamamının yeri değişir. Bu durum, cache kullanan sistemlerde "Cache Miss Storm" felaketine yol açar.
* **Sonuç:** Ölçeklenebilir sistemler için uygun değildir.

### 2. consistent_caching (Halka Topolojisi)

Consistent Hashing algoritmasının temel implementasyonudur. Sunucular ve veriler `0` ile `2^64` arasındaki bir sayı doğrusuna (halka) yerleştirilir. Veri, saat yönünde kendisine en yakın olan sunucuya atanır.

* **İyileştirme:** Bir sunucu eklendiğinde veya çıkarıldığında sadece komşu düğümler etkilenir. Veri taşıma maliyeti minimize edilir.
* **Yeni Problem:** Veri dağılımı homojen değildir. Bazı sunuculara çok az, bazılarına aşırı yük binebilir (Data Hotspots).

### 3. modular_virtual_hashing (Final Çözüm)

Dağılım dengesizliğini çözmek için **Virtual Nodes (Sanal Düğümler)** tekniğinin uygulandığı ve performans optimizasyonu yapılmış son versiyondur.

#### Teknik Özellikler:

* **Virtual Nodes:** Her fiziksel sunucu için halka üzerine belirli sayıda (örneğin 100) sanal kopya yerleştirilir. Bu sayede sunucular halkaya homojen bir şekilde dağılır ve yük dengesi matematiksel olarak sağlanır.
* **Adil Yük Dağılımı:** Bir sunucu sistemden çıkarıldığında, üzerindeki yük tek bir sunucuya binmez; diğer sunuculara eşit oranda paylaştırılır.

#### Performans Optimizasyonu (Go 1.21+):

Önceki versiyonlarda sunucu silme işlemi (`RemoveNode`), her sanal kopya için tekrar tekrar hash hesaplaması ve array kaydırma işlemi gerektiriyordu. Bu durum büyük ölçekte CPU ve RAM darboğazı yaratıyordu.

Bu versiyonda **Single-Pass Filtering** tekniği kullanılmıştır:

* Dizi üzerinde sadece tek bir geçiş yapılır.
* Pahalı hash hesaplamaları yerine `Map Lookup` (O(1)) kullanılır.
* `slices.DeleteFunc` kullanılarak silinecek elemanlar "in-place" (yerinde) filtrelenir.
* Bu sayede karmaşıklık `O(M * N)` seviyesinden `O(N)` seviyesine indirilmiştir.

## Nasıl Çalıştırılır?

Test etmek istediğiniz modülün dizinine gidip `go run` komutunu kullanabilirsiniz.

```bash
cd modular_virtual_hashing
go run main.go

```

### Örnek Test Sonucu

1 milyon isteğin 4 sunucuya dağıtılması ve bir sunucunun (`node_1`) sistemden çıkarılması durumunda yükün nasıl yeniden dağıtıldığı aşağıdaki gibidir:

```text
--- Başlangıç Dağılımı ---
node_1: 242,551 (%24.2)
node_2: 310,862 (%31.0)
node_3: 220,996 (%22.0)
node_4: 225,591 (%22.5)

--- node_1 Silindikten Sonra (Yeniden Dağılım) ---
node_2: 376,598 (+66k)
node_3: 298,681 (+78k)
node_4: 324,721 (+99k)

```

Görüldüğü üzere, silinen sunucunun yükü kalan sunuculara adil bir şekilde paylaştırılmıştır.

## Geliştirme Notları

Bu proje, algoritmik mantığı ve veri yapılarını (Sorted Slice, Hash Map) kavramak amacıyla geliştirilmiştir. Canlı ortamda (Production) kullanılmadan önce, eşzamanlı istekleri yönetebilmek adına `sync.RWMutex` gibi mekanizmalarla Thread-Safety sağlanmalıdır.