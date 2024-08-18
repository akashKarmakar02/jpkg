#!/bin/bash

# Create the bin directory if it doesn't exist
mkdir -p ~/.amber/bin

# Download the first binary
curl -L -o ~/.amber/bin/binary1 "https://example.com/path/to/binary1"
chmod +x ~/.amber/bin/binary1

# Download the second binary
curl -L -o ~/.amber/bin/binary2 "https://example.com/path/to/binary2"
chmod +x ~/.amber/bin/binary2

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
