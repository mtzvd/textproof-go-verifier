
# TextProof

[English](README_en.md) | [Русский](README_ru.md)

## System potwierdzania autorstwa tekstów z wykorzystaniem technologii blockchain

TextProof to aplikacja webowa do rejestrowania autorstwa dokumentów tekstowych w blockchainie. System wykorzystuje skróty kryptograficzne i Proof-of-Work do stworzenia niezmiennego zapisu istnienia tekstu w określonym momencie.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

---

## Możliwości

- **Deponowanie tekstów** — Zarejestruj autorstwo swojego tekstu w blockchainie
- **Weryfikacja autentyczności** — Sprawdź tekst po ID lub pełnej treści
- **Blockchain z Proof-of-Work** — Ochrona przed fałszowaniem przez wydobywanie bloków
- **Niezawodne przechowywanie** — WAL (Write-Ahead Logging) + automatyczne kopie zapasowe
- **Kody QR** — Do szybkiej weryfikacji na urządzeniach mobilnych
- **Osadzalne odznaki** — Widżety HTML dla stron internetowych
- **Szybkie wyszukiwanie** — O(1) wyszukiwanie duplikatów przez indeksowanie
- **Nowoczesny interfejs** — Bulma CSS + Alpine.js

---

## Szybki start

### Wymagania

