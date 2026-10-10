# Technical Debt: CI/CD Cache Poisoning Risk (TD-0001)

- **ID:** TD-0001
- **Tarih:** 2026-10-10
- **Durum:** Accepted (Debt Recorded)
- **İlgili Kural:** zizmor `cache-poisoning` (`https://docs.zizmor.sh/audits/#cache-poisoning`)
- **İlgili Dosya:** `.github/workflows/ci.yml`

---

## 1. Problem ve Bağlam

GitHub Actions üzerindeki `actions/setup-go` adımı, Go derleme ve modül bağımlılıklarını (`~/go/pkg/mod`, `~/.cache/go-build`) hızlandırmak amacıyla varsayılan olarak önbellek (`cache: true`) kullanmaktadır.

Pipeline'ımız (`ci.yml`) monolitik bir yapıda tasarlanmıştır:
1. `push` (main/tags) ve `pull_request` olaylarında çalışır.
2. İlk aşamalarda kod doğrulama, linting, test ve security taramaları yapar (Go cache burada kullanılır).
3. Sonraki aşamalarda (`build-api`, `tag-latest`), `pull_request` olmayan event'lerde Docker imajı derleyip GHCR'a pushlar ve Cosign ile imzalar.

Zizmor statik güvenlik tarayıcısı, aynı workflow içinde hem runtime artifact publishing / container release adımlarının bulunmasını hem de `actions/setup-go` ile genel cache kullanımını bir **Cache Poisoning** vektörü olarak işaretlemektedir:
> *"Düşük yetkili veya harici bir PR akışında bağımlılık/cache zehirlenmesi gerçekleşirse, aynı repo cache'ini okuyan release/imzalama süreçleri bu zehirlenmiş durumdan etkilenebilir."*

## 2. Neden Şu Anda Doğrudan Çözülmedi?

1. **Geliştirici Deneyimi ve CI Süreleri:** 
   `actions/setup-go` cache'inin tamamen kapatılması (`cache: false`), her PR ve push koşumunda tüm Go modüllerinin ve test derlemelerinin sıfırdan indirilip derlenmesine yol açarak CI sürelerini 3-4 katına çıkaracaktır.
2. **Mimari Ayrım İhtiyacı:**
   En temiz ve kalıcı mimari çözüm; PR doğrulama workflow'u (`ci.yml`) ile imaj derleme/yayınlama workflow'unun (`release.yml` veya `publish.yml`) birbirinden bağımsız iki ayrı workflow dosyasına ayrılmasıdır. Bu ayrım ayrı bir refactoring kapsamındadır.

## 3. Geçici Çözüm (Mitigation)

- `.github/zizmor.yml` içinde `cache-poisoning` kuralı bu teknik borç dosyası (`docs/techdebts/0001-ci-cache-poisoning.md`) referans verilerek ignore listesine alınmıştır.
- `ci.yml` içindeki imaj derleme adımı (`build-api`), Go binary'sini runner üzerinde derleyip imaja kopyalamak yerine, izole bir Docker container (`Dockerfile` multi-stage build) içinde sıfırdan derlemektedir. Dolayısıyla runner üzerindeki Go modül cache'i doğrudan Docker imajının içine sızmamaktadır.

## 4. Kalıcı Çözüm Planı (Remediation Plan)

1. İmaj build & push aşamaları `ci.yml` dosyasından çıkarılıp yalnızca `main` branch push ve `v*.*.*` tag push olaylarında tetiklenen ayrı bir `.github/workflows/release.yml` dosyasına taşınacak.
2. `release.yml` workflow'unda `cache-mode: none` uygulanarak sıfır cache ile temiz ve izole bir derleme ortamı garanti edilecek.
3. `.github/zizmor.yml` dosyasındaki kural istisnası kaldırılacak.
