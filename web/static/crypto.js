// Client-side encryption utilities using Web Crypto API

// Generate a random AES-128 key
async function generateKey() {
    return await crypto.subtle.generateKey(
        { name: "AES-GCM", length: 128 },
        true,
        ["encrypt", "decrypt"]
    );
}

// Encrypt data with AES-GCM
async function encrypt(data, key) {
    const iv = crypto.getRandomValues(new Uint8Array(12)); // 96-bit IV for GCM
    const encoded = new TextEncoder().encode(data);

    const ciphertext = await crypto.subtle.encrypt(
        { name: "AES-GCM", iv: iv },
        key,
        encoded
    );

    // Prepend IV to ciphertext
    const combined = new Uint8Array(iv.length + ciphertext.byteLength);
    combined.set(iv, 0);
    combined.set(new Uint8Array(ciphertext), iv.length);

    // Return base64-encoded result
    return btoa(String.fromCharCode(...combined));
}

// Decrypt data with AES-GCM
async function decrypt(base64Data, key) {
    // Decode base64
    const combined = Uint8Array.from(atob(base64Data), c => c.charCodeAt(0));

    // Extract IV and ciphertext
    const iv = combined.slice(0, 12);
    const ciphertext = combined.slice(12);

    const decrypted = await crypto.subtle.decrypt(
        { name: "AES-GCM", iv: iv },
        key,
        ciphertext
    );

    return new TextDecoder().decode(decrypted);
}

// Export key to base64url format (URL-safe)
async function exportKey(key) {
    const exported = await crypto.subtle.exportKey("raw", key);
    const base64 = btoa(String.fromCharCode(...new Uint8Array(exported)));
    // Convert to URL-safe base64
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

// Import key from base64url format
async function importKey(base64url) {
    // Convert from URL-safe base64
    const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
    // Add padding if needed
    const padding = '='.repeat((4 - base64.length % 4) % 4);
    const raw = Uint8Array.from(atob(base64 + padding), c => c.charCodeAt(0));

    return await crypto.subtle.importKey(
        "raw",
        raw,
        { name: "AES-GCM" },
        true,
        ["encrypt", "decrypt"]
    );
}

// Encrypt file data (for images)
async function encryptFile(fileData, key) {
    const iv = crypto.getRandomValues(new Uint8Array(12));

    const ciphertext = await crypto.subtle.encrypt(
        { name: "AES-GCM", iv: iv },
        key,
        fileData
    );

    // Prepend IV to ciphertext
    const combined = new Uint8Array(iv.length + ciphertext.byteLength);
    combined.set(iv, 0);
    combined.set(new Uint8Array(ciphertext), iv.length);

    return combined;
}

// Decrypt file data (for images)
async function decryptFile(encryptedData, key) {
    // Extract IV and ciphertext
    const iv = encryptedData.slice(0, 12);
    const ciphertext = encryptedData.slice(12);

    return await crypto.subtle.decrypt(
        { name: "AES-GCM", iv: iv },
        key,
        ciphertext
    );
}
