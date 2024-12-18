
/**
 * Feed Manager Module
 *
 * This module handles the fetching and display of social media posts and mapping links to GitHub repository information.
 * It provides functionality to:
 * - Fetch and display posts from a bluesky media feed
 * - Extract and display GitHub repository information from posts
 * - Format timestamps and calculate relative time
 * - Convert URLs in text to clickable links
 * - Create visual cards for GitHub repositories with metadata
 */

const advancedUrlRegex = /(?:(?:https?|ftp):\/\/)?(?:www\.)?(?:[a-zA-Z0-9-]+(?:\.[a-zA-Z]{2,})+)(?:\/[^\s]*)?/g;

function linkifyText(text) {
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

function formatTimeUs(timeUs) {

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

function createRepoCard(repoUrl) {
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


function getTimeAgo(timestamp) {
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

async function fetchPosts() {
    try {
        // Fetch posts
        const response = await fetch('/api/v1/posts');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();

        console.log('Received data:', data);
        let feedHtml = '';

        // Process each post
        for (const post of data) {
            // Extract GitHub URL if it exists
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
                        <strong>Post: <a href="https://bsky.app/profile/${post.Did}/post/${post.Rkey}"
                            style="word-wrap: break-word; word-break: break-all;">${post.Rkey}</a></strong>
                        <small class="text-muted float-end">Posted: ${formatTimeUs(post.TimeUs)} UTC</small>
                    </div>
                    <div class="post-content">
                        ${contentHtml}
                    </div>
                </div>
            `;
        }

        // Update the DOM with posts
        document.getElementById('postContainer').innerHTML = feedHtml;

        // After rendering, fetch repository details for each repo card
        const repoCards = document.querySelectorAll('.repo-card');
        for (const card of repoCards) {
            try {
                const repoUrl = card.querySelector('a').getAttribute('href');
                const [, username, repository] = repoUrl.match(/github\.com\/([^\/]+)\/([^\/\s]+)/);

                // Fetch repository details from GitHub API
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

                document.getElementById(`repo-${username}-${repository}`).innerHTML = repoDetails;
            } catch (error) {
                console.error('Error fetching repository details:', error);
                document.getElementById(`repo-${username}-${repository}`).innerHTML =
                    '<div class="error">Error loading repository details</div>';
            }
        }
    } catch (error) {
        console.error('Error fetching posts:', error);
        document.getElementById('postContainer').innerHTML =
            '<div class="alert alert-danger">Error loading posts. Please try again later.</div>';
    }
}


async function updateTimestamp() {
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

export {
    linkifyText,
    formatTimeUs,
    createRepoCard,
    getTimeAgo,
    fetchPosts,
    updateTimestamp
};