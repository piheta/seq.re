// Shared application utilities

// Copy to clipboard with visual feedback
function copyToClipboard(text, buttonId) {
    navigator.clipboard.writeText(text).then(() => {
        const button = document.getElementById(buttonId);
        if (!button) return;

        const originalHTML = button.innerHTML;
        button.innerHTML = '<svg class="w-5 h-5 text-dr-green dark:text-dr-green-dark" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>';

        setTimeout(() => {
            button.innerHTML = originalHTML;
        }, 2000);
    }).catch(err => {
        console.error('Failed to copy:', err);
    });
}

// Copy from input field (reads current value including any JS modifications like key fragments)
function copyFromInput(inputId, buttonId) {
    const input = document.getElementById(inputId);
    if (!input) return;

    copyToClipboard(input.value, buttonId);
}

// Copy decrypted text from a global variable (for viewers)
function copyDecryptedText(varName, buttonId) {
    const text = window[varName];
    if (!text) return;

    navigator.clipboard.writeText(text).then(() => {
        const button = document.getElementById(buttonId);
        if (!button) return;

        const originalText = button.textContent;
        button.textContent = 'Copied!';

        setTimeout(() => {
            button.textContent = originalText;
        }, 2000);
    }).catch(err => {
        console.error('Failed to copy:', err);
    });
}

// Common form submission helper for encrypted forms
async function submitEncryptedForm(options) {
    const {
        url,
        data,
        encrypted,
        resultDivId,
        buttonId,
        onSuccess
    } = options;

    const resultDiv = document.getElementById(resultDivId);

    try {
        let key;
        let requestData = { ...data };

        // Handle encryption if needed
        if (encrypted && data.content) {
            key = await generateKey();
            requestData.content = await encrypt(data.content, key);
        } else if (encrypted && data.url) {
            key = await generateKey();
            requestData.url = await encrypt(data.url, key);
        } else if (encrypted && data.data) {
            key = await generateKey();
            requestData.data = await encrypt(data.data, key);
        }

        // Make API request
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'HX-Request': 'true'
            },
            body: JSON.stringify(requestData)
        });

        if (!response.ok) {
            throw new Error(`Request failed: ${response.statusText}`);
        }

        // Get HTML response
        const html = await response.text();
        resultDiv.innerHTML = html;

        // Append key fragment to the input if encrypted
        if (encrypted && key) {
            const input = resultDiv.querySelector('input');
            if (input) {
                const keyFragment = await exportKey(key);
                input.value = input.value + '#' + keyFragment;
            }
        }

        // Call success callback if provided
        if (onSuccess) {
            onSuccess();
        }

    } catch (error) {
        resultDiv.innerHTML = `<div class="mt-4 text-dr-orange dark:text-dr-orange-dark">Error: ${error.message}</div>`;
    }
}

// Dark mode initialization and toggle
function initDarkMode() {
    const savedMode = localStorage.getItem('darkMode');
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    const shouldBeDark = savedMode === 'dark' || (!savedMode && prefersDark);

    if (shouldBeDark) {
        document.documentElement.classList.add('dark');

        const sunIcon = document.getElementById('sun-icon');
        const moonIcon = document.getElementById('moon-icon');

        if (sunIcon && moonIcon) {
            sunIcon.classList.remove('hidden');
            moonIcon.classList.add('hidden');
        }
    }
}

function toggleDarkMode() {
    const html = document.documentElement;
    const isDark = html.classList.toggle('dark');
    localStorage.setItem('darkMode', isDark ? 'dark' : 'light');

    const sunIcon = document.getElementById('sun-icon');
    const moonIcon = document.getElementById('moon-icon');

    if (sunIcon && moonIcon) {
        sunIcon.classList.toggle('hidden');
        moonIcon.classList.toggle('hidden');
    }
}

// Tab management
function setActiveTab(button) {
    const buttons = document.querySelectorAll('.tab-button');
    buttons.forEach(btn => {
        btn.classList.remove('bg-dr-bg-subtle', 'dark:bg-dr-bg-subtle-dark', 'text-dr-text-heading', 'dark:text-dr-text-heading-dark', 'active-tab');
        btn.classList.add('text-dr-text-gray', 'dark:text-dr-text-gray-light');
    });

    button.classList.remove('text-dr-text-gray', 'dark:text-dr-text-gray-light');
    button.classList.add('bg-dr-bg-subtle', 'dark:bg-dr-bg-subtle-dark', 'text-dr-text-heading', 'dark:text-dr-text-heading-dark', 'active-tab');
}

// Initialize on page load
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initDarkMode);
} else {
    initDarkMode();
}
