// Alpine.js components and utilities for TextProof

// Clipboard copy function
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        alert('Скопировано в буфер обмена');
    }).catch(err => {
        console.error('Ошибка копирования: ', err);
    });
}

// Format timestamp to local string
function formatTimestamp(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleString('ru-RU');
}

// Calculate SHA-256 hash of a string (using Web Crypto API)
async function calculateSHA256(text) {
    const encoder = new TextEncoder();
    const data = encoder.encode(text);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
    return hashHex;
}

// Alpine.js data components
document.addEventListener('alpine:init', () => {
    // Global store for deposit results
    Alpine.store('depositResult', {
        id: null,
        qrCodeUrl: null,
        badgeUrl: null,
        showResult: false,
        setResult(id, qrCodeUrl, badgeUrl) {
            this.id = id;
            this.qrCodeUrl = qrCodeUrl;
            this.badgeUrl = badgeUrl;
            this.showResult = true;
        },
        clearResult() {
            this.id = null;
            this.qrCodeUrl = null;
            this.badgeUrl = null;
            this.showResult = false;
        }
    });

    // Global store for verification results
    Alpine.store('verificationResult', {
        block: null,
        verified: false,
        error: null,
        setResult(block, verified) {
            this.block = block;
            this.verified = verified;
            this.error = null;
        },
        setError(error) {
            this.error = error;
            this.block = null;
            this.verified = false;
        },
        clearResult() {
            this.block = null;
            this.verified = false;
            this.error = null;
        }
    });
});