# TextProof

## Text Authorship Proof System Using Blockchain Technology

TextProof is a web application for recording authorship of text documents in a blockchain. The system uses cryptographic hashes and Proof-of-Work to create an immutable record of a text's existence at a specific point in time.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

---

## Features

- **Text Deposit** — Record the authorship of your text in the blockchain
- **Authenticity Verification** — Verify a text by ID or by its full content
- **Proof-of-Work Blockchain** — Protection against forgery through block mining
- **Reliable Storage** — WAL (Write-Ahead Logging) + automatic backups
- **QR Codes** — For quick verification on mobile devices
- **Embeddable Badges** — HTML widgets for websites
- **Fast Search** — O(1) duplicate search via indexing
- **Modern UI** — Bulma CSS + Alpine.js

---

## Quick Start

### Requirements

- [Go](https://golang.org/dl/) 1.21 or newer
- [Templ](https://templ.guide/) for template generation

### Installation

#### Clone the repository

```bash
git https://github.com/mtzvd/textproof-go-verifier.git
cd textproof
```

#### Install dependencies

```bash
go mod download
```

#### Install templ (if not already installed)

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

#### Generate templates

```bash
templ generate
```

#### Start the server

```bash
go run cmd/server/main.go
```

The application will be available at: **<http://localhost:8080>**

---

## Usage

### Depositing Text

1. Go to `/deposit`
2. Fill out the form:
   - Author's name (full name or pseudonym)
   - Work title
   - Full text of the document
   - (Optional) Public key for digital signature
3. Click "Record in Blockchain"
4. Receive a unique ID and QR code

### Verifying Text

**By identifier:**

1. Go to `/verify`
2. Select the "By Identifier" tab
3. Enter the block ID (e.g., `000-000-001`)
4. Get information about the text

**By content:**

1. Go to `/verify`
2. Select the "By Text" tab
3. Paste the full text of the document
4. The system will compute the hash and check its presence in the blockchain

**Direct link:**

- Open `/verify/{id}` for automatic verification

---

## Architecture

### Project Structure

```text
textproof/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/                     # HTTP handlers and routes
│   │   ├── api.go
│   │   ├── flash.go             # Flash messages (cookies)
│   │   └── map_stats.go
│   ├── blockchain/              # Blockchain logic
│   │   ├── block.go             # Block structure
│   │   ├── blockchain.go        # Main chain logic
│   │   ├── storage.go           # File operations
│   │   ├── errors.go            # Error types
│   │   └── id_generator.go      # Block ID generator
│   ├── config/                  # Configuration
│   │   └── config.go
│   └── viewmodels/              # Data models for UI
│       ├── types.go
│       ├── navbar.go
│       └── build-navbar.go
├── web/
│   ├── static/                  # Static files
│   │   ├── css/
│   │   │   └── styles.css
│   │   └── js/
│   │       └── app.js
│   └── templates/               # Templ templates
│       ├── base.templ
│       ├── home.templ
│       ├── deposit.templ
│       ├── deposit_result_page.templ
│       ├── verify.templ
│       ├── verify_result.templ
│       └── components/          # Reusable components
├── data/                        # Blockchain data (not in git)
│   ├── blockchain.json          # Main chain
│   ├── wal.json                 # Write-Ahead Log
│   └── backups/                 # Automatic backups
├── go.mod
├── go.sum
├── modd.conf                    # Hot reload configuration
├── .gitignore
└── README.md
```

### Blockchain

**Block structure:**

```go
type Block struct {
    ID        string       // "000-000-001"
    PrevHash  string       // Hash of the previous block
    Timestamp time.Time    // Creation time
    Data      DepositData  // Text data
    Nonce     int          // Proof-of-Work nonce
    Hash      string       // SHA-256 block hash
}

type DepositData struct {
    AuthorName  string  // Author's name
    Title       string  // Title
    TextStart   string  // First 3 words
    TextEnd     string  // Last 3 words
    ContentHash string  // SHA-256 hash of the full text
    PublicKey   string  // (Optional) Public key
}
```

**Proof-of-Work:**

- Configurable difficulty (default: 4 leading zeros)
- Block mining takes a few seconds
- Protection against forgery of past records

**Storage:**

- JSON files for simplicity
- WAL for crash protection
- Automatic backups (last 5 are kept)
- Atomic writes via temporary files

---

## Configuration

### Command Line Flags

```bash
go run cmd/server/main.go [options]
```

Options:
  -data-dir string
        Data storage directory (default "data")
  -port int
        Port for HTTP server (default 8080)
  -difficulty int
        Mining difficulty (number of zeros) (default 4)
  -debug
        Enable debug mode

### Examples

#### Run on port 9090 with data in ./my_data

```bash
go run cmd/server/main.go -data-dir ./my_data -port 9090
```

#### Run with reduced difficulty for testing

```bash
go run cmd/server/main.go -difficulty 3 -debug
```
---

## Development

### Hot Reload with modd

#### Install modd

```bash
go install github.com/cortesi/modd/cmd/modd@latest
```

#### Run with auto-reload

```bash
modd
```

When `.templ` files change, `templ generate` will run automatically and the server will restart.

### API Structure

| Method | Path | Description |
| --- | --- | --- |
| GET | `/` | Home page |
| GET | `/deposit` | Deposit form |
| POST | `/api/deposit` | Deposit processing |
| GET | `/deposit/result/{id}` | Deposit result |
| GET | `/verify` | Verification form |
| POST | `/api/verify/id` | Verify by ID |
| POST | `/api/verify/text` | Verify by text |
| GET | `/verify/result/{id}` | Verification result |
| GET | `/verify/{id}` | Direct verification link |
| GET | `/api/qrcode/{id}` | QR code generation |
| GET | `/api/badge/{id}` | HTML badge for embedding |
| GET | `/api/stats` | Blockchain statistics |

---

## Security

### Implemented Measures

- ✅ **Input validation** — maximum text length, field validation
- ✅ **HttpOnly cookies** — protection of flash messages from XSS
- ✅ **Proof-of-Work** — protection against spam
- ✅ **Content hash index** — prevention of duplicates
- ✅ **Atomic writes** — protection against data corruption

### Recommendations for Production

- ⚠️ Add **rate limiting** (e.g., via middleware)
- ⚠️ Use **HTTPS** (TLS certificates)
- ⚠️ Configure **CSRF protection**
- ⚠️ Add **logging** (zerolog, zap)
- ⚠️ Implement **monitoring** (Prometheus + Grafana)
- ⚠️ Set up data **backup**

---

## Performance

| Operation | Complexity | Time |
| --- | --- | --- |
| Search by ID | O(n) | ~1ms for 1000 blocks |
| Search by hash | O(1) | <1ms (indexing) |
| Block mining | - | ~2-5s (difficulty=4) |
| Chain validation | O(n) | ~10ms for 1000 blocks |

---

## Testing

#### Run all tests

```bash
go test ./...
```

#### Tests with coverage

```bash
go test -cover ./...
```

#### Generate coverage report

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## TODO / Roadmap

- [ ] Add unit tests for blockchain
- [ ] Add integration tests for API
- [ ] Implement rate limiting
- [ ] Add structured logging
- [ ] Support PostgreSQL/MySQL instead of JSON
- [ ] API keys for automation
- [ ] Export blockchain in various formats
- [ ] Support for digital signatures (ECDSA, RSA)
- [ ] Docker containerization
- [ ] CI/CD pipeline (GitHub Actions)
- [ ] Prometheus metrics
- [ ] Swagger/OpenAPI documentation

---

## Contributing

Contributions are welcome! Please:

1. Fork the project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## License

This project is distributed under the MIT license. See the `LICENSE` file for details.

---

## 👤 Author

### **Georgiy Agafonov**

- GitHub: [@mtzvd](https://github.com/mtzvd)
- Email: <info@web-n-roll.pl>

---

## Acknowledgments

- [Bulma](https://bulma.io/) — CSS framework
- [Alpine.js](https://alpinejs.dev/) — lightweight JS framework
- [Templ](https://templ.guide/) — type-safe templates for Go
- [Gorilla Mux](https://github.com/gorilla/mux) — HTTP router
- [go-qrcode](https://github.com/skip2/go-qrcode) — QR code generation

---

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Templ Guide](https://templ.guide/)
- [Blockchain Basics](https://en.wikipedia.org/wiki/Blockchain)

---

**⭐ If you like the project — give it a star on GitHub!**