// feed.js

/**
 * Saves posts data to localStorage
 * @param {Array} data - Array of post data to cache
 */
function cachePosts(data) {
    try {
        localStorage.setItem('cachedPosts', JSON.stringify({
            timestamp: Date.now(),
            posts: data
        }));
    } catch (error) {
        console.error('Error caching posts:', error);
    }
}

/**
 * Retrieves cached posts from localStorage
 * @returns {Array|null} Array of cached posts or null if no cache exists
 */
function getCachedPosts() {
    try {
        const cached = localStorage.getItem('cachedPosts');
        if (!cached) return null;

        const { timestamp, posts } = JSON.parse(cached);
        // Optional: Check if cache is too old (e.g., more than 1 hour)
        if (Date.now() - timestamp > 3600000) {
            localStorage.removeItem('cachedPosts');
            return null;
        }

        return posts;
    } catch (error) {
        console.error('Error retrieving cached posts:', error);
        return null;
    }
}

/**
 * Renders posts to the container
 * @param {Array} posts - Array of posts to render
 */
function renderPosts(posts) {
    let feedHtml = '';
    for (const post of posts) {
        const githubRegex = /(https:\/\/github\.com\/[^\/]+\/[^\/\s]+)/g;
        const githubMatch = post.URI ? post.URI.match(githubRegex) : null;

        let contentHtml;
        if (githubMatch) {
            contentHtml = createRepoCard(githubMatch[0]);
        } else {
            contentHtml = linkifyText(post.URI || '');
        }

        feedHtml += `
            <div class="post-card">
                <div class="post-header">
                    <strong>Post: <a href="https://bsky.app/profile/${post.Did}/post/${post.Rkey}">${post.Rkey}</a></strong>
                    <small class="text-muted float-end">Posted: ${formatTimeUs(post.TimeUs)} UTC</small>
                </div>
                <div class="post-content">
                    ${contentHtml}
                </div>
            </div>
        `;
    }
    document.getElementById('postContainer').innerHTML = feedHtml;
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

    if (!match) return repoUrl; // Return original URL if not a GitHub repo

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
    const cachedPosts = getCachedPosts();
    if (cachedPosts) {
        console.log('Found cached posts:', cachedPosts);
        renderPosts(cachedPosts);
    } else {
        console.log('No cached posts found');
        document.getElementById('postContainer').innerHTML =
            '<div class="loading">Loading posts...</div>';
    }

    try {
        console.log('Fetching new posts...');
        const response = await fetch('/api/v1/posts');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();

        // Cache the new data
        cachePosts(data);

        // Render the new posts
        renderPosts(data);

        // After rendering, fetch repository details for each repo card
        const repoCards = document.querySelectorAll('.repo-card');
        for (const card of repoCards) {
            try {
                const repoUrl = card.querySelector('a').getAttribute('href');
                const match = repoUrl.match(/github\.com\/([^\/]+)\/([^\/\s]+)/);
                if (!match) {
                    continue;
                }
                const [, username, repository] = match;

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

                const repoElement = document.getElementById(`repo-${username}-${repository}`);
                if (repoElement) {
                    repoElement.innerHTML = repoDetails;
                }
            } catch (error) {
                console.error('Error fetching repository details:', error);
                // Only try to update the error message if we have the username and repository
                if (typeof username !== 'undefined' && typeof repository !== 'undefined') {
                    const repoElement = document.getElementById(`repo-${username}-${repository}`);
                    if (repoElement) {
                        repoElement.innerHTML = '<div class="error">Error loading repository details</div>';
                    }
                }
            }
        }
    } catch (error) {
        console.error('Error fetching posts:', error);

        // If we have cached content, keep showing it with an error message
        if (cachedPosts) {
            document.getElementById('postContainer').insertAdjacentHTML('beforebegin',
                '<div class="alert alert-warning">Error loading new posts. Showing cached content.</div>');
        } else {
            document.getElementById('postContainer').innerHTML =
                '<div class="alert alert-danger">Error loading posts. Please try again later.</div>';
        }
    }
}