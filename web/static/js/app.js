// Tab switching
function switchTab(tabName) {
    // Hide all tabs
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.remove('active');
    });
    
    // Remove active class from all buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Show selected tab
    document.getElementById(tabName + '-tab').classList.add('active');
    
    // Add active class to clicked button
    event.target.closest('.tab-btn').classList.add('active');
}

// Handle shorten response
function handleShortenResponse(event) {
    const xhr = event.detail.xhr;
    const response = JSON.parse(xhr.responseText);
    
    if (xhr.status === 201) {
        // Success
        const baseUrl = window.location.origin;
        const shortUrl = `${baseUrl}/${response.short_key}`;
        
        document.getElementById('result').innerHTML = `
            <div class="result-card">
                <div class="result-header">
                    <span style="font-size: 2rem;">✅</span>
                    <h3>URL Shortened Successfully!</h3>
                </div>
                
                <div class="result-item">
                    <span class="result-label">Short URL</span>
                    <div class="result-value">
                        <a href="${shortUrl}" class="short-url" target="_blank">${shortUrl}</a>
                        <button class="copy-btn" onclick="copyToClipboard('${shortUrl}', this)">📋 Copy</button>
                    </div>
                </div>
                
                <div class="result-item">
                    <span class="result-label">Short Key</span>
                    <div class="result-value">
                        <span>${response.short_key}</span>
                        <button class="copy-btn" onclick="copyToClipboard('${response.short_key}', this)">📋 Copy</button>
                    </div>
                </div>
                
                <div class="result-item">
                    <span class="result-label">Original URL</span>
                    <div class="result-value">
                        <span>${response.long_url}</span>
                    </div>
                </div>
                
                <div class="result-item">
                    <span class="result-label">Created At</span>
                    <div class="result-value">
                        <span>${new Date(response.created_at).toLocaleString()}</span>
                    </div>
                </div>
                
                ${response.expires_at ? `
                    <div class="result-item">
                        <span class="result-label">Expires At</span>
                        <div class="result-value">
                            <span>${new Date(response.expires_at).toLocaleString()}</span>
                        </div>
                    </div>
                ` : ''}
            </div>
        `;
        
        // Save to recent URLs
        saveRecentURL({
            shortKey: response.short_key,
            shortUrl: shortUrl,
            longUrl: response.long_url,
            createdAt: response.created_at
        });
        
        // Show success notification
        showNotification('URL shortened successfully! 🎉', 'success');
        
        // Clear form
        document.getElementById('long_url').value = '';
        document.getElementById('custom_key').value = '';
        document.getElementById('expires_in').value = '';
        
    } else {
        // Error
        document.getElementById('result').innerHTML = `
            <div class="error-card">
                <div class="error-header">
                    <span style="font-size: 2rem;">❌</span>
                    <h3>Error: ${response.error}</h3>
                </div>
                <p class="error-message">${response.message}</p>
            </div>
        `;
        
        showNotification(`Error: ${response.message}`, 'error');
    }
}

// Handle stats response (legacy, now using displayStatsResult)
function handleStatsResponse(event) {
    // This function is kept for compatibility but main logic is in submitStatsForm
}

// Submit stats form with dynamic URL
function submitStatsForm(event) {
    event.preventDefault();
    const shortKey = document.getElementById('short_key').value.trim();
    if (!shortKey) {
        showNotification('Please enter a short key', 'error');
        return false;
    }
    
    const form = document.getElementById('stats-form');
    const url = `/api/stats/${shortKey}`;
    
    // Trigger HTMX GET request and handle response
    fetch(url)
        .then(response => {
            if (!response.ok) {
                return response.json().then(err => {
                    throw err;
                });
            }
            return response.json();
        })
        .then(data => {
            displayStatsResult(data);
            showNotification('Statistics loaded successfully! 📈', 'success');
        })
        .catch(error => {
            displayStatsError(error);
            showNotification(`Error: ${error.message}`, 'error');
        });
    
    return false;
}