- [Go](https://golang.org/dl/) 1.21 lub nowszy
- [Templ](https://templ.guide/) do generowania szablonów

### Instalacja

#### Sklonuj repozytorium

```bash
git clone https://github.com/yourusername/textproof.git
cd textproof
```

#### Zainstaluj zależności

```bash
go mod download
```

#### Zainstaluj templ (jeśli jeszcze nie zainstalowany)

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

#### Wygeneruj szablony

```bash
templ generate
```

#### Uruchom serwer

```bash
go run cmd/server/main.go
```

Aplikacja będzie dostępna pod adresem: **<http://localhost:8080>**

---

## Użycie

### Deponowanie tekstu

1. Przejdź do `/deposit`
2. Wypełnij formularz:
   - Imię autora (Imię i nazwisko lub pseudonim)
   - Tytuł utworu
   - Pełny tekst dokumentu
   - (Opcjonalnie) Klucz publiczny dla podpisu elektronicznego
3. Kliknij "Zarejestruj w blockchainie"
4. Otrzymaj unikalny ID i kod QR

### Weryfikacja tekstu

**Po identyfikatorze:**

1. Przejdź do `/verify`
2. Wybierz zakładkę "Po identyfikatorze"
3. Wprowadź ID bloku (np.: `000-000-001`)
4. Otrzymaj informacje o tekście

**Po treści:**

1. Przejdź do `/verify`
2. Wybierz zakładkę "Po tekście"
3. Wklej pełny tekst dokumentu
4. System obliczy skrót i sprawdzi obecność w blockchainie

**Bezpośredni link:**

- Otwórz `/verify/{id}` dla automatycznej weryfikacji

---

## Architektura

### Struktura projektu

```bash
textproof/
├── cmd/
│   └── server/
│       └── main.go              # Punkt wejścia aplikacji
├── internal/
│   ├── api/                     # Obsługa HTTP i routing
│   │   ├── api.go
│   │   ├── flash.go             # Wiadomości flash (cookies)
│   │   └── map_stats.go
│   ├── blockchain/              # Logika blockchain
│   │   ├── block.go             # Struktura bloku
│   │   ├── blockchain.go        # Główna logika łańcucha
│   │   ├── storage.go           # Praca z plikami
│   │   ├── errors.go            # Typy błędów
│   │   └── id_generator.go      # Generator ID bloków
│   ├── config/                  # Konfiguracja
│   │   └── config.go
│   └── viewmodels/              # Modele danych dla UI
│       ├── types.go
│       ├── navbar.go
│       └── build-navbar.go
├── web/
│   ├── static/                  # Pliki statyczne
│   │   ├── css/
│   │   │   └── styles.css
│   │   └── js/
│   │       └── app.js
│   └── templates/               # Szablony Templ
│       ├── base.templ
│       ├── home.templ
│       ├── deposit.templ
│       ├── deposit_result_page.templ
│       ├── verify.templ
│       ├── verify_result.templ
│       └── components/          # Komponenty wielokrotnego użytku
├── data/                        # Dane blockchain (nie w git)
│   ├── blockchain.json          # Główny łańcuch
│   ├── wal.json                 # Write-Ahead Log
│   └── backups/                 # Automatyczne kopie zapasowe
├── go.mod
├── go.sum
├── modd.conf                    # Konfiguracja hot reload
├── .gitignore
└── README.md
```

### Blockchain

**Struktura bloku:**

```bash
type Block struct {
    ID        string       // "000-000-001"
    PrevHash  string       // Skrót poprzedniego bloku
    Timestamp time.Time    // Czas utworzenia
    Data      DepositData  // Dane o tekście
    Nonce     int          // Proof-of-Work nonce
    Hash      string       // Skrót SHA-256 bloku
}

type DepositData struct {
    AuthorName  string  // Imię autora
    Title       string  // Tytuł
    TextStart   string  // Pierwsze 3 słowa
    TextEnd     string  // Ostatnie 3 słowa
    ContentHash string  // Skrót SHA-256 pełnego tekstu
    PublicKey   string  // (Opcjonalnie) Klucz publiczny
}
```

**Proof-of-Work:**

- Konfigurowalna trudność (domyślnie: 4 zera)
- Wydobywanie bloku zajmuje kilka sekund
- Ochrona przed fałszowaniem przeszłych wpisów

**Przechowywanie:**

- Pliki JSON dla prostoty
- WAL dla ochrony przed awariami
- Automatyczne kopie zapasowe (przechowywane ostatnie 5)
- Atomic write przez pliki tymczasowe

---

## Konfiguracja

### Flagi wiersza poleceń

```bash
go run cmd/server/main.go [opcje]
```

Opcje:
  -data-dir string
        Katalog do przechowywania danych (domyślnie "data")
  -port int
        Port dla serwera HTTP (domyślnie 8080)
  -difficulty int
        Trudność wydobywania (liczba zer) (domyślnie 4)
  -debug
        Włącz tryb debugowania

### Przykłady

#### Uruchom na porcie 9090 z danymi w ./my_data

```bash
go run cmd/server/main.go -data-dir ./my_data -port 9090
```

#### Uruchom z obniżoną trudnością dla testów

```bash
go run cmd/server/main.go -difficulty 3 -debug
```

---

## Rozwój

### Hot Reload z modd

#### Zainstaluj modd

```bash
go install github.com/cortesi/modd/cmd/modd@latest
```

#### Uruchom z automatycznym przeładowaniem

```bash
modd
```

Przy zmianie plików `.templ` automatycznie uruchomi się `templ generate` i serwer się zrestartuje.

### Struktura API

| Metoda | Ścieżka | Opis |
| --- | --- | --- |
| GET | `/` | Strona główna |
| GET | `/deposit` | Formularz deponowania |
| POST | `/api/deposit` | Przetwarzanie deponowania |
| GET | `/deposit/result/{id}` | Wynik deponowania |
| GET | `/verify` | Formularz weryfikacji |
| POST | `/api/verify/id` | Weryfikacja po ID |
| POST | `/api/verify/text` | Weryfikacja po tekście |
| GET | `/verify/result/{id}` | Wynik weryfikacji |
| GET | `/verify/{id}` | Bezpośredni link do weryfikacji |
| GET | `/api/qrcode/{id}` | Generowanie kodu QR |
| GET | `/api/badge/{id}` | Odznaka HTML do osadzania |
| GET | `/api/stats` | Statystyki blockchain |

---

## Bezpieczeństwo

### Zaimplementowane środki

- ✅ **Walidacja danych wejściowych** — maksymalna długość tekstu, sprawdzanie pól
- ✅ **HttpOnly cookies** — ochrona wiadomości flash przed XSS
- ✅ **Proof-of-Work** — ochrona przed spamem
- ✅ **Indeks skrótów treści** — zapobieganie duplikatom
- ✅ **Atomic writes** — ochrona przed uszkodzeniem danych

### Zalecenia dla produkcji

- ⚠️ Dodaj **ograniczenie żądań** (np. przez middleware)
- ⚠️ Używaj **HTTPS** (certyfikaty TLS)
- ⚠️ Skonfiguruj **ochronę CSRF**
- ⚠️ Dodaj **logowanie** (zerolog, zap)
- ⚠️ Zaimplementuj **monitorowanie** (Prometheus + Grafana)
- ⚠️ Skonfiguruj **kopie zapasowe** danych

---

## Wydajność

| Operacja | Złożoność | Czas |
| --- | --- | --- |
| Wyszukiwanie po ID | O(n) | ~1ms dla 1000 bloków |
| Wyszukiwanie po skrócie | O(1) | <1ms (indeksowanie) |
| Wydobywanie bloku | - | ~2-5s (difficulty=4) |
| Walidacja łańcucha | O(n) | ~10ms dla 1000 bloków |

---

##### Testowanie

#### Uruchom wszystkie testy

```bash
go test ./...
```
#### Testy z pokryciem

```bash
go test -cover ./...
```

#### Generuj raport pokrycia

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## TODO / Plan rozwoju

- [ ] Dodaj testy jednostkowe dla blockchain
- [ ] Dodaj testy integracyjne dla API
- [ ] Zaimplementuj ograniczenie żądań
- [ ] Dodaj strukturyzowane logowanie
- [ ] Wsparcie dla PostgreSQL/MySQL zamiast JSON
- [ ] Klucze API dla automatyzacji
- [ ] Eksport blockchain w różnych formatach
- [ ] Wsparcie dla podpisów cyfrowych (ECDSA, RSA)
- [ ] Konteneryzacja Docker
- [ ] Pipeline CI/CD (GitHub Actions)
- [ ] Metryki Prometheus
- [ ] Dokumentacja Swagger/OpenAPI

---

## Wkład w rozwój

Contributions are welcome! Proszę:

1. Sforkuj projekt
2. Utwórz branch z funkcją (`git checkout -b feature/AmazingFeature`)
3. Zatwierdź zmiany (`git commit -m 'Add some AmazingFeature'`)
4. Wypchnij do brancha (`git push origin feature/AmazingFeature`)
5. Otwórz Pull Request

---

## Licencja

Ten projekt jest dystrybuowany na licencji MIT. Szczegóły w pliku `LICENSE`.

---

## 👤 Autor

### **Georgiy Agafonov**

- GitHub: [@mtzvd](https://github.com/mtzvd)
- Email: <info@web-n-roll.pl>

---

## Podziękowania

- [Bulma](https://bulma.io/) — framework CSS
- [Alpine.js](https://alpinejs.dev/) — lekki framework JS
- [Templ](https://templ.guide/) — type-safe szablony dla Go
- [Gorilla Mux](https://github.com/gorilla/mux) — router HTTP
- [go-qrcode](https://github.com/skip2/go-qrcode) — generowanie kodów QR

---

## Dodatkowe zasoby

- [Dokumentacja Go](https://golang.org/doc/)
- [Przewodnik Templ](https://templ.guide/)
- [Podstawy Blockchain](https://en.wikipedia.org/wiki/Blockchain)

---

**⭐ Jeśli podoba Ci się projekt — dodaj gwiazdkę na GitHubie!**
