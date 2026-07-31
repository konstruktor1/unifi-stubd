# Produkt-Roadmap

Stand: 2026-07-31.

Die Roadmap besitzt nur die Produktphasen. Aktuelle Prioritaet steht im
[Backlog](backlog.md), abgenommene Laeufe gehoeren in die
[Kompatibilitaet](compatibility.md).

## Phase 1: US8 mit echten Controllermetriken

Eine Nicht-UniFi-Linux-Bridge ausliefern, die ein Betreiber in UniFi Network
als `US8` hinzufuegen kann.

Die Phase endet, wenn dieselben Produktbytes auf aktuellem Stable und Release
Candidate gefunden und adoptiert werden, verbunden bleiben und echten
Link-State, Speed, Counter und Raten der benannten lokalen Quelle zeigen, ohne
Host-Provisioning zu erlauben.

## Phase 2: Nutzbarer Linux-Release

Das abgenommene Verhalten als installierbares Paket und Betreiberpfad auf einem
unterstuetzten Linux-Host liefern.

Die Phase endet, wenn Installation, unterstuetzte Bridge-Konfiguration,
paketierter Servicestart, Adoption, sichtbare Metriken, Grenzen und Recovery als
eine dokumentierte Reise funktionieren.

## Phase 3: Weitere Repraesentationen bei Bedarf

Ein weiteres Profil oder eine Plattform kommt nur fuer ein konkretes
Betreiberergebnis hinzu. Jede neue Repraesentation braucht eine benannte
read-only Quelle, eine explizite Sicherheitsentscheidung und eigene dynamische
Stable-/RC-Abnahme; Evidence eines benachbarten Modells erweitert den
Produktscope nicht.
