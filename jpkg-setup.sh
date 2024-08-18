#!/bin/bash

# Create the bin directory if it doesn't exist
mkdir -p ~/.amber/bin

# Download the first binary
curl -L -o ~/.amber/bin/jpkg "https://github.com/akashKarmakar02/jpkg/releases/download/0.1.0-ALPHA02/jpkg"
chmod +x ~/.amber/bin/jpkg

# Download the second binary
curl -L -o ~/.amber/bin/jpx "https://github.com/akashKarmakar02/jpkg/releases/download/0.1.0-ALPHA02/jpx"
chmod +x ~/.amber/bin/jpx

# Add ~/.amber/bin to PATH in .zshrc (for macOS)
echo 'export PATH="$HOME/.amber/bin:$PATH"' >> ~/.zshrc

# Add ~/.amber/bin to PATH in .bashrc (for Windows via WSL)
echo 'export PATH="$HOME/.amber/bin:$PATH"' >> ~/.bashrc

# Source the appropriate file
if [[ "$SHELL" == *"zsh"* ]]; then
    source ~/.zshrc
else
    source ~/.bashrc
fi

echo "Binaries downloaded and PATH updated successfully."
