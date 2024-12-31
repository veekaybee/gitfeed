import { fetchPosts, updateTimestamp } from './feed.js';


console.log('Main.js loaded');

document.addEventListener('DOMContentLoaded', async () => {
    console.log('DOM Content Loaded');
    try {
        await fetchPosts();
        await updateTimestamp();
    } catch (error) {
        console.error('Error in main initialization:', error);
    }
});