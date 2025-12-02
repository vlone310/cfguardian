#!/bin/bash

# Setup Git hooks for GoConfig Guardian

set -e

HOOKS_DIR=".git/hooks"
SCRIPTS_DIR="scripts"

echo "🔧 Setting up Git hooks..."

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "❌ Not a git repository. Please run 'git init' first."
    exit 1
fi

# Create pre-commit hook
echo "📝 Creating pre-commit hook..."
cat > "$HOOKS_DIR/pre-commit" << 'EOF'
#!/bin/bash
./scripts/pre-commit.sh
EOF

chmod +x "$HOOKS_DIR/pre-commit"

echo "✅ Git hooks installed successfully!"
echo "   Pre-commit hook will run: formatting, linting, and tests"
echo ""
echo "To skip hooks (not recommended), use: git commit --no-verify"

