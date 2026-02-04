# Security and E2EE (Draft)

## Goals
- Server never sees plaintext task data.
- Seamless offline use.
- Easy recovery via Recovery Key.

## Keys
- KEK: derived from password via Argon2id.
- DEK: random data key used to encrypt tasks.
- Recovery Key: user-visible key for account recovery.

## Flow
Signup:
- Generate DEK + Recovery Key.
- Encrypt DEK with KEK (store locally).
- Encrypt DEK with Recovery Key and upload to server.

Login:
- Derive KEK from password.
- Decrypt DEK locally.

Password Reset:
- Use Recovery Key to decrypt DEK.
- Re-encrypt DEK with new KEK.

## Notes
- If password and Recovery Key are lost, data cannot be recovered.
- Email recovery is optional and user-controlled.
