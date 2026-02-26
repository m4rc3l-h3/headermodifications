sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://dl.yarnpkg.com/debian/pubkey.gpg | sudo gpg --dearmor -o /etc/apt/keyrings/yarn.gpg

echo "deb [signed-by=/etc/apt/keyrings/yarn.gpg] https://dl.yarnpkg.com/debian stable main" | sudo tee /etc/apt/sources.list.d/yarn.list > /dev/null

sudo apt-get update
sudo apt-get install -y --no-install-recommends make curl
sudo rm -rf /var/lib/apt/lists/*
/bin/bash -lc 'set -e; go install github.com/go-delve/delve/cmd/dlv@latest; go mod tidy'