import { fetchPosts, updateTimestamp } from './feed.js';


document.addEventListener('DOMContentLoaded', async () => {
    await fetchPosts();
    await updateTimestamp();
});