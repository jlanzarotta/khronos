# Read token from file in project root
$GITHUB_TOKEN = (Get-Content ".\GITHUB_TOKEN" -Raw).Trim()

podman run -it --rm -v ${PWD}:/app:Z -w /app -e GITHUB_TOKEN=$GITHUB_TOKEN go-dev
