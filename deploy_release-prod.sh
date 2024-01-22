#!/bin/sh

BRANCH_NAME="release/prod"

# Abort on errors
set -e

# Prepare the private key from an environment variable
PRIVATE_KEY_PATH="$(mktemp)"
# Extract the header
HEADER=$(echo "$SSH_PRIVATE_KEY" | grep -o -e "-----BEGIN [^-]* PRIVATE KEY-----")
# Extract the footer
FOOTER=$(echo "$SSH_PRIVATE_KEY" | grep -o -e "-----END [^-]* PRIVATE KEY-----")
# Remove the header and footer to isolate the body
BODY=$(echo "$SSH_PRIVATE_KEY" | sed -e "s/$HEADER//" | sed -e "s/$FOOTER//")
# Format the body of the key: Insert a newline every 64 characters
FORMATTED_BODY=$(echo "$BODY" | fold -w 64)
# Combine the formatted parts into one key, adding newlines after header and before footer
FORMATTED_KEY="$HEADER\n$FORMATTED_BODY\n$FOOTER"
# Print the key to key file
echo -e "$FORMATTED_KEY" > "$PRIVATE_KEY_PATH"
# Set 600 permission to private key file
chmod 600 "$PRIVATE_KEY_PATH"

# SSH into the EC2 instance and run npm commands
ssh -o StrictHostKeyChecking=no -i "$PRIVATE_KEY_PATH" "$HOST_ADDRESS" << ENDSSH
  BRANCH_NAME="${BRANCH_NAME}" # This assigns the value of the local variable to the remote variable
  cd ~/workspace/paxintrade/backend

  # Ensure the target branch exists and check it out
  branch_exists_locally=\$(git branch --list \$BRANCH_NAME)
  branch_exists_remotely=\$(git ls-remote --heads origin \$BRANCH_NAME)

  if [[ -z "\$branch_exists_locally" ]]; then
    if [[ -n "\$branch_exists_remotely" ]]; then
      git remote prune origin
      git fetch origin
      git checkout -b \$BRANCH_NAME origin/\$BRANCH_NAME
    else
      echo "Error: The branch '\$BRANCH_NAME' does not exist locally or on 'origin'."
      exit 1
    fi
  else
    git checkout \$BRANCH_NAME
  fi

  # Continue to pull, install, build, and restart
  git pull origin \$BRANCH_NAME
  docker compose down paxintrade-api
  docker rmi paxintrade-api:latest-prod
  docker compose up -d --build
ENDSSH

# Clean up the private key
rm "$PRIVATE_KEY_PATH"