// Display stats result in a nice visual format
function displayStatsResult(response) {
    const baseUrl = window.location.origin;
    const shortUrl = `${baseUrl}/${response.short_key}`;
    
    const expiresDate = response.expires_at ? new Date(response.expires_at) : null;
    const createdDate = new Date(response.created_at);
    const now = new Date();
    const isExpired = expiresDate && expiresDate < now;
    
    // Calculate time remaining
    let timeRemaining = '';
    if (expiresDate) {
        const diff = expiresDate - now;
        if (diff > 0) {
            const hours = Math.floor(diff / (1000 * 60 * 60));
            const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
            timeRemaining = `${hours}h ${minutes}m`;
        } else {
            timeRemaining = 'Expired';
        }
    }
    
    document.getElementById('stats-result').innerHTML = `
        <div class="result-card">
            <div class="result-header">
                <span style="font-size: 2rem;">📊</span>
                <h3>URL Statistics</h3>
            </div>
            
            <div class="result-item">
                <span class="result-label">Short URL</span>
                <div class="result-value">
                    <a href="${shortUrl}" class="short-url" target="_blank">${shortUrl}</a>
                    <button class="copy-btn" onclick="copyToClipboard('${shortUrl}', this)">📋 Copy</button>
                </div>
            </div>
            
            <div class="result-item">
                <span class="result-label">Original URL</span>
                <div class="result-value">
                    <a href="${response.long_url}" class="short-url" target="_blank" style="word-break: break-all;">${response.long_url}</a>
                </div>
            </div>
            
            <div class="stats-grid">
                <div class="stat-box">
                    <span class="stat-value" style="font-size: 2.5rem;">👁️</span>
                    <span class="stat-value">${response.visit_count}</span>
                    <span class="stat-label">Total Visits</span>
                </div>
                
                <div class="stat-box">
                    <span class="stat-value" style="font-size: 2.5rem;">📅</span>
                    <span class="stat-value" style="font-size: 1.2rem;">${createdDate.toLocaleDateString()}</span>
                    <span class="stat-label">Created</span>
                    <span class="stat-label" style="font-size: 0.75rem; opacity: 0.7;">${createdDate.toLocaleTimeString()}</span>
                </div>
                
                ${expiresDate ? `
                    <div class="stat-box ${isExpired ? 'stat-box-expired' : ''}">
                        <span class="stat-value" style="font-size: 2.5rem;">${isExpired ? '⏰' : '⌛'}</span>
                        <span class="stat-value" style="font-size: 1.2rem;">${expiresDate.toLocaleDateString()}</span>
                        <span class="stat-label">${isExpired ? 'Expired' : 'Expires'}</span>
                        ${!isExpired ? `<span class="stat-label" style="font-size: 0.85rem; color: var(--warning-color);">${timeRemaining} left</span>` : ''}
                    </div>
                ` : `
                    <div class="stat-box">
                        <span class="stat-value" style="font-size: 2.5rem;">♾️</span>
                        <span class="stat-value" style="font-size: 1.2rem;">Never</span>
                        <span class="stat-label">No Expiration</span>
                    </div>
                `}
            </div>
            
            ${response.visit_count === 0 ? `
                <div style="margin-top: 20px; padding: 15px; background: var(--surface-light); border-radius: 8px; text-align: center;">
                    <span style="font-size: 1.5rem;">🎯</span>
                    <p style="margin-top: 10px; color: var(--text-secondary);">This URL hasn't been visited yet. Share it to start tracking visits!</p>
                </div>
            ` : ''}
        </div>
    `;
}

// Display stats error
function displayStatsError(error) {
    document.getElementById('stats-result').innerHTML = `
        <div class="error-card">
            <div class="error-header">
                <span style="font-size: 2rem;">❌</span>
                <h3>Error: ${error.error || 'not_found'}</h3>
            </div>
            <p class="error-message">${error.message || 'URL not found or has expired'}</p>
        </div>
    `;
}

