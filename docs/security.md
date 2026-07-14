# NetBridge Security Model

## Encryption at Rest

All profile files are encrypted using AES-256-GCM.

### Key Derivation
- Primary key derived from machine ID
- Optional passphrase for additional security
- Key stored in memory only, never written to disk

### File Permissions
- Config directory: `700`
- Profile files: `600`
- Log files: `640`

## Sensitive Data Handling

### Masking
- Passwords masked in logs and output
- Keys masked in `netbridge show` output
- Use `--reveal` flag to show sensitive values

### Shell History
- No credentials written to shell history
- Use `eval $(netbridge env)` for safe variable export

## Network Security

### TLS Verification
- TLS verification enabled by default
- `--allow-insecure` flag for testing only
- Reality protocol support for advanced obfuscation

### Credential Storage
- Credentials never stored in plaintext
- Memory-safe handling with zeroing after use

## Recommendations

1. Run with minimal privileges
2. Use strong passphrases for encryption
3. Regularly rotate credentials
4. Monitor logs for unauthorized access
5. Keep NetBridge updated
