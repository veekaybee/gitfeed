#!/bin/bash

# https://gist.github.com/fritzprix/5d2aa93a1d1e117ed374bb18dbd3bb27
# Get the git diff and save it to a temporary file
git diff --cached > /tmp/git_diff.txt

# If there's no diff, exit
if [ ! -s /tmp/git_diff.txt ]; then
    echo "No staged changes to commit"
    exit 1
fi

# Create the prompt with the diff content
prompt="Given the following git diff, please write a clear and concise git commit message that explains the changes. Focus on WHAT changed and WHY. Use present tense, imperative mood. Keep it under 72 characters for the first line, then add more details if needed after a blank line:\n\n$(cat /tmp/git_diff.txt)"

# Run the prompt through ollama and save the response
ollama run mistral:latest "$prompt" > /tmp/commit_msg.txt

# Show the proposed commit message and ask for confirmation
echo -e "\nProposed commit message:"
echo "------------------------"
cat /tmp/commit_msg.txt
echo "------------------------"
echo -e "\nDo you want to proceed with this commit message? (y/n)"
read answer

if [ "$answer" = "y" ]; then
    # Perform the commit using the generated message
    git commit -F /tmp/commit_msg.txt
    echo "Changes committed successfully!"
else
    echo "Commit canceled"
fi

# Clean up temporary files
rm /tmp/git_diff.txt /tmp/commit_msg.txt
