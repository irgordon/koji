#!/usr/bin/env sh
set -eu

artifact_dir="${1:-build/release}"

fail() {
	printf '%s\n' "operator smoke failed: $1" >&2
	exit 1
}

require_command() {
	command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

require_file() {
	[ -f "$1" ] || fail "missing required file: $1"
}

wait_for_file_socket() {
	socket_path="$1"
	for _ in $(seq 1 50); do
		[ -S "$socket_path" ] && return 0
		sleep 0.2
	done
	fail "agent socket did not appear"
}

wait_for_http() {
	base_url="$1"
	for _ in $(seq 1 80); do
		if curl --fail --silent "$base_url/healthz" >/dev/null 2>&1; then
			return 0
		fi
		sleep 0.25
	done
	fail "kojid did not become healthy"
}

json_field() {
	jq -r "$1"
}

csrf_header() {
	printf '%s' "X-CSRF-Token: $1"
}

api_post_json() {
	cookie_file="$1"
	csrf="$2"
	url="$3"
	body="$4"
	curl --fail --silent \
		--cookie "$cookie_file" \
		--cookie-jar "$cookie_file" \
		--header "Content-Type: application/json" \
		--header "$(csrf_header "$csrf")" \
		--data "$body" \
		"$url"
}

api_post_empty() {
	cookie_file="$1"
	csrf="$2"
	url="$3"
	curl --fail --silent \
		--cookie "$cookie_file" \
		--cookie-jar "$cookie_file" \
		--header "$(csrf_header "$csrf")" \
		--request POST \
		"$url"
}

api_get() {
	cookie_file="$1"
	url="$2"
	curl --fail --silent \
		--cookie "$cookie_file" \
		--cookie-jar "$cookie_file" \
		"$url"
}

api_delete() {
	cookie_file="$1"
	csrf="$2"
	url="$3"
	curl --fail --silent \
		--cookie "$cookie_file" \
		--cookie-jar "$cookie_file" \
		--header "$(csrf_header "$csrf")" \
		--request DELETE \
		"$url"
}

http_status_post_empty() {
	cookie_file="$1"
	csrf="$2"
	url="$3"
	curl --silent --output /dev/null --write-out '%{http_code}' \
		--cookie "$cookie_file" \
		--cookie-jar "$cookie_file" \
		--header "$(csrf_header "$csrf")" \
		--request POST \
		"$url"
}

http_status_get() {
	cookie_file="$1"
	url="$2"
	curl --silent --output /dev/null --write-out '%{http_code}' \
		--cookie "$cookie_file" \
		--cookie-jar "$cookie_file" \
		"$url"
}

expect_status() {
	actual="$1"
	expected="$2"
	label="$3"
	[ "$actual" = "$expected" ] || fail "$label returned $actual, expected $expected"
}

bootstrap_super_admin() {
	curl --fail --silent \
		--cookie "$admin_cookie" \
		--cookie-jar "$admin_cookie" \
		--header "Content-Type: application/json" \
		--data '{"username":"admin","password":"CorrectHorseBatteryStaple!43"}' \
		"$base_url/api/bootstrap"
}

create_operator_user() {
	api_post_json "$admin_cookie" "$admin_csrf" "$base_url/api/admin/users" '{"username":"operator"}'
}

grant_operator_capabilities() {
	for capability in \
		identity.users.manage \
		jobs.read \
		jobs.approve \
		host.services.control \
		audit.events.read \
		observability.metrics.read \
		host.services.read; do
		api_post_json \
			"$admin_cookie" \
			"$admin_csrf" \
			"$base_url/api/admin/users/$operator_id/capabilities" \
			"{\"capability\":\"$capability\"}" >/dev/null
	done
}

verify_disabled_user_cannot_receive_magic_token() {
	api_post_empty "$admin_cookie" "$admin_csrf" "$base_url/api/admin/users/$operator_id/disable" >/dev/null
	status="$(http_status_post_empty "$admin_cookie" "$admin_csrf" "$base_url/api/admin/users/$operator_id/magic-token")"
	expect_status "$status" "409" "disabled user magic-token issue"
	api_post_empty "$admin_cookie" "$admin_csrf" "$base_url/api/admin/users/$operator_id/enable" >/dev/null
}

issue_operator_magic_token() {
	api_post_empty "$admin_cookie" "$admin_csrf" "$base_url/api/admin/users/$operator_id/magic-token"
}

login_with_magic_token() {
	token="$1"
	curl --fail --silent \
		--cookie "$operator_cookie" \
		--cookie-jar "$operator_cookie" \
		--header "Content-Type: application/json" \
		--data "{\"token\":\"$token\"}" \
		"$base_url/api/login/magic-token"
}

verify_operator_identity_admin_flow() {
	api_get "$operator_cookie" "$base_url/api/admin/users" >/dev/null
	api_post_json \
		"$operator_cookie" \
		"$operator_csrf" \
		"$base_url/api/admin/users/$operator_id/capabilities" \
		'{"capability":"host.metrics.read"}' >/dev/null
	api_delete \
		"$operator_cookie" \
		"$operator_csrf" \
		"$base_url/api/admin/users/$operator_id/capabilities/host.metrics.read" >/dev/null
}

create_service_job() {
	action="$1"
	api_post_empty "$operator_cookie" "$operator_csrf" "$base_url/api/services/nginx.service/$action"
}

wait_for_job_terminal_status() {
	job_id="$1"
	for _ in $(seq 1 40); do
		job_json="$(api_get "$operator_cookie" "$base_url/api/jobs/$job_id")"
		status="$(printf '%s' "$job_json" | json_field '.status')"
		case "$status" in
			completed|failed|not_implemented|rejected)
				printf '%s\n' "$job_json"
				return 0
				;;
		esac
		sleep 0.5
	done
	fail "job $job_id did not reach a terminal status"
}

