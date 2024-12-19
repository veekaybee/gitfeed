// feed.js

function isGithubRepo(uri) {
    const githubRegex = /(https:\/\/github\.com\/[^\/]+\/[^\/\s]+)/g;
    return uri ? uri.match(githubRegex) : null;
}

function renderContent(uri) {
    const githubMatch = isGithubRepo(uri);
    if (githubMatch) {
        return createRepoCard(githubMatch[0]);
    }
    return linkifyText(uri || '');
}

function renderPost(post) {
    return `
        <div class="post-card">
            <div class="post-header">
                <strong>Post: <a href="https://bsky.app/profile/${post.Did}/post/${post.Rkey}">${post.Rkey}</a></strong>
                <small class="text-muted float-end">Posted: ${formatTimeUs(post.TimeUs)} UTC</small>
            </div>
            <div class="post-content">
                ${renderContent(post.URI)}
            </div>
        </div>
    `;
}

export function formatTimeUs(timeUs) {
    const timeMs = Math.floor(timeUs / 1000);
    const date = new Date(timeMs);

    const options = {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: true
    };

    return date.toLocaleString('en-US', options);
}

export function getTimeAgo(timestamp) {
    const now = new Date();
    const past = new Date(timestamp);
    const diffInMinutes = Math.floor((now - past) / (1000 * 60));

    if (diffInMinutes < 1) {
        return 'just now';
    } else if (diffInMinutes === 1) {
        return '1 minute ago';
    } else if (diffInMinutes < 60) {
        return `${diffInMinutes} minutes ago`;
    } else if (diffInMinutes < 1440) {
        const hours = Math.floor(diffInMinutes / 60);
        return `${hours} ${hours === 1 ? 'hour' : 'hours'} ago`;
    } else {
        const days = Math.floor(diffInMinutes / 1440);
        return `${days} ${days === 1 ? 'day' : 'days'} ago`;
    }
}


const advancedUrlRegex = /(?:(?:https?|ftp):\/\/)?(?:www\.)?(?:[a-zA-Z0-9-]+(?:\.[a-zA-Z]{2,})+)(?:\/[^\s]*)?/g;
export function linkifyText(text) {
    return text.replace(advancedUrlRegex, url => {
        let href = url;
        if (!url.startsWith('http')) {
            href = 'https://' + url;
        }

        // Extract everything after github.com/
        let displayText = url;
        if (url.includes('github.com/')) {
            displayText = url.split('github.com/')[1];
        }

        return `<a href="${href}" target="_blank" rel="noopener noreferrer" style="word-wrap: break-word; word-break: break-all; max-width: 100%; display: inline-block;">${displayText}</a>`;
    });
}

export function createRepoCard(repoUrl) {
    // Extract username and repository from GitHub URL
    const githubRegex = /github\.com\/([^\/]+)\/([^\/\s]+)/;
    const match = repoUrl.match(githubRegex);

    // Return original URL if not a GitHub repo
    if (!match) return repoUrl;
    const [, username, repository] = match;

    return `
        <div class="repo-card">
            <div class="repo-header">
                <i class="bi bi-github"></i>
                <a href="${repoUrl}" target="_blank" rel="noopener noreferrer">
                    ${username}/${repository}
                </a>
            </div>
            <div class="repo-details" id="repo-${username}-${repository}">
                <div class="loading">Loading repository details...</div>
            </div>
        </div>
    `;
}

export async function updateTimestamp() {
    try {
        const response = await fetch('/api/v1/timestamp');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        const timestamp = new Date(parseInt(data.timestamp));
        console.log('Received timestamp:', timestamp);

        const formattedTime = timestamp.toLocaleString();
        const xTimeAgo = getTimeAgo(formattedTime);
        document.getElementById('lastUpdated').textContent = 'Last updated: ' + xTimeAgo;

    } catch (error) {
        console.error('Error fetching timestamp:', error);
        document.getElementById('lastUpdated').textContent = 'Last updated: Error loading timestamp';
    }
}

