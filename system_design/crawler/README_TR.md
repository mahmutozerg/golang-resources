# Go & Playwright Web Crawler

Bu proje, Go ve Playwright kullanılarak geliştirilmiş, ölçeklenebilir ve kurallara uyan (polite) bir web crawler uygulamasıdır. Basit bir HTML ayrıştırıcısı olmak yerine, sayfaları gerçek bir tarayıcı motoruyla işler ve içerikleri **MHTML** formatında (tüm CSS, JS ve görsellerle birlikte tek dosya) kaydeder.

## Özellikler

- **MHTML Snapshot:** Sayfaları sadece metin olarak değil, o anki görsel durumuyla (DOM state) tam bir paket olarak kaydeder.
- **Host-Bazlı Hız Limiti (Smart Concurrency):** Global bir hız limiti yerine, her domain için ayrı bir hız limiti uygular. Örneğin, hızlı yanıt veren bir siteyi tararken, yavaş yanıt veren başka bir site (örn. yüksek `Crawl-delay` isteyenler) yüzünden tarama işlemi duraksamaz.
- **Şeffaf Bekleme Logları:** Eğer bir site `robots.txt` üzerinden uzun bir bekleme süresi (örn. 30 saniye) zorunlu kılıyorsa, worker sessizce donmak yerine ne kadar bekleyeceğini loglar.
- **Robots.txt Uyumluluğu:** Her domainin `robots.txt` dosyasını kontrol eder, `Disallow` kurallarına uyar ve `Crawl-delay` yönergelerine göre hızını otomatik ayarlar.
- **Playwright Altyapısı:** Modern web sitelerini (SPA, dinamik içerik) sorunsuz tarayabilmek için Chromium motorunu kullanır.
- **Graceful Shutdown:** `CTRL+C` veya terminasyon sinyali geldiğinde, elindeki işleri bitirmeden ve kaynakları temizlemeden kapanmaz.

## Kurulum

Projeyi çalıştırmadan önce Go'nun(1.25.6) yüklü olduğundan emin ol. Ayrıca tarayıcı otomasyonu için Playwright bağımlılıklarını kurman gerekiyor.

1. Projeyi klonla ve dizine gir.
2. Go modüllerini indir:

```bash
go mod tidy
```

3. Playwright sürücülerini ve bağımlılıklarını yükle:

```bash
go run [github.com/playwright-community/playwright-go/cmd/playwright@latest](https://github.com/playwright-community/playwright-go/cmd/playwright@latest) install --with-deps

```

## Kullanım

Crawler'ın hangi adreslerden başlayacağını belirlemek için projenin çalıştığı dizinde `seed.txt` adında bir dosya olması gerekir.

1. `seed.txt` dosyasını oluştur ve taramak istediğin başlangıç URL'lerini satır satır ekle:

```text
https://go.dev/
https://news.ycombinator.com/
https://www.wikipedia.org/
```

2. Uygulamayı çalıştır:

```bash
go run main.go
```

## Yapılandırma

Crawler'ın ayarları `internal/config` paketi altında sabit (const) olarak tanımlanmıştır. İhtiyaçlarına göre bu değerleri kod üzerinden değiştirebilirsin:

- **JobQueueSize:** İş kuyruğunun kapasitesi (Varsayılan: 1000).
- **MaxDepth:** Linklerin ne kadar derinlemesine takip edileceği (Varsayılan: 3).
- **ConcurrentWorkerCount:** Aynı anda kaç tane worker çalışacağı (Varsayılan: 5).
- **GoToRegularTimeOutMs:** Sayfa yüklenmesi için maksimum bekleme süresi.
- **JitterMin/Max:** İstekler arasına eklenen rastgele gecikme süreleri (İnsansı davranış için).

## Nasıl Çalışır?

Sistem temel olarak bir **Producer-Consumer** modelidir, ancak "Registry" dediğimiz akıllı bir ara katman kullanır.

1. **Seed Loading:** `seed.txt` dosyasındaki URL'ler okunur ve iş kuyruğuna atılır.
2. **Worker Havuzu:** Belirlenen sayıda (ConcurrentWorkerCount) worker ayağa kalkar ve kuyruktan iş almaya başlar.
3. **Policy Registry:** Bir worker URL'i aldığında, önce Registry'ye gider. Registry, o domain için daha önce `robots.txt` indirildi mi diye bakar.

- Eğer ilk kez gidiliyorsa: `robots.txt` indirilir, kurallar parse edilir ve o domain için özel bir `RateLimiter` oluşturulur.
- Bu işlem thread-safe (Double-Checked Locking) bir şekilde yapılır.

4. **Fetch & Save:** İzin varsa ve hız limiti aşılmadıysa, Playwright sayfaya gider, MHTML snapshot'ını alır ve diske kaydeder.
5. **Discovery:** Sayfadaki linkler (`a href`) bulunur, filtrelenir (OnlySameOrigin) ve tekrar kuyruğa eklenir.

## Çıktılar

Kaydedilen sayfalar, projenin iki üst dizinindeki `files` klasöründe (`../../files`), domain ve path yapısına uygun klasörler halinde saklanır.

Örnek çıktı yolu:
`../../files/go.dev/doc/tutorial/index.html/20260209T120000.mhtml`
