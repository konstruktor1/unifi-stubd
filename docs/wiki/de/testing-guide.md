# Testleitfaden

Alle Go-abhaengigen Pruefungen laufen ueber den konfigurierten Build Runner.
Host-Go und lokal installierte Go-Werkzeuge werden nicht unterstuetzt.

## Kandidatenpruefungen

`make check`, `make race` und das benoetigte Paketziel ueber den Build Runner
ausfuehren. Die reduzierte Go-Suite schuetzt unabhaengige Wire-Vektoren,
begrenztes Adoption-/Host-Safety-Verhalten und die lokale Metrikabbildung; sie
bildet keine Controllerreise nach.

## Produktabnahme

Bei Aenderungen an Adoption, Metriken oder Controllerkompatibilitaet:

```sh
BUILD_RUNNER_DIR=/pfad/zum/build-runner make integration-nonunifi-metrics
```

Der Gate loest zu Laufbeginn das neueste veroeffentlichte Stable und den
neuesten veroeffentlichten Release Candidate auf und nutzt in beiden Spuren
dasselbe Produktimage. Er belegt Modell `US8`, Adoption, stabile Verbindung,
beobachtete Linkmetadaten der Quelle, steigende echte Zaehler, positive Raten
und vollstaendiges Cleanup. Eine feste Controllerversion begruendet keinen
Releaseclaim.

Der manuelle Workflow `Controller Integration` ruft fuer einen
vertrauenswuerdigen Ref dasselbe Ziel auf und laedt weder Captures noch
Controller-State hoch.

## Externe Kompatibilitaet

Firmware-Research, QEMU-/VM-Runtimes, physische Topologien, reale Hosts und
deployed Controller gehoeren zu `unifi-stubd-labs`. Deren Evidence erweitert
den unterstuetzten Core-Produktscope nicht ohne eigene Produktentscheidung.