// Copy to clipboard
function copyToClipboard(text, button) {
    navigator.clipboard.writeText(text).then(() => {
        const originalText = button.innerHTML;
        button.innerHTML = '✓ Copied!';
        button.classList.add('copied');
        
        setTimeout(() => {
            button.innerHTML = originalText;
            button.classList.remove('copied');
        }, 2000);
        
        showNotification('Copied to clipboard! 📋', 'success');
    }).catch(err => {
        showNotification('Failed to copy to clipboard', 'error');
    });
}

// Show notification
function showNotification(message, type) {
    const notification = document.getElementById('notification');
    notification.textContent = message;
    notification.className = `notification ${type} show`;
    
    setTimeout(() => {
        notification.classList.remove('show');
    }, 3000);
}

// Recent URLs functionality
function saveRecentURL(urlData) {
    let recentUrls = JSON.parse(localStorage.getItem('recentUrls') || '[]');
    
    // Add to beginning
    recentUrls.unshift(urlData);
    
    // Keep only last 5
    recentUrls = recentUrls.slice(0, 5);
    
    localStorage.setItem('recentUrls', JSON.stringify(recentUrls));
    displayRecentURLs();
}

function displayRecentURLs() {
    const recentUrls = JSON.parse(localStorage.getItem('recentUrls') || '[]');
    const section = document.getElementById('recent-urls-section');
    const list = document.getElementById('recent-urls-list');
    
    if (recentUrls.length === 0) {
        section.style.display = 'none';
        return;
    }
    
    section.style.display = 'block';
    
    list.innerHTML = recentUrls.map(url => `
        <div class="recent-url-item">
            <div class="recent-url-info">
                <a href="${url.shortUrl}" class="recent-short-url" target="_blank">${url.shortKey}</a>
                <div class="recent-long-url" title="${url.longUrl}">${url.longUrl}</div>
            </div>
            <div class="recent-url-actions">
                <button class="icon-btn" onclick="copyToClipboard('${url.shortUrl}', this)" title="Copy URL">
                    📋
                </button>
                <button class="icon-btn" onclick="viewStats('${url.shortKey}')" title="View Stats">
                    📊
                </button>
            </div>
        </div>
    `).join('');
}

function viewStats(shortKey) {
    // Switch to stats tab
    document.querySelectorAll('.tab-content').forEach(tab => tab.classList.remove('active'));
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    document.getElementById('stats-tab').classList.add('active');
    document.querySelectorAll('.tab-btn')[1].classList.add('active');
    
    // Fill in short key and trigger search
    document.getElementById('short_key').value = shortKey;
    document.querySelector('#stats-tab form').dispatchEvent(new Event('submit', { bubbles: true }));
}

function clearRecentURLs() {
    localStorage.removeItem('recentUrls');
    displayRecentURLs();
    showNotification('Recent URLs cleared', 'success');
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    displayRecentURLs();
    
    // Pre-process form data before HTMX sends it
    document.body.addEventListener('htmx:configRequest', (event) => {
        if (event.detail.path === '/api/shorten') {
            // Get the original parameters object from json-enc
            const params = event.detail.parameters;
            
            // Remove empty custom_key
            if (!params.custom_key || params.custom_key.trim() === '') {
                delete params.custom_key;
            } else {
                params.custom_key = params.custom_key.trim();
            }
            
            // Convert expires_in to number or remove if empty
            if (!params.expires_in || params.expires_in.toString().trim() === '') {
                delete params.expires_in;
            } else {
                const num = parseInt(params.expires_in);
                if (!isNaN(num) && num > 0) {
                    params.expires_in = num;
                } else {
                    delete params.expires_in;
                }
            }
            
            // Trim long_url
            if (params.long_url) {
                params.long_url = params.long_url.trim();
            }
        }
    });
});