verify_job_approval_flow() {
	job_json="$(create_service_job restart)"
	job_id="$(printf '%s' "$job_json" | json_field '.jobId')"
	[ -n "$job_id" ] && [ "$job_id" != "null" ] || fail "service-control job id missing"

	api_get "$operator_cookie" "$base_url/api/jobs" >/dev/null
	api_post_json \
		"$operator_cookie" \
		"$operator_csrf" \
		"$base_url/api/jobs/$job_id/approve" \
		'{"reason":"RC operator smoke approval"}' >/dev/null

	terminal_json="$(wait_for_job_terminal_status "$job_id")"
	terminal_status="$(printf '%s' "$terminal_json" | json_field '.status')"
	terminal_reason="$(printf '%s' "$terminal_json" | json_field '.status_reason')"
	[ "$terminal_status" = "failed" ] || fail "approved job ended as $terminal_status, expected failed"
	[ "$terminal_reason" = "mutation_disabled" ] || fail "approved job reason was $terminal_reason, expected mutation_disabled"
}

verify_job_rejection_flow() {
	job_json="$(create_service_job stop)"
	job_id="$(printf '%s' "$job_json" | json_field '.jobId')"
	[ -n "$job_id" ] && [ "$job_id" != "null" ] || fail "reject job id missing"

	rejected_json="$(api_post_json \
		"$operator_cookie" \
		"$operator_csrf" \
		"$base_url/api/jobs/$job_id/reject" \
		'{"reason":"RC operator smoke rejection"}')"
	status="$(printf '%s' "$rejected_json" | json_field '.status')"
	[ "$status" = "rejected" ] || fail "rejected job status was $status"
}

verify_activity_and_observability() {
	activity_json="$(api_get "$operator_cookie" "$base_url/api/activity")"
	activity_count="$(printf '%s' "$activity_json" | json_field '.events | length')"
	[ "$activity_count" -gt 0 ] || fail "activity returned no events"

	metrics_json="$(api_get "$operator_cookie" "$base_url/api/observability/metrics")"
	jobs_created="$(printf '%s' "$metrics_json" | json_field '.counters.jobs_created_total')"
	agent_failures="$(printf '%s' "$metrics_json" | json_field '.counters.agent_rpc_failures_total')"
	[ "$jobs_created" -ge 2 ] || fail "jobs_created_total was $jobs_created"
	[ "$agent_failures" -ge 1 ] || fail "agent_rpc_failures_total was $agent_failures"
}

verify_disabled_session_revoked() {
	api_post_empty "$admin_cookie" "$admin_csrf" "$base_url/api/admin/users/$operator_id/disable" >/dev/null
	status="$(http_status_get "$operator_cookie" "$base_url/api/jobs")"
	expect_status "$status" "401" "disabled user protected request"
}

start_runtime() {
	"$agent_bin" -socket "$socket_path" >"$tmpdir/agent.log" 2>&1 &
	agent_pid="$!"
	wait_for_file_socket "$socket_path"

	"$kojid_bin" \
		-config packaging/examples/koji.yaml \
		-database "$tmpdir/koji.db" \
		-static-dir "$static_dir" \
		-agent-socket "$socket_path" \
		-port "$port" >"$tmpdir/kojid.log" 2>&1 &
	kojid_pid="$!"
	wait_for_http "$base_url"
}

cleanup() {
	if [ -n "${kojid_pid:-}" ]; then
		kill "$kojid_pid" >/dev/null 2>&1 || true
		wait "$kojid_pid" >/dev/null 2>&1 || true
	fi
	if [ -n "${agent_pid:-}" ]; then
		kill "$agent_pid" >/dev/null 2>&1 || true
		wait "$agent_pid" >/dev/null 2>&1 || true
	fi
}

require_command curl
require_command jq
require_command tar

require_file "$artifact_dir/kojid-linux-amd64"
require_file "$artifact_dir/koji-agent-linux-amd64"
require_file "$artifact_dir/koji-rootfs-linux-amd64.tar.gz"

tmpdir="$(mktemp -d)"
trap cleanup EXIT INT TERM

tar -xzf "$artifact_dir/koji-rootfs-linux-amd64.tar.gz" -C "$tmpdir"
static_dir="$tmpdir/rootfs/usr/share/koji/dist"
[ -d "$static_dir" ] || fail "missing static assets in rootfs"

kojid_bin="$artifact_dir/kojid-linux-amd64"
agent_bin="$artifact_dir/koji-agent-linux-amd64"
chmod 0755 "$kojid_bin" "$agent_bin"

port="${KOJI_OPERATOR_SMOKE_PORT:-18080}"
base_url="http://127.0.0.1:$port"
socket_path="$tmpdir/agent.sock"
admin_cookie="$tmpdir/admin.cookies"
operator_cookie="$tmpdir/operator.cookies"

start_runtime

curl --fail --silent "$base_url/healthz" >/dev/null
curl --fail --silent "$base_url/readyz" | jq -e '.status == "ok"' >/dev/null

admin_session="$(bootstrap_super_admin)"
admin_csrf="$(printf '%s' "$admin_session" | json_field '.csrfToken')"
[ -n "$admin_csrf" ] && [ "$admin_csrf" != "null" ] || fail "admin csrf missing"

operator_json="$(create_operator_user)"
operator_id="$(printf '%s' "$operator_json" | json_field '.id')"
[ -n "$operator_id" ] && [ "$operator_id" != "null" ] || fail "operator id missing"

grant_operator_capabilities
verify_disabled_user_cannot_receive_magic_token

magic_token_json="$(issue_operator_magic_token)"
magic_token="$(printf '%s' "$magic_token_json" | json_field '.token')"
[ -n "$magic_token" ] && [ "$magic_token" != "null" ] || fail "magic token missing"

operator_session="$(login_with_magic_token "$magic_token")"
operator_csrf="$(printf '%s' "$operator_session" | json_field '.csrfToken')"
[ -n "$operator_csrf" ] && [ "$operator_csrf" != "null" ] || fail "operator csrf missing"

verify_operator_identity_admin_flow
verify_job_approval_flow
verify_job_rejection_flow
verify_activity_and_observability
verify_disabled_session_revoked

printf '%s\n' "operator smoke passed"
