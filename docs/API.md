# WhatsApp Promo Code API Documentation

## Base URL
```
http://localhost:8080/api/v1
```

## Endpoints

### POST /redeem

Redeem a promo code and deliver the reward.

**Request Body:**
```json
{
  "phone_number": "+237612345678",
  "code": "VALID100",
  "message_id": "whatsapp_msg_123",
  "language": "fr",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Success Response (200):**
```json
{
  "status": "success",
  "transaction_id": "txn_abc123def456",
  "reward": {
    "type": "data",
    "amount": "1GB de data mobile",
    "description": "1GB de data mobile"
  },
  "message": {
    "fr": "Felicitations ! 1GB de data mobile de data a ete ajoute a votre compte.",
    "en": "Congratulations! 1GB of mobile data of data has been added to your account."
  }
}
```

**Error Response (200):**
```json
{
  "status": "error",
  "transaction_id": "txn_abc123def456",
  "error_code": "CODE_EXPIRED",
  "message": {
    "fr": "Ce code promo a expire le 01/01/2024.",
    "en": "This promo code expired on 01/01/2024."
  }
}
```

### GET /health

Check API health status.

**Response (200):**
```json
{
  "status": "healthy",
  "db_connected": true,
  "whatsapp_connected": true,
  "uptime_seconds": 3600
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `CODE_VALID` | Code valid, reward delivered |
| `CODE_INVALID` | Code not recognized |
| `CODE_EXPIRED` | Code has expired |
| `CODE_ALREADY_USED` | Code already redeemed |
| `CODE_NOT_ELIGIBLE` | User not eligible for code |
| `RATE_LIMITED` | Too many attempts |
| `REWARD_FAILED` | Failed to deliver reward |
| `FRAUD_BLOCKED` | Transaction blocked for security |
| `SYSTEM_ERROR` | Internal system error |

## Test Codes

| Code | Expected Result |
|------|-----------------|
| `VALID100` | Success - 1GB data |
| `VALID200` | Success - 500 FCFA credit |
| `EXPIRED01` | Error - Code expired |
| `USED001` | Error - Already used |
| `INVALID` | Error - Invalid code |
| Any other | Error - Not recognized |

## Rate Limits

- **Per minute:** 5 requests per phone number
- **Per hour:** 20 requests per phone number

When rate limited, the response will include:
```json
{
  "status": "error",
  "error_code": "RATE_LIMITED",
  "message": {
    "fr": "Trop de tentatives. Veuillez patienter 1 minute(s).",
    "en": "Too many attempts. Please wait 1 minute(s)."
  }
}
```
