#!/bin/sh
set -eu

OUT_DIR="${1:-./secrets}"
FORCE="${FORCE:-0}"

mkdir -p "$OUT_DIR"
umask 077

generate_hex() {
	bytes="$1"
	if command -v openssl >/dev/null 2>&1; then
		openssl rand -hex "$bytes"
		return
	fi

	od -An -tx1 -N "$bytes" /dev/urandom | tr -d ' \n'
	printf '\n'
}

write_key() {
	bytes="$1"
	file="$2"

	if [ -f "$file" ] && [ "$FORCE" != "1" ]; then
		echo "exists: $file"
		return
	fi

	generate_hex "$bytes" > "$file"
	chmod 600 "$file"
	echo "created: $file"
}

jwt_sign_key_file="$OUT_DIR/jwt_sign.key"
encryption_key_file="$OUT_DIR/encryption.key"

write_key 64 "$jwt_sign_key_file"
write_key 32 "$encryption_key_file"

cat <<EOF

Use these paths in .env:
JWT_SIGN_KEY_FILE=$jwt_sign_key_file
ENCRYPTION_KEY_FILE=$encryption_key_file
EOF
