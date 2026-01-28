# Toy DynamoDB: Distributed Key-Value Store (Go)

**Toy DynamoDB**, Amazon DynamoDB ve Cassandra mimarilerinden esinlenerek **Go** dilinde sıfırdan geliştirilmiş, dağıtık, kalıcı ve yüksek performanslı bir Key-Value veri tabanıdır.

Bu proje bir "kopyala-yapıştır" tutorial ürünü değildir; dağıtık sistemlerin temel problemleri olan **Race Condition**, **Network Partition**, **Consistency** ve **Durability** sorunlarını çözerek öğrenmek amacıyla geliştirilmiştir.

## Temel Özellikler

### Mimari & Dağıtım

- **Hexagonal Architecture:** Çekirdek mantık (Storage), ağ katmanından (gRPC) tamamen izole edilmiştir.
- **gRPC & Protobuf:** Node'lar ve İstemci arasında binary protokol ile tip güvenli iletişim.
- **Docker & Orchestration:** Her node kendi izole konteynerinde ve dosya sisteminde çalışır.

### Veri Tutarlılığı & Algoritmalar

- **Consistent Hashing:** `xxhash` tabanlı, sanal düğüm (Virtual Nodes) destekli yük dağıtımı.
- **Tunable Consistency:** İstemci, her işlem için `W` (Write Quorum) ve `R` (Read Quorum) seviyesini belirleyebilir ($R + W > N$).
- **Latency Hiding:** En yavaş sunucu beklenmez, çoğunluk (Quorum) sağlandığı an cevap dönülür (Early Exit).

### Depolama & Kalıcılık (Storage Engine)

- **LSM-Tree Benzeri Yapı:** Veriler RAM'de (MemTable) tutulur, diske (WAL) yazılır.
- **Write-Ahead Log (WAL):** Her yazma işlemi önce diske eklenir (`Append-Only`) ve `fsync` ile garanti altına alınır.
- **Crash Recovery:** Node yeniden başlatıldığında WAL dosyası okunur (Replay) ve hafıza restore edilir.
- **Data Integrity:** Log formatı `COMMAND,KEY,BASE64_VAL` şeklindedir, veri bozulmasına karşı korumalıdır.

## Kurulum ve Çalıştırma

### Gereksinimler

- Docker & Docker Compose
- Go 1.25+ (Lokal geliştirmeler için)

### Dağıtık Mod (Production Simulation)

3 node'lu bir cluster ayağa kaldırmak ve test etmek için:

```bash
# 1. Cluster'ı Başlat
docker-compose up --build -d

# 2. Logları İzle (Opsiyonel)
docker-compose logs -f

# 3. İstemciyi Çalıştır (Test)
go run cmd/docker_test/main.go
```

### In-Memory Mod (Hızlı Geliştirme)

Docker olmadan, tüm sistemi tek bir RAM bloğu içinde simüle etmek için (Network Bypass):

```bash
go run cmd/local_test/main.go

```

## Proje Yapısı

```text
.
├── cmd/
│   ├── server/           # Docker içinde çalışan gRPC Sunucusu (Entry Point)
│   ├── docker_test/      # Ağ üzerinden bağlanan CLI İstemcisi
│   └── local_test/       # Docker gerektirmeyen In-Memory Test Runner
├── pkg/
│   ├── adapter/          # LocalClient wrapper (Test için)
│   ├── node/             # Storage Engine (WAL + Map)
│   └── ring/             # Coordinator Logic (Hashing + Quorum)
├── proto/                # Protobuf tanımları (.proto) ve Go kodları
├── Errors/               # Özel hata tanımları
├── wal/                  # Docker Containerlerinin volumeleri
└── docker-compose.yaml

```

> **Not:** Mimari kararlar projenin kök dizinindeki `docs/decisions` klasöründe tutulmaktadır.

## Mimari Kararlar (ADR)

Projenin gelişim sürecinde alınan kritik teknik kararlar `../docs/decisions` altında belgelenmiştir:

- **0001:** Record Architecture Decisions
- **0002:** Consistent Hashing with Virtual Nodes
- **0003:** Replication and Quorum Semantics
- **0004:** Write-Through Cache and WAL First
- **0005:** Concurrency and Locking Strategy
- **0006:** Transition to Distributed Architecture (gRPC + Docker)
- **0007:** Define gRPC API Contract

## Kaynaklar & İlham

- _System Design Interview (Volume 1)_ - Alex Xu
