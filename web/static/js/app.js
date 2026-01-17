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
        const shortUrl = `${baseUrl}/s/${response.short_key}`;

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
            createdAt: response.created_at,
            expiresAt: response.expires_at,
            visitCount: 0 // New URLs start with 0 visits
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
    const url = `/stats/${shortKey}`;

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
    const shortUrl = `${baseUrl}/s/${response.short_key}`;

    const expiresDate = response.expires_at ? new Date(response.expires_at) : null;
    const createdDate = new Date(response.created_at);
    const lastAccessedDate = response.last_accessed_at ? new Date(response.last_accessed_at) : null;
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

    // Calculate time since last access
    let lastAccessInfo = '';
    if (lastAccessedDate) {
        const timeSinceAccess = now - lastAccessedDate;
        const daysSince = Math.floor(timeSinceAccess / (1000 * 60 * 60 * 24));
        const hoursSince = Math.floor(timeSinceAccess / (1000 * 60 * 60));
        const minutesSince = Math.floor(timeSinceAccess / (1000 * 60));

        if (daysSince > 0) {
            lastAccessInfo = `${daysSince} day${daysSince > 1 ? 's' : ''} ago`;
        } else if (hoursSince > 0) {
            lastAccessInfo = `${hoursSince} hour${hoursSince > 1 ? 's' : ''} ago`;
        } else if (minutesSince > 0) {
            lastAccessInfo = `${minutesSince} minute${minutesSince > 1 ? 's' : ''} ago`;
        } else {
            lastAccessInfo = 'Just now';
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

                ${lastAccessedDate ? `
                    <div class="stat-box">
                        <span class="stat-value" style="font-size: 2.5rem;">🔗</span>
                        <span class="stat-value" style="font-size: 1.2rem;">${lastAccessedDate.toLocaleDateString()}</span>
                        <span class="stat-label">Last Accessed</span>
                        <span class="stat-label" style="font-size: 0.8rem; color: var(--text-secondary);">${lastAccessInfo}</span>
                    </div>
                ` : response.visit_count > 0 ? `
                    <div class="stat-box">
                        <span class="stat-value" style="font-size: 2.5rem;">🔗</span>
                        <span class="stat-value" style="font-size: 1.2rem;">Unknown</span>
                        <span class="stat-label">Last Accessed</span>
                    </div>
                ` : ''}

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
    // Try using the modern clipboard API first
    if (navigator.clipboard && window.isSecureContext) {
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
            // Fallback to older method
            fallbackCopyToClipboard(text, button);
        });
    } else {
        // Use fallback method for non-HTTPS contexts
        fallbackCopyToClipboard(text, button);
    }
}