export async function fetchPosts() {
    const container = document.getElementById('postContainer');
    container.innerHTML = '<div class="loading">Loading posts...</div>';

    try {
        console.log('Fetching new posts...');
        const response = await fetch('/api/v1/posts');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const posts = await response.json();

        // Clear container
        container.innerHTML = '';

        // Immediately render all posts from database
        posts.forEach(post => {
            container.insertAdjacentHTML('beforeend', renderPost(post));
        });

        // Process GitHub details in parallel without waiting
        posts.forEach(post => {
            const githubRegex = /(https:\/\/github\.com\/[^\/]+\/[^\/\s]+)/g;
            const githubMatch = post.URI ? post.URI.match(githubRegex) : null;

            if (githubMatch) {
                processGithubRepo(githubMatch[0])
                    .then(githubData => {
                        if (githubData) {
                            const postElement = container.querySelector(`[data-post-id="${post.id}"]`);
                            updatePostWithGithubData(postElement, githubData);
                        }
                    })
                    .catch(error => {
                        console.error(`Error fetching GitHub data for ${githubMatch[0]}:`, error);
                        const postElement = container.querySelector(`[data-post-id="${post.id}"]`);
                        if (postElement) {
                            const githubInfo = postElement.querySelector('.github-info');
                            if (githubInfo) {
                                githubInfo.innerHTML = '<div class="text-sm text-gray-500">Unable to load GitHub data</div>';
                            }
                        }
                    });
            }
        });

    } catch (error) {
        console.error('Error fetching posts:', error);
        container.innerHTML =
            '<div class="alert alert-danger">Error loading posts. Please try again later.</div>';
    }
}

// Helper function to update post with GitHub data
function updatePostWithGithubData(postElement, githubData) {
    if (!postElement) return;

    const githubInfo = postElement.querySelector('.github-info');
    if (githubInfo) {
        if (githubData.error) {
            githubInfo.innerHTML = '<div class="text-sm text-gray-500">Unable to load GitHub data</div>';
            return;
        }

        githubInfo.innerHTML = `
            <div class="github-stats flex gap-4 text-sm text-gray-600">
                <span title="Stars">⭐ ${formatNumber(githubData.stars)}</span>
                <span title="Forks">🔄 ${formatNumber(githubData.forks)}</span>
                <span title="Watchers">👀 ${formatNumber(githubData.watchers)}</span>
            </div>
            ${githubData.description ?
            `<div class="github-description text-sm mt-2">${escapeHtml(githubData.description)}</div>`
            : ''
        }
        `;
    }
}

// Helper function to format numbers (e.g., 1000 -> 1k)
function formatNumber(num) {
    if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'k';
    }
    return num.toString();
}

// Helper function to escape HTML to prevent XSS
function escapeHtml(unsafe) {
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

async function processGithubRepo(repoUrl) {
    try {
        const [, username, repository] = repoUrl.match(/github\.com\/([^\/]+)\/([^\/\s]+)/);

        // Fetch repository details
        const repoResponse = await fetch(`https://api.github.com/repos/${username}/${repository}`);
        if (!repoResponse.ok) {
            throw new Error(`HTTP error! status: ${repoResponse.status}`);
        }

        const repoData = await repoResponse.json();
        const repoDetails = `
            <div class="repo-info">
                <p>${repoData.description || 'No description available'}</p>
                <div class="repo-stats">
                    <span><i class="bi bi-star"></i> ${repoData.stargazers_count}</span>
                    <span><i class="bi bi-diagram-2"></i> ${repoData.forks_count}</span>
                    <span>${repoData.language || 'Unknown language'}</span>
                </div>
            </div>
        `;

        // Update the repo details
        const repoElement = document.getElementById(`repo-${username}-${repository}`);
        if (repoElement) {
            repoElement.innerHTML = repoDetails;
        }
    } catch (error) {
        console.error('Error processing repository:', repoUrl, error);
        const [, username, repository] = repoUrl.match(/github\.com\/([^\/]+)\/([^\/\s]+)/);
        const repoElement = document.getElementById(`repo-${username}-${repository}`);
        if (repoElement) {
            repoElement.innerHTML = '<div class="error">Error loading repository details</div>';
        }
    }
}