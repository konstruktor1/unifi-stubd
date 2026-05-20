# UXG-Pro Tools

This directory is for small, project-owned helper programs used while analyzing
UXG-Pro firmware and inform traffic. Tools here should be reproducible from
the repository, operate on explicit input files, and avoid hidden dependency on
local controller state.

The first tool, `decode-inform/`, decodes one captured UniFi inform packet body
with the repository's inform decoder. It exists so research notes can be
derived from sanitized lab captures without embedding one-off decode snippets
in documentation.

Keep this tree source-only. Raw captures, controller tokens, private URLs, and
real inform auth keys stay in ignored local paths.
