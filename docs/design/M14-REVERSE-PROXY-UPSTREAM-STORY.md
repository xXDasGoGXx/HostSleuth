# M14 — Reverse Proxy / Upstream Story + Admin Console v2

Status: implemented, CI-accepted, and merged; release/production remain separate gates.

## Admin question

> DNS resolves and the public port answers. Where does the reverse-proxy request path actually break?

M14 connects evidence HostSleuth already knows how to collect into one explicit public-endpoint-to-upstream story.

## Bounded v1 input

One request contains:

- required public HTTP/HTTPS URL;
- required expected upstream HTTP/HTTPS URL.

Examples:

```text
public:   https://photos.example.com/
upstream: http://192.0.2.40:8080/
```

```text
public:   https://app.example.com/admin/
upstream: https://10.0.0.25:8443/admin/
```

URLs containing embedded credentials are refused.

## Deterministic path

M14 evaluates, in order:

1. public DNS;
2. public route/TCP;
3. public TLS when the public URL uses HTTPS;
4. public HTTP response;
5. local listener/Docker publication context when the public endpoint resolves to this host;
6. upstream DNS;
7. upstream route/TCP;
8. upstream TLS when the upstream URL uses HTTPS;
9. direct upstream HTTP using the upstream URL's native Host/SNI;
10. when public and upstream hostnames differ, the same upstream address is probed again using the public hostname as HTTP Host and TLS SNI.

The forwarded-Host/SNI probe is derived only from the explicit public URL. M14 does not accept arbitrary custom request headers.

## HTTP probe boundary

HTTP evidence is metadata only.

M14 issues bounded HEAD requests with:

- no request body;
- no credentials;
- no cookies;
- no Authorization header;
- no environment proxy;
- bounded timeouts;
- selected response metadata only: status, Location, Server, Content-Type;
- no response body.

For the public URL, same-host redirects may be followed up to a small fixed limit. A redirect to a different hostname is recorded and not followed.

For upstream probes, redirects are recorded but not followed. This avoids accidentally turning an explicitly supplied upstream into a redirect-controlled server-side probe.

## Stage status

Each stage is one of:

- `pass` — evidence supports the expected request path;
- `warn` — the stage answered but returned behavior that often explains an application/proxy mismatch, such as HTTP 4xx;
- `fail` — direct evidence shows a broken stage, such as DNS/TCP/TLS failure or HTTP 5xx;
- `unknown` — HostSleuth cannot prove the stage either way with the available evidence;
- `info` — contextual evidence that is not itself a pass/fail gate.

Overall outcome:

1. first proven `fail`;
2. otherwise first `warn`;
3. otherwise first `unknown`;
4. otherwise `pass`.

A later proven failure is not hidden by an earlier unknown.

## HTTP status semantics

For request-path diagnosis:

- 1xx, 2xx, and 3xx = `pass`;
- 4xx = `warn` because the path is reachable but the application/vhost/path rejected the request;
- 5xx = `fail` because the responding server reported a server-side/proxy/upstream failure.

M14 does not claim that a 4xx is always wrong; authentication and policy responses may be expected.

## Native vs forwarded Host/SNI comparison

When public host and upstream host differ, M14 performs two upstream HEAD probes against the same explicit upstream address:

1. native upstream Host/SNI;
2. public Host/SNI.

This can expose cases such as:

- upstream works natively but fails when the public Host is forwarded;
- upstream only serves the intended virtual host when the public Host/SNI is used;
- upstream TLS certificate matches the public name but not the private/IP URL;
- both variants fail, so the problem is below virtual-host routing.

Host/SNI differences are evidence, not proof of a particular proxy configuration. M14 never claims nginx/NPM/Caddy/Traefik is configured a certain way unless direct configuration evidence exists.

## Local proxy evidence

If the public endpoint resolves to a local interface, M14 surfaces existing deterministic local evidence:

- listener ownership;
- Docker port publication/internal-only evidence.

If the public endpoint is remote from the HostSleuth machine, local proxy ownership is informationally unavailable and is shown as context rather than a failure.

M14 v1 does not parse Nginx Proxy Manager, nginx, Caddy, Traefik, Apache, HAProxy, or other proxy configuration files.

## Interfaces

CLI:

```text
hostsleuth proxy --upstream http://192.0.2.40:8080/ https://photos.example.com/
```

HTTP API:

```text
GET /api/proxy-story?public=https%3A%2F%2Fphotos.example.com%2F&upstream=http%3A%2F%2F192.0.2.40%3A8080%2F
```

Web UI:

- dedicated **Proxy Path** view;
- visual left-to-right request path;
- first proven problem highlighted;
- expandable evidence per stage;
- public vs upstream Host/SNI comparison;
- one-click copy of a bounded text evidence summary;
- one-click handoff to Diagnose / DNS Detective / Expectations where useful.

## Admin Console v2

M14 extends the console shell with:

- a Proxy Path action in Quick Target;
- status rail/visual request path;
- consistent evidence-copy behavior;
- stable hash/deep-link to the Proxy Path view;
- clearer "front door" vs "upstream" grouping;
- responsive stacked flow on narrow screens.

No decorative monitoring graphs are added.

## Security boundary

M14 is read-only and adds no privilege.

It does not:

- read arbitrary proxy configuration;
- submit request bodies;
- accept cookies, credentials, Authorization headers, or arbitrary custom headers;
- return response bodies;
- follow cross-host redirects;
- perform automatic discovery/scanning;
- poll endpoints;
- change DNS/proxy/container/service configuration;
- restart/reload anything;
- remediate automatically.

The public and upstream URLs are explicit operator inputs. As with Diagnose and DNS Detective, a deliberately LAN-bound HostSleuth UI must be treated as trusted-network access; loopback remains the safer default.

## Explicit non-goals for v1

Not included:

- proxy config parser;
- automatic upstream discovery;
- HTTP body inspection;
- WebSocket transaction testing;
- authentication flows;
- POST/PUT/DELETE requests;
- arbitrary request headers;
- automatic proxy reload/fix;
- multi-host agents;
- monitoring/history of URL health.

## Acceptance

Required before merge:

- gofmt;
- git diff --check;
- go vet ./...;
- go test ./...;
- native build;
- JavaScript syntax checks;
- deterministic unit tests for stage ordering and first-problem precedence;
- HTTP fixture tests for 2xx/4xx/5xx, same-host redirect, cross-host redirect stop, and public-Host forwarding;
- disposable CLI/API acceptance;
- headless-browser DOM validation of Proxy Path view and Admin Console v2;
- existing Docker/native Safe Actions smoke unchanged;
- existing M13 DNS Detective smoke unchanged.
