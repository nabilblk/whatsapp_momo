# WhatsApp Promo Code Chatbot POC

A proof-of-concept WhatsApp transactional chatbot for promo code redemption using Go.

## Architecture

```
┌─────────────────┐        HTTP          ┌─────────────────┐
│  WhatsApp       │  POST /api/v1/redeem │  Transactional  │
│  Bridge         │ ──────────────────── │  API Server     │
│  (whatsmeow)    │                      │  (Gin)          │
└────────┬────────┘                      └────────┬────────┘
         │ QR Auth                                │
         │ Session                               │
         │                                       │
    ┌────┴────────────────────────────────────────┴────┐
    │              SQLite Database                      │
    │  - promo_codes | transactions | rate_limits      │
    └──────────────────────────────────────────────────┘
```

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Start all services
docker-compose up --build

# Watch the bridge logs for QR code
docker-compose logs -f bridge

# Scan QR code with WhatsApp, then send test codes
```

### Local Development

```bash
# Install dependencies
cd api && go mod tidy && cd ..
cd bridge && go mod tidy && cd ..

# Run API server
cd api && go run ./cmd/api

# In another terminal, run bridge
cd bridge && CGO_ENABLED=1 go run ./cmd/bridge
```

## Test Codes

| Code | Result | Response (FR) |
|------|--------|---------------|
| `VALID100` | Success | 1GB de data mobile |
| `VALID200` | Success | 500 FCFA de credit |
| `EXPIRED01` | Error | Code expire |
| `USED001` | Error | Deja utilise |
| `INVALID` | Error | Code invalide |
| Other | Error | Code non reconnu |

## Testing the API

```bash
# Run test script
./scripts/test_api.sh

# Or test manually
curl -X POST http://localhost:8080/api/v1/redeem \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+237612345678", "code": "VALID100", "language": "fr"}'
```

## Configuration

Copy `.env.example` to `.env` and adjust as needed:

```env
# API Configuration
API_HOST=0.0.0.0
API_PORT=8080
API_DB_PATH=/data/promo.db

# Bridge Configuration
BRIDGE_API_URL=http://api:8080
BRIDGE_SESSION_PATH=/data/whatsapp.db

# Rate Limiting
RATE_LIMIT_PER_MINUTE=5
RATE_LIMIT_PER_HOUR=20

# Language
DEFAULT_LANGUAGE=fr
```

## Project Structure

```
whatsapp-promo-poc/
├── api/                    # API Server (Go + Gin)
│   ├── cmd/api/           # Entry point
│   └── internal/          # Business logic
├── bridge/                 # WhatsApp Bridge (Go + whatsmeow)
│   ├── cmd/bridge/        # Entry point
│   └── internal/          # WhatsApp integration
├── pkg/                    # Shared packages
│   ├── models/            # Data models
│   ├── i18n/              # Bilingual messages
│   └── errors/            # Error types
├── docs/                   # Documentation
└── scripts/               # Test scripts
```

## Rate Limiting

- 5 requests per minute per phone number
- 20 requests per hour per phone number

## Bilingual Support

The system supports French (fr) and English (en). Language is detected from:
1. Message keywords (bonjour, hello, etc.)
2. Default configuration

## API Endpoints

- `POST /api/v1/redeem` - Redeem a promo code
- `GET /api/v1/health` - Health check

See [docs/API.md](docs/API.md) for full documentation.

## Important: WhatsApp Relay Numbers (whatsmeow limitation)

When using this POC, you may notice that **responses come from a different phone number** (e.g., a Kenya +254 number) instead of the number you connected with the QR code.

### Why does this happen?

This is a known behavior when using **whatsmeow** (unofficial WhatsApp Web library):

1. whatsmeow uses the reverse-engineered WhatsApp Web protocol
2. WhatsApp's infrastructure may route messages through **relay/proxy numbers**
3. The sender ID displayed to recipients can vary

### Does it affect functionality?

No - the POC works correctly:
- Messages are received and processed
- Responses are sent back to the correct recipient
- All business logic (validation, rate limiting, etc.) works as expected

### POC vs Production

| Aspect | POC (whatsmeow) | Production (Official API) |
|--------|-----------------|---------------------------|
| Sender Number | May vary (relay numbers) | Your registered business number |
| Authentication | QR Code scan | Business verification |
| Reliability | Good for testing | Enterprise-grade SLA |
| Cost | Free | Pay per conversation |
| Compliance | Not for production | WhatsApp approved |

### For Production Use

To ensure messages always come from your registered business number, migrate to the **official WhatsApp Business API**:

- [WhatsApp Business Platform](https://business.whatsapp.com/)
- [Cloud API Documentation](https://developers.facebook.com/docs/whatsapp/cloud-api)

The API service in this POC is designed to be reusable - only the bridge component needs to be replaced with an official API integration.

## Production Without Official API (Not Recommended)

It is technically possible to run whatsmeow-based solutions in production without the official WhatsApp Business API. However, this approach comes with significant risks and is **not recommended** for business-critical applications.

### What Would Be Needed

To make this POC more production-ready without the official API:

- **Database**: Migrate from SQLite to PostgreSQL or MySQL for better concurrency
- **Reconnection Logic**: Implement exponential backoff for handling disconnections
- **Outgoing Rate Limiting**: Throttle outgoing messages to avoid detection
- **Session Persistence**: Robust session management across restarts
- **Monitoring**: Add health checks, metrics, and alerting for connection status
- **Multi-instance**: Consider message queuing for horizontal scaling

### Risks and Disclaimers

**Account Ban Risk**: Using unofficial libraries like whatsmeow violates WhatsApp's Terms of Service. Accounts can be permanently banned without warning. This risk increases with:
- High message volumes
- Automated/bot-like behavior patterns
- User reports

**No Official Support**: WhatsApp does not provide support for unofficial integrations. When things break, you're on your own.

**Protocol Changes**: WhatsApp can change their protocol at any time, breaking whatsmeow functionality. Updates may take days or weeks to be released by the community.

**Legal/Compliance**: Unofficial API usage may violate regulations in some jurisdictions, especially for financial or healthcare applications.

**Sender Number Issues**: As documented above, messages may appear to come from relay numbers rather than your authenticated number.

### When to Use the Official API Instead

Choose the official WhatsApp Business API if you need:

- Guaranteed message delivery with SLA
- Consistent sender identity (your business number)
- High message volumes (thousands per day)
- Business-critical reliability
- Compliance with regulations (GDPR, financial services, etc.)
- Official support and documentation

### Disclaimer

This POC is provided for **educational and testing purposes only**. The authors do not endorse using unofficial WhatsApp libraries for production systems or any use that violates WhatsApp's Terms of Service. Use at your own risk.

## License

MIT
