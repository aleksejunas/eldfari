# Raketti

Gravitasjons-cave-flyer i vektorstil, inspirert av Turboraketti (Amiga, Heikki Kosola).
Ingen grafikk eller baner er kopiert fra originalen. Bare ideen.

## Kjør

```sh
go mod tidy      # henter Ebitengine v2 (go.mod har ingen require ennå)
go run .
go test ./sim    # fysikk, kollisjon og landing
```

Taster: `A`/`D` eller piler = roter, `W`/pil opp/mellomrom = thrust, `R` = restart.

Linux trenger de vanlige Ebitengine-bibliotekene (X11/GL/ALSA). Se
<https://ebitengine.org/en/documents/install.html> hvis bygget klager.

## Status

**Verifisert:** `sim/` kompilerer og alle testene går grønt.
**Ikke verifisert:** `main.go` og `fx.go` ble skrevet uten å kunne hente Ebitengine
(Go-proxyen var blokkert i sandboxen), så de er aldri kompilert. Forvent å fikse
småting i API-bruk første gang du bygger.

## Oppbygging

| Fil | Innhold |
|---|---|
| `sim/level.go` | Terreng som polygoner, rasteriseres til usynlig bitmask. Plattformer. `DemoLevel()`. |
| `sim/world.go` | Skip, fysikk (gravitasjon, thrust/masse, drag), kollisjon, landing. Alle konstanter øverst. |
| `sim/sim_test.go` | Tester. |
| `main.go` | Ebitengine: input, kamera, tegning, HUD. |
| `fx.go` | Eksos og vrakdeler (linjestykker). |

Tanken: `sim` vet ingenting om Ebitengine, så den er deterministisk og testbar uten vindu.
Fast tidssteg (60 Hz), ett `Step` per `Update`.

## Fysikken i korte trekk

- Akselerasjon fra motor = `Thrust / masse`, og masse = `BaseMass + FuelMass * fuel`.
  Fullt drivstoff gir tregere skip. Det er kjernen i Turboraketti-følelsen.
- Landing krever: alle treffpunkter på samme plattform, `|VY| <= 45`, `|VX| <= 30`,
  vinkel innen 0,3 rad fra rett opp. Ellers krasj.
- Plattformer fyller på drivstoff når du står på dem.
- Alle tallene er gjetninger. Spill og juster i `sim/world.go`.

## Kjente begrensninger

- Ingen substeps: over ca. 600 px/s kan skipet teoretisk gå gjennom tynne vegger.
  Terrenget i testbanen er tykt nok, men det bør fikses før høyere hastigheter.
- Kollisjon er punkttester (hjørner + kantmidtpunkt), ikke full linje-mot-maske.
- Skipet krasjer i verdensgrensen (alt utenfor banen er fast).

## Neste steg

1. ~~Ett skip, gravitasjon, thrust, bane, krasj~~ (dette)
2. To spillere, split-screen (to kameraer mot hver sin offscreen-`Image`)
3. Basestasjon: velg våpen/drivstoff, reparasjon, masse påvirker akselerasjon
4. Våpen
5. Væske (drag per piksel), flere baner, ranking