// Fallback copy method for older browsers or non-HTTPS contexts
function fallbackCopyToClipboard(text, button) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    try {
        const successful = document.execCommand('copy');
        if (successful) {
            const originalText = button.innerHTML;
            button.innerHTML = '✓ Copied!';
            button.classList.add('copied');

            setTimeout(() => {
                button.innerHTML = originalText;
                button.classList.remove('copied');
            }, 2000);

            showNotification('Copied to clipboard! 📋', 'success');
        } else {
            throw new Error('Copy command failed');
        }
    } catch (err) {
        showNotification('Failed to copy to clipboard', 'error');
    } finally {
        document.body.removeChild(textArea);
    }
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
        list.innerHTML = '<div class="no-recent-urls">No recent URLs yet. Create your first shortened URL above! 🚀</div>';
        return;
    }

    list.innerHTML = recentUrls.map(url => {
        const createdDate = new Date(url.createdAt);
        const expiresDate = url.expiresAt ? new Date(url.expiresAt) : null;
        const now = new Date();
        const isExpired = expiresDate && expiresDate < now;

        // Calculate time since creation
        const timeDiff = now - createdDate;
        const minutes = Math.floor(timeDiff / 60000);
        const hours = Math.floor(minutes / 60);
        const days = Math.floor(hours / 24);

        let timeAgo;
        if (days > 0) {
            timeAgo = `${days} day${days > 1 ? 's' : ''} ago`;
        } else if (hours > 0) {
            timeAgo = `${hours} hour${hours > 1 ? 's' : ''} ago`;
        } else if (minutes > 0) {
            timeAgo = `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
        } else {
            timeAgo = 'Just now';
        }

        // Calculate expiration info
        let expirationInfo = '';
        if (expiresDate) {
            if (isExpired) {
                expirationInfo = '<span class="expired-badge">⚠️ Expired</span>';
            } else {
                const timeToExpire = expiresDate - now;
                const expireHours = Math.floor(timeToExpire / (1000 * 60 * 60));
                const expireDays = Math.floor(expireHours / 24);

                if (expireDays > 0) {
                    expirationInfo = `<span class="expire-info">⏳ Expires in ${expireDays} day${expireDays > 1 ? 's' : ''}</span>`;
                } else if (expireHours > 0) {
                    expirationInfo = `<span class="expire-info">⏳ Expires in ${expireHours} hour${expireHours > 1 ? 's' : ''}</span>`;
                } else {
                    expirationInfo = '<span class="expire-warning">⚠️ Expires soon</span>';
                }
            }
        }

        return `
            <div class="recent-url-item ${isExpired ? 'expired' : ''}">
                <div class="recent-url-header">
                    <div class="recent-url-main">
                        <a href="${url.shortUrl}" class="recent-short-url" target="_blank">
                            <span class="short-key">${url.shortKey}</span>
                        </a>
                        <div class="recent-long-url" title="${url.longUrl}">
                            ${url.longUrl.length > 60 ? url.longUrl.substring(0, 60) + '...' : url.longUrl}
                        </div>
                    </div>
                    <div class="recent-url-actions">
                        <button class="icon-btn" onclick="copyToClipboard('${url.shortUrl}', this)" title="Copy Short URL">
                            📋
                        </button>
                        <button class="icon-btn" onclick="viewStats('${url.shortKey}')" title="View Statistics">
                            📊
                        </button>
                        <button class="icon-btn" onclick="removeRecentURL('${url.shortKey}')" title="Remove from Recent">
                            🗑️
                        </button>
                    </div>
                </div>
                <div class="recent-url-meta">
                    <span class="created-time">📅 Created ${timeAgo}</span>
                    <span class="visit-count">👁️ ${url.visitCount || 0} visits</span>
                    ${expirationInfo}
                </div>
            </div>
        `;
    }).join('');
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

function removeRecentURL(shortKey) {
    let recentUrls = JSON.parse(localStorage.getItem('recentUrls') || '[]');
    recentUrls = recentUrls.filter(url => url.shortKey !== shortKey);
    localStorage.setItem('recentUrls', JSON.stringify(recentUrls));
    displayRecentURLs();
    showNotification('URL removed from recent list', 'success');
}

// Update visit counts for recent URLs
async function updateRecentURLStats() {
    const recentUrls = JSON.parse(localStorage.getItem('recentUrls') || '[]');
    if (recentUrls.length === 0) return;

    let hasUpdates = false;

    for (let url of recentUrls) {
        try {
            const response = await fetch(`/stats/${url.shortKey}`);
            if (response.ok) {
                const stats = await response.json();
                if (stats.visit_count !== url.visitCount) {
                    url.visitCount = stats.visit_count;
                    hasUpdates = true;
                }
            }
        } catch (error) {
            // Silently ignore errors
        }
    }

    if (hasUpdates) {
        localStorage.setItem('recentUrls', JSON.stringify(recentUrls));
        displayRecentURLs();
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    displayRecentURLs();

    // Update visit counts every 30 seconds
    setInterval(updateRecentURLStats, 30000);

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

// Clean Stats functionality
function loadCleanupStats() {
    const statsDiv = document.getElementById('cleanup-stats');
    statsDiv.innerHTML = '<div class="loading">Loading cleanup statistics...</div>';

    fetch('/api/admin/cleanup/stats')
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                throw new Error(data.message);
            }
            displayCleanupStats(data);
        })
        .catch(error => {
            statsDiv.innerHTML = `
                <div class="error-card">
                    <div class="error-header">
                        <span>❌</span>
                        <h3>Error Loading Stats</h3>
                    </div>
                    <p>${error.message}</p>
                </div>
            `;
        });
}

function displayCleanupStats(stats) {
    const statusColor = stats.is_running ? 'var(--success-color)' : 'var(--text-secondary)';
    const statusText = stats.is_running ? 'Running' : 'Stopped';
    const statusIcon = stats.is_running ? '🟢' : '🔴';

    const lastCleanupTime = stats.last_cleanup_time ?
        new Date(stats.last_cleanup_time).toLocaleString() :
        'Never';

    document.getElementById('cleanup-stats').innerHTML = `
        <div class="stats-grid">
            <div class="stat-item">
                <div class="stat-header">
                    <span class="stat-icon">🔄</span>
                    <span class="stat-label">Service Status</span>
                </div>
                <div class="stat-value" style="color: ${statusColor}">
                    ${statusIcon} ${statusText}
                </div>
            </div>

            <div class="stat-item">
                <div class="stat-header">
                    <span class="stat-icon">🗑️</span>
                    <span class="stat-label">Total Cleaned</span>
                </div>
                <div class="stat-value">${stats.total_cleaned.toLocaleString()}</div>
            </div>

            <div class="stat-item">
                <div class="stat-header">
                    <span class="stat-icon">📦</span>
                    <span class="stat-label">Last Batch Size</span>
                </div>
                <div class="stat-value">${stats.last_batch_size}</div>
            </div>

            <div class="stat-item">
                <div class="stat-header">
                    <span class="stat-icon">✅</span>
                    <span class="stat-label">Successful Runs</span>
                </div>
                <div class="stat-value">${stats.successful_runs}</div>
            </div>

            <div class="stat-item">
                <div class="stat-header">
                    <span class="stat-icon">❌</span>
                    <span class="stat-label">Failed Runs</span>
                </div>
                <div class="stat-value">${stats.failed_runs}</div>
            </div>

            <div class="stat-item">
                <div class="stat-header">
                    <span class="stat-icon">⏱️</span>
                    <span class="stat-label">Avg Cleanup Time</span>
                </div>
                <div class="stat-value">${Math.round(stats.average_cleanup_ms)}ms</div>
            </div>

            <div class="stat-item span-2">
                <div class="stat-header">
                    <span class="stat-icon">🕒</span>
                    <span class="stat-label">Last Cleanup</span>
                </div>
                <div class="stat-value">${lastCleanupTime}</div>
            </div>
        </div>
    `;
}

function triggerManualCleanup() {
    const batchSize = document.getElementById('batch_size').value || 1000;
    const resultDiv = document.getElementById('cleanup-result');
    const button = document.getElementById('manual-cleanup');

    // Disable button and show loading
    button.disabled = true;
    button.innerHTML = '<span class="btn-icon">⏳</span>Running Cleanup...';

    resultDiv.innerHTML = '<div class="loading">Running manual cleanup...</div>';

    fetch('/api/admin/cleanup/manual', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            batch_size: parseInt(batchSize)
        })
    })
        .then(response => response.json())
        .then(data => {
            if (data.error) {
                throw new Error(data.message);
            }

            resultDiv.innerHTML = `
                <div class="result-card">
                    <div class="result-header">
                        <span style="font-size: 2rem;">✅</span>
                        <h3>Cleanup Completed Successfully!</h3>
                    </div>

                    <div class="result-item">
                        <span class="result-label">URLs Cleaned</span>
                        <div class="result-value">
                            <span>${data.cleaned_count} expired URLs removed</span>
                        </div>
                    </div>

                    <div class="result-item">
                        <span class="result-label">Cleanup Duration</span>
                        <div class="result-value">
                            <span>${Math.round(data.duration_ms)}ms</span>
                        </div>
                    </div>

                    <div class="result-item">
                        <span class="result-label">Batch Size</span>
                        <div class="result-value">
                            <span>${data.batch_size}</span>
                        </div>
                    </div>
                </div>
            `;

            showNotification(`Successfully cleaned ${data.cleaned_count} expired URLs! 🧹`, 'success');

            // Refresh stats after cleanup
            setTimeout(loadCleanupStats, 1000);
        })
        .catch(error => {
            resultDiv.innerHTML = `
                <div class="error-card">
                    <div class="error-header">
                        <span>❌</span>
                        <h3>Cleanup Failed</h3>
                    </div>
                    <p>${error.message}</p>
                </div>
            `;

            showNotification('Cleanup operation failed', 'error');
        })
        .finally(() => {
            // Re-enable button
            button.disabled = false;
            button.innerHTML = '<span class="btn-icon">🧹</span>Run Manual Cleanup';
        });
}

// Auto-load cleanup stats when clean tab is opened
document.addEventListener('DOMContentLoaded', function() {
    // Override the switchTab function to include cleanup stats loading
    const originalSwitchTab = window.switchTab;
    window.switchTab = function(tabName) {
        originalSwitchTab(tabName);

        if (tabName === 'clean') {
            loadCleanupStats();
        }
    };
});
