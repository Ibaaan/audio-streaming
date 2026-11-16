#!/usr/bin/env bash
# Run from any directory: /path/to/audio-streaming/ci-local.sh
# Requires Go (see backend/go.mod), golangci-lint v2.15.0, make, Docker,
# ffmpeg, ffprobe, and curl. Docker must also be available to testcontainers.
# Optional: GO_BIN=/absolute/path/to/go LINT_BIN=/absolute/path/to/golangci-lint
# No .env is loaded; all integration configuration uses disposable services.
set -Eeuo pipefail

project_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
cd "$project_dir/backend"

fail() {
    printf 'Error: %s\n' "$*" >&2
    exit 1
}

if [[ $# -gt 0 ]]; then
    if [[ $# -eq 1 && ( $1 == --help || $1 == -h ) ]]; then
        sed -n '2,6s/^# //p' "$project_dir/ci-local.sh"
        exit 0
    fi
    fail "No arguments expected. Use --help for requirements."
fi

# Make uses tools on PATH, so overrides must also apply to its subprocesses.
for override in "${GO_BIN:-}" "${LINT_BIN:-}"; do
    if [[ -n "$override" ]]; then
        [[ "$override" == /* && -x "$override" ]] ||
            fail "Tool overrides must be absolute paths to executable files."
        export PATH="$(dirname -- "$override"):$PATH"
    fi
done

for tool in go golangci-lint make docker ffmpeg ffprobe curl; do
    command -v "$tool" >/dev/null || fail "Required command not found: $tool"
done
verify_tool() {
    local tool=$1 fallback=$2 output
    if output="$("$tool" version 2>&1)"; then
        printf '%s\n' "$output"
        return
    fi
    # Some development environments cannot run snap-confine. Use the installed
    # executable directly when the selected command is that Snap launcher.
    if [[ "$(command -v "$tool")" == "/snap/bin/$tool" && -x "$fallback" ]]; then
        printf 'Using installed executable: %s\n' "$fallback"
        export PATH="$(dirname -- "$fallback"):$PATH"
        "$tool" version || fail "Cannot start $tool using $fallback."
    else
        printf '%s\n' "$output" >&2
        fail "Cannot start $tool. Set GO_BIN or LINT_BIN to a working executable."
    fi
}
verify_tool go /snap/go/current/bin/go
verify_tool golangci-lint /snap/golangci-lint/current/golangci-lint
docker info >/dev/null || fail "Cannot connect to Docker. Start the container runtime first."

# Honor an explicitly selected Docker host; otherwise align testcontainers with
# the Docker CLI's context (including local Podman Docker-compatible sockets).
if [[ -z "${DOCKER_HOST:-}" ]]; then
    export DOCKER_HOST="$(docker context inspect --format '{{.Endpoints.docker.Host}}')"
fi

printf '\nRunning dependency, lint, race-unit, and vulnerability checks...\n'
make check

work_dir=""
container_id=""
cleanup() {
    local status=$?
    trap - EXIT
    if [[ -n "$container_id" ]]; then
        docker rm --force "$container_id" >/dev/null 2>&1 ||
            printf 'Warning: could not remove temporary container %s\n' "$container_id" >&2
    fi
    if [[ -n "$work_dir" ]]; then
        rm -rf -- "$work_dir"
    fi
    exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

work_dir="$(mktemp -d "${TMPDIR:-/tmp}/audio-streaming-ci.XXXXXXXX")"
cat > "$work_dir/s3.json" <<'JSON'
{"identities":[{"name":"local-ci","credentials":[{"accessKey":"ci-access","secretKey":"ci-secret"}],"actions":["Admin","Read","Write","List","Tagging"]}]}
JSON

printf '\nStarting disposable SeaweedFS...\n'
container_id="$(docker run --detach --rm \
    --publish 127.0.0.1::8333 \
    --volume "$work_dir/s3.json:/etc/seaweedfs/s3.json:ro,z" \
    chrislusf/seaweedfs:4.48 \
    server -dir=/data -s3 -s3.config=/etc/seaweedfs/s3.json)"

export TEST_POSTGRES_USER=test
export TEST_POSTGRES_PASSWORD=test
export TEST_POSTGRES_DB=test
export MINIO_ENDPOINT="$(docker port "$container_id" 8333/tcp)"
export MINIO_ACCESS_KEY=ci-access
export MINIO_SECRET_KEY=ci-secret
export MINIO_USE_SSL=false
export MINIO_BUCKET=ci

ready=false
deadline=$((SECONDS + 60))
while (( SECONDS < deadline )); do
    if [[ "$(docker inspect --format '{{.State.Running}}' "$container_id")" != true ]]; then
        docker logs "$container_id" >&2 || true
        fail "Temporary SeaweedFS container stopped before becoming ready."
    fi
    # An unauthenticated S3 request may return 403; that still confirms readiness.
    if curl --silent --output /dev/null --connect-timeout 1 --max-time 2 \
        "http://$MINIO_ENDPOINT/"; then
        ready=true
        break
    fi
    sleep 1
done
if [[ "$ready" != true ]]; then
    docker logs "$container_id" >&2 || true
    fail "SeaweedFS did not respond within 60 seconds."
fi

printf '\nRunning integration tests with race detection...\n'
make integration
printf '\nAll local CI checks passed.\n'
