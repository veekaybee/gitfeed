// feed.js

function isGithubRepo(uri) {
    const githubRegex = /(https:\/\/github\.com\/[^\/]+\/[^\/\s]+)/g;
    return uri ? uri.match(githubRegex) : null;
}

function getUserAndRepoFromURL(url) {
    var match = url.match(/https:\/\/github\.com\/([^/]+)\/([^\/]+)/);
    if (match) {
        return [match[1], match[2]];
    } else {
        throw new Error("Invalid GitHub URL");
    }
}

async function hydratePost(post, repoUrl) {
    try {
        const [username, repository] = getUserAndRepoFromURL(repoUrl);
        console.log(username, repository);

        // Fetch repository details
        const repoResponse = await fetch(`/github/${username}/${repository}`);
        if (!repoResponse.ok) {
            throw new Error(`HTTP error! status: ${repoResponse.status}`);
        }
        const repoData = await repoResponse.json();

        return `
            <div class="post-card">
                <div class="post-header">
                    <strong>Post: <a href="https://bsky.app/profile/${post.Did}/post/${post.Rkey}">${post.Rkey}</a></strong>
                    <small class="text-muted float-end">Posted: ${formatTimeUs(post.TimeUs)} UTC</small>
                </div>
                <div class="post-content">
                <div class="repo-info">
                <div class="repo-header">
                    <i class="bi bi-github"></i>
                    <p>${repoData.description || 'No description available'}</p>
                    <div class="repo-stats">
                        <span><i class="bi bi-star"></i> ${repoData.stargazers_count}</span>
                        <span><i class="bi bi-diagram-2"></i> ${repoData.forks_count}</span>
                        <span>${repoData.language || 'Unknown language'}</span>
                    </div>
                </div>
                </div>
            </div>`;
    } catch (error) {
        console.error('Error processing repository:', repoUrl, error);
    }
}

function renderSkeletonPost(post,uri) {
    return `
        <div class="post-card">
            <div class="post-header">
                <strong>Post: <a href="https://bsky.app/profile/${post.Did}/post/${post.Rkey}">${post.Rkey}</a></strong>
                <small class="text-muted float-end">Posted: ${formatTimeUs(post.TimeUs)} UTC</small>
            </div>
            <div class="post-content">
                ${linkifyText(uri || '')}
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

export function linkifyText(text) {
    const advancedUrlRegex = /(?:(?:https?|ftp):\/\/)?(?:www\.)?(?:[a-zA-Z0-9-]+(?:\.[a-zA-Z]{2,})+)(?:\/[^\s]*)?/g;
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

function formatNumber(num) {
    if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'k';
    }
    return num.toString();
}

function escapeHtml(unsafe) {
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
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
        // Clear container after "Loading posts"
        container.innerHTML = '';

        const renderedPosts = [];

        for (const post of posts) {
            container.insertAdjacentHTML('beforeend', renderSkeletonPost(post,post.URI));
            const githubMatch = isGithubRepo(post.URI);
            if (githubMatch) {
                try {
                    const hydratedPost = await hydratePost(post, githubMatch[0]);


                } catch (error) {
                    console.error('Error fetching GitHub data for post:', error);
                }
            }
        }

        container.insertAdjacentHTML('beforeend', renderedPosts);
    } catch (error) {
        console.error('Error fetching posts:', error);
        container.innerHTML = '<div class="alert alert-danger">Error loading posts. Please try again later.</div>';
    }
}