# Webportal Data

Data files here are tiny deterministic payloads used by Docker webportal
wrappers. They keep wrapper behavior stable across test runs without embedding
private console state.

`ubnt-tools-id.txt` is the current example: it gives the `ubnt-tools` wrapper a
stable identity response for Core/support-bundle paths that expect the command
to exist.

Do not place real console identifiers, generated support bundles, tokens,
browser session data, or local user data here.
