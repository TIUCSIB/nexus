#!/bin/bash
# Quick release script for Nexus
# Usage: ./release.sh v1.0.0

set -e

red='\033[0;31m'
green='\033[0;32m'
yellow='\033[0;33m'
cyan='\033[0;36m'
plain='\033[0m'

if [ -z "$1" ]; then
    echo -e "${red}Error: Version tag required${plain}"
    echo "Usage: ./release.sh v1.0.0"
    exit 1
fi

VERSION=$1

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${yellow}Warning: Version format should be v1.0.0${plain}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo -e "${cyan}========================================${plain}"
echo -e "${cyan}       Nexus Release Script             ${plain}"
echo -e "${cyan}========================================${plain}"
echo -e "  Version: ${green}${VERSION}${plain}"
echo ""

# Check git status
if [[ -n $(git status -s) ]]; then
    echo -e "${yellow}[1/5]${plain} Uncommitted changes detected:"
    git status -s
    echo ""
    read -p "Commit all changes? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        read -p "Commit message: " commit_msg
        git add .
        git commit -m "${commit_msg}"
    else
        echo -e "${red}Aborted${plain}"
        exit 1
    fi
else
    echo -e "${green}[1/5]${plain} Working directory clean"
fi

# Push commits
echo -e "${yellow}[2/5]${plain} Pushing commits..."
git push

# Check if tag exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo -e "${red}Error: Tag ${VERSION} already exists${plain}"
    read -p "Delete and recreate? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        git tag -d "$VERSION"
        git push origin ":refs/tags/$VERSION"
    else
        exit 1
    fi
fi

# Create tag
echo -e "${yellow}[3/5]${plain} Creating tag ${VERSION}..."
git tag -a "$VERSION" -m "Release ${VERSION}"

# Push tag
echo -e "${yellow}[4/5]${plain} Pushing tag..."
git push origin "$VERSION"

echo -e "${yellow}[5/5]${plain} Triggering GitHub Actions..."
sleep 2

echo ""
echo -e "${green}========================================${plain}"
echo -e "${green}  Release ${VERSION} initiated!         ${plain}"
echo -e "${green}========================================${plain}"
echo ""
echo -e "GitHub Actions is building your release."
echo -e "Check progress at:"
echo -e "  ${cyan}https://github.com/TIUCSIB/Nexus/actions${plain}"
echo ""
echo -e "Release will be available at:"
echo -e "  ${cyan}https://github.com/TIUCSIB/Nexus/releases/tag/${VERSION}${plain}"
echo ""